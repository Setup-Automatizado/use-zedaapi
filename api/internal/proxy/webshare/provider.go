package webshare

import (
	"context"
	"fmt"
	"log/slog"

	proxy "go.mau.fi/whatsmeow/api/internal/proxy"
)

// Provider implements proxy.ProxyProvider for Webshare.
type Provider struct {
	client *Client
	log    *slog.Logger
}

// NewProvider creates a new Webshare proxy provider.
// planID selects the Webshare subscription plan (e.g. ISP/static vs rotating residential).
// mode is "direct" for static/ISP proxies or "backbone" for rotating residential.
func NewProvider(apiKey, endpoint, planID, mode string, log *slog.Logger) *Provider {
	return &Provider{
		client: NewClient(apiKey, endpoint, planID, mode, log),
		log:    log.With(slog.String("component", "webshare_provider")),
	}
}

func (p *Provider) Type() proxy.ProviderType {
	return proxy.ProviderTypeWebshare
}

func (p *Provider) ListProxies(ctx context.Context, filter proxy.ProxyFilter) ([]proxy.ProxyEntry, int, error) {
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 100
	}

	resp, err := p.client.ListProxies(ctx, filter.CountryCodes, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	entries := make([]proxy.ProxyEntry, 0, len(resp.Results))
	for _, item := range resp.Results {
		entry := itemToEntry(item)
		if item.Port <= 0 || item.Port > 65535 {
			continue // skip proxies with invalid TCP ports
		}
		entries = append(entries, entry)
	}
	return entries, resp.Count, nil
}

func (p *Provider) ValidateProxy(ctx context.Context, externalID string) (bool, error) {
	// Webshare doesn't have a single-proxy validation endpoint.
	// We rely on the "valid" field from ListProxies.
	return true, nil
}

func (p *Provider) ReplaceProxy(ctx context.Context, externalID string) (*proxy.ReplaceResult, error) {
	// Backbone mode: externalID is the Webshare proxy ID (e.g., "b-BR-1")
	// The replace API needs the backbone host and port, but we use the ID-based endpoint
	resp, err := p.client.ReplaceProxyByID(ctx, externalID)
	if err != nil {
		return &proxy.ReplaceResult{
			OldExternalID: externalID,
			Success:       false,
			Message:       err.Error(),
		}, nil
	}

	address := resp.ProxyAddress
	if address == "" {
		address = backboneHost
	}

	newEntry := proxy.ProxyEntry{
		ExternalID:  resp.ID,
		ProxyURL:    fmt.Sprintf("socks5://%s:%s@%s:%d", resp.Username, resp.Password, address, resp.Port),
		CountryCode: resp.CountryCode,
		City:        resp.CityName,
		Username:    resp.Username,
		Password:    resp.Password,
		Address:     address,
		Port:        resp.Port,
		Valid:       resp.Valid,
	}

	return &proxy.ReplaceResult{
		OldExternalID: externalID,
		NewProxy:      &newEntry,
		Success:       true,
		Message:       "proxy replaced successfully",
	}, nil
}

func (p *Provider) GetInfo(ctx context.Context) (*proxy.ProviderInfo, error) {
	resp, err := p.client.ListProxies(ctx, nil, 1, 1)
	if err != nil {
		return nil, err
	}
	return &proxy.ProviderInfo{
		Type:            proxy.ProviderTypeWebshare,
		Name:            "Webshare",
		TotalProxies:    resp.Count,
		SupportsReplace: true,
		RateLimitRPM:    240,
	}, nil
}

func (p *Provider) SyncAll(ctx context.Context) ([]proxy.ProxyEntry, error) {
	var allEntries []proxy.ProxyEntry
	page := 1
	pageSize := 100

	for {
		select {
		case <-ctx.Done():
			if len(allEntries) > 0 {
				return allEntries, fmt.Errorf("context cancelled at page %d (%d entries fetched): %w", page, len(allEntries), ctx.Err())
			}
			return nil, ctx.Err()
		default:
		}

		resp, err := p.client.ListProxies(ctx, nil, page, pageSize)
		if err != nil {
			// Return partial results instead of discarding everything.
			if len(allEntries) > 0 {
				p.log.Warn("sync interrupted, returning partial results",
					slog.Int("page", page),
					slog.Int("fetched_entries", len(allEntries)),
					slog.String("error", err.Error()))
				return allEntries, fmt.Errorf("sync interrupted at page %d (%d entries fetched): %w", page, len(allEntries), err)
			}
			return nil, fmt.Errorf("sync page %d: %w", page, err)
		}

		for _, item := range resp.Results {
			if item.Port <= 0 || item.Port > 65535 {
				continue // skip proxies with invalid TCP ports
			}
			allEntries = append(allEntries, itemToEntry(item))
		}

		p.log.Debug("synced proxy page",
			slog.Int("page", page),
			slog.Int("fetched", len(resp.Results)),
			slog.Int("total", resp.Count))

		if resp.Next == nil || len(resp.Results) == 0 {
			break
		}
		page++
	}

	return allEntries, nil
}

func (p *Provider) Close() error {
	return p.client.Close()
}

const backboneHost = "p.webshare.io"

// itemToEntry converts a Webshare ProxyItem to a generic ProxyEntry.
// Returns valid=false for entries with invalid ports (> 65535).
func itemToEntry(item ProxyItem) proxy.ProxyEntry {
	address := item.ProxyAddress
	if address == "" {
		address = backboneHost
	}

	valid := item.Valid
	if item.Port <= 0 || item.Port > 65535 {
		valid = false
	}

	return proxy.ProxyEntry{
		ExternalID:  item.ID,
		ProxyURL:    fmt.Sprintf("socks5://%s:%s@%s:%d", item.Username, item.Password, address, item.Port),
		CountryCode: item.CountryCode,
		City:        item.CityName,
		Username:    item.Username,
		Password:    item.Password,
		Address:     address,
		Port:        item.Port,
		Valid:       valid,
	}
}
