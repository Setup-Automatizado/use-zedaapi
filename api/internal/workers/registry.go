package workers

import (
	"context"
	"hash/fnv"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	HeartbeatInterval time.Duration
	Expiry            time.Duration
	AdvertiseAddr     string
}

type Info struct {
	ID            string
	Hostname      string
	AppEnv        string
	LastSeen      time.Time
	AdvertiseAddr string
}

type Registry struct {
	pool          *pgxpool.Pool
	workerID      string
	hostname      string
	appEnv        string
	advertiseAddr string
	cfg           Config
	log           *slog.Logger

	stopOnce sync.Once
	stopCh   chan struct{}
	doneCh   chan struct{}

	cache atomic.Value // []Info
}

func NewRegistry(pool *pgxpool.Pool, workerID, hostname, appEnv string, cfg Config, log *slog.Logger) *Registry {
	if cfg.HeartbeatInterval <= 0 {
		cfg.HeartbeatInterval = 5 * time.Second
	}
	if cfg.Expiry <= cfg.HeartbeatInterval {
		cfg.Expiry = cfg.HeartbeatInterval * 2
	}

	r := &Registry{
		pool:          pool,
		workerID:      workerID,
		hostname:      hostname,
		appEnv:        appEnv,
		advertiseAddr: cfg.AdvertiseAddr,
		cfg:           cfg,
		log:           log,
		stopCh:        make(chan struct{}),
		doneCh:        make(chan struct{}),
	}
	r.cache.Store([]Info{})
	return r
}

func (r *Registry) Start(ctx context.Context) error {
	go r.run(ctx)
	return nil
}

func (r *Registry) Stop(ctx context.Context) {
	r.stopOnce.Do(func() {
		close(r.stopCh)
	})

	select {
	case <-r.doneCh:
	case <-ctx.Done():
	}
}

func (r *Registry) WorkerID() string {
	return r.workerID
}

func (r *Registry) ActiveWorkers() []Info {
	raw := r.cache.Load().([]Info)
	out := make([]Info, len(raw))
	copy(out, raw)
	if len(out) == 0 {
		out = append(out, Info{ID: r.workerID, Hostname: r.hostname, AppEnv: r.appEnv, LastSeen: time.Now(), AdvertiseAddr: r.advertiseAddr})
	}
	return out
}

func (r *Registry) ResolveAddr(workerID string) (string, bool) {
	workers := r.ActiveWorkers()
	for _, w := range workers {
		if w.ID == workerID && w.AdvertiseAddr != "" {
			return w.AdvertiseAddr, true
		}
	}
	return "", false
}

func (r *Registry) AssignedOwner(instanceID uuid.UUID) string {
	workers := r.ActiveWorkers()
	if len(workers) == 0 {
		return r.workerID
	}

	payload := instanceID.String()
	var bestScore uint64
	var owner string

	for _, info := range workers {
		h := fnv.New64a()
		_, _ = h.Write([]byte(payload))
		_, _ = h.Write([]byte(info.ID))
		score := h.Sum64()
		if owner == "" || score > bestScore || (score == bestScore && info.ID > owner) {
			bestScore = score
			owner = info.ID
		}
	}

	if owner == "" {
		return r.workerID
	}
	return owner
}

func (r *Registry) ForceRefresh(ctx context.Context) {
	r.refreshWorkers(ctx)
}

func (r *Registry) run(ctx context.Context) {
	ticker := time.NewTicker(r.cfg.HeartbeatInterval)
	defer ticker.Stop()
	defer close(r.doneCh)

	r.beat(ctx)

	for {
		select {
		case <-ctx.Done():
			r.deregister(context.Background())
			return
		case <-r.stopCh:
			r.deregister(context.Background())
			return
		case <-ticker.C:
			r.beat(ctx)
		}
	}
}

func (r *Registry) beat(ctx context.Context) {
	hbCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := r.upsertWorker(hbCtx); err != nil {
		if r.log != nil {
			r.log.Warn("worker heartbeat failed",
				slog.String("worker_id", r.workerID),
				slog.String("error", err.Error()))
		}
	}

	r.refreshWorkers(ctx)
}

func (r *Registry) upsertWorker(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `
        INSERT INTO worker_sessions (worker_id, hostname, app_env, last_seen, metadata)
        VALUES ($1, $2, $3, NOW(), jsonb_build_object('advertise_addr', $4))
        ON CONFLICT (worker_id) DO UPDATE
        SET hostname = EXCLUDED.hostname,
            app_env = EXCLUDED.app_env,
            last_seen = EXCLUDED.last_seen,
            metadata = jsonb_set(worker_sessions.metadata, '{advertise_addr}', to_jsonb($4::text))
    `, r.workerID, r.hostname, r.appEnv, r.advertiseAddr)
	return err
}

func (r *Registry) refreshWorkers(ctx context.Context) {
	refreshCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	threshold := time.Now().Add(-r.cfg.Expiry)
	rows, err := r.pool.Query(refreshCtx, `
        SELECT worker_id, hostname, last_seen, COALESCE(metadata->>'advertise_addr', '') as advertise_addr
        FROM worker_sessions
        WHERE app_env = $1 AND last_seen >= $2
        ORDER BY last_seen DESC
    `, r.appEnv, threshold)
	if err != nil {
		if r.log != nil {
			r.log.Warn("list workers failed",
				slog.String("worker_id", r.workerID),
				slog.String("error", err.Error()))
		}
		return
	}
	defer rows.Close()

	var workers []Info
	for rows.Next() {
		var info Info
		if err := rows.Scan(&info.ID, &info.Hostname, &info.LastSeen, &info.AdvertiseAddr); err != nil {
			if r.log != nil {
				r.log.Warn("scan worker failed",
					slog.String("worker_id", r.workerID),
					slog.String("error", err.Error()))
			}
			return
		}
		info.AppEnv = r.appEnv
		workers = append(workers, info)
	}

	if len(workers) == 0 {
		workers = append(workers, Info{ID: r.workerID, Hostname: r.hostname, AppEnv: r.appEnv, LastSeen: time.Now(), AdvertiseAddr: r.advertiseAddr})
	}

	r.cache.Store(workers)
}

func (r *Registry) deregister(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if _, err := r.pool.Exec(ctx, `DELETE FROM worker_sessions WHERE worker_id = $1`, r.workerID); err != nil {
		if r.log != nil {
			r.log.Warn("failed to deregister worker",
				slog.String("worker_id", r.workerID),
				slog.String("error", err.Error()))
		}
	}
}
