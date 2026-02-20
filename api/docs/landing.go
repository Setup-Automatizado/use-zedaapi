package docs

import (
	"fmt"
	"net/http"
	"sync"

	"go.mau.fi/whatsmeow/api/internal/version"
)

var (
	landingOnce sync.Once
	landingHTML string
	landingErr  error
)

func generateLandingHTML(baseURL string) (string, error) {
	ver := version.String()

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Zé da API</title>
  <link rel="icon" type="image/x-icon" href="/favicon.ico">
  <link rel="icon" type="image/png" sizes="32x32" href="/favicon-32x32.png">
  <link rel="icon" type="image/png" sizes="16x16" href="/favicon-16x16.png">
  <link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png">
  <link rel="manifest" href="/manifest.json">
  <meta name="theme-color" content="#0a0a14">
  <meta name="description" content="WhatsApp Business API powered by Whatsmeow">
  <style>
    @import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&display=swap');

    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

    :root {
      --bg-primary: #0a0a14;
      --bg-secondary: #0f0f1e;
      --bg-card: rgba(255, 255, 255, 0.03);
      --bg-card-hover: rgba(255, 255, 255, 0.06);
      --border: rgba(255, 255, 255, 0.06);
      --border-hover: rgba(255, 255, 255, 0.12);
      --text-primary: #f1f5f9;
      --text-secondary: #94a3b8;
      --text-muted: #64748b;
      --accent: #818cf8;
      --accent-glow: rgba(129, 140, 248, 0.15);
      --green: #22c55e;
      --green-glow: rgba(34, 197, 94, 0.15);
      --red: #ef4444;
    }

    body {
      font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      background: var(--bg-primary);
      color: var(--text-primary);
      min-height: 100vh;
      overflow-x: hidden;
      -webkit-font-smoothing: antialiased;
    }

    /* ── Background ────────────────────────────────── */
    .bg-grid {
      position: fixed;
      inset: 0;
      background-image:
        linear-gradient(rgba(255,255,255,0.02) 1px, transparent 1px),
        linear-gradient(90deg, rgba(255,255,255,0.02) 1px, transparent 1px);
      background-size: 64px 64px;
      pointer-events: none;
    }

    .bg-glow {
      position: fixed;
      top: -200px;
      left: 50%%;
      transform: translateX(-50%%);
      width: 800px;
      height: 600px;
      background: radial-gradient(ellipse, var(--accent-glow), transparent 70%%);
      pointer-events: none;
      opacity: 0.6;
    }

    /* ── Layout ────────────────────────────────────── */
    .container {
      position: relative;
      max-width: 960px;
      margin: 0 auto;
      padding: 0 24px;
      display: flex;
      flex-direction: column;
      align-items: center;
      min-height: 100vh;
    }

    /* ── Header ────────────────────────────────────── */
    .header {
      width: 100%%;
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 20px 0;
      border-bottom: 1px solid var(--border);
    }

    .header-left {
      display: flex;
      align-items: center;
      gap: 12px;
    }

    .header-logo {
      width: 32px;
      height: 32px;
      border-radius: 8px;
      object-fit: cover;
    }

    .header-title {
      font-size: 14px;
      font-weight: 600;
      color: var(--text-primary);
      letter-spacing: -0.01em;
    }

    .header-right {
      display: flex;
      align-items: center;
      gap: 12px;
    }

    .header-badge {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      padding: 4px 10px;
      background: var(--bg-card);
      border: 1px solid var(--border);
      border-radius: 6px;
      font-size: 11px;
      font-weight: 500;
      color: var(--text-muted);
    }

    .status-dot {
      width: 6px;
      height: 6px;
      border-radius: 50%%;
      background: var(--text-muted);
      transition: background 0.3s ease, box-shadow 0.3s ease;
    }

    .status-dot.online {
      background: var(--green);
      box-shadow: 0 0 8px var(--green-glow);
    }

    .status-dot.offline {
      background: var(--red);
    }

    /* ── Hero ──────────────────────────────────────── */
    .hero {
      display: flex;
      flex-direction: column;
      align-items: center;
      text-align: center;
      padding: 80px 0 48px;
      gap: 24px;
    }

    .hero-avatar {
      width: 140px;
      height: 140px;
      border-radius: 28px;
      object-fit: cover;
      border: 2px solid var(--border);
      box-shadow: 0 16px 48px rgba(0, 0, 0, 0.4), 0 0 80px var(--accent-glow);
      transition: transform 0.4s cubic-bezier(0.22, 1, 0.36, 1), box-shadow 0.4s ease;
    }

    .hero-avatar:hover {
      transform: scale(1.05) translateY(-4px);
      box-shadow: 0 24px 64px rgba(0, 0, 0, 0.5), 0 0 120px var(--accent-glow);
    }

    .hero h1 {
      font-size: 48px;
      font-weight: 800;
      letter-spacing: -0.03em;
      line-height: 1.1;
      background: linear-gradient(135deg, #fff 0%%, #818cf8 50%%, #a78bfa 100%%);
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
      background-clip: text;
    }

    .hero-sub {
      font-size: 16px;
      color: var(--text-secondary);
      line-height: 1.6;
      max-width: 480px;
    }

    .hero-version {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      padding: 4px 12px;
      background: var(--accent-glow);
      border: 1px solid rgba(129, 140, 248, 0.2);
      border-radius: 20px;
      font-size: 12px;
      font-weight: 500;
      color: #a5b4fc;
    }

    /* ── Cards ─────────────────────────────────────── */
    .cards {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 16px;
      width: 100%%;
      max-width: 640px;
    }

    .card {
      display: flex;
      flex-direction: column;
      gap: 12px;
      padding: 24px;
      background: var(--bg-card);
      border: 1px solid var(--border);
      border-radius: 16px;
      text-decoration: none;
      color: var(--text-primary);
      transition: all 0.2s ease;
      cursor: pointer;
    }

    .card:hover {
      background: var(--bg-card-hover);
      border-color: var(--border-hover);
      transform: translateY(-2px);
      box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
    }

    .card-icon {
      width: 40px;
      height: 40px;
      border-radius: 10px;
      display: flex;
      align-items: center;
      justify-content: center;
    }

    .card-icon svg {
      width: 20px;
      height: 20px;
    }

    .card-icon.scalar {
      background: rgba(129, 140, 248, 0.1);
      color: var(--accent);
    }

    .card-icon.swagger {
      background: rgba(34, 197, 94, 0.1);
      color: var(--green);
    }

    .card-title {
      font-size: 15px;
      font-weight: 600;
      letter-spacing: -0.01em;
    }

    .card-desc {
      font-size: 13px;
      color: var(--text-secondary);
      line-height: 1.5;
    }

    .card-arrow {
      display: flex;
      align-items: center;
      gap: 4px;
      font-size: 12px;
      font-weight: 500;
      color: var(--text-muted);
      margin-top: auto;
      transition: color 0.15s ease, gap 0.15s ease;
    }

    .card:hover .card-arrow {
      color: var(--accent);
      gap: 8px;
    }

    .card-arrow svg {
      width: 14px;
      height: 14px;
    }

    /* ── Spec Links ────────────────────────────────── */
    .spec-links {
      display: flex;
      align-items: center;
      gap: 12px;
      margin-top: 32px;
    }

    .spec-link {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      padding: 6px 14px;
      background: var(--bg-card);
      border: 1px solid var(--border);
      border-radius: 8px;
      font-size: 12px;
      font-weight: 500;
      color: var(--text-muted);
      text-decoration: none;
      transition: all 0.15s ease;
    }

    .spec-link:hover {
      color: var(--text-secondary);
      border-color: var(--border-hover);
      background: var(--bg-card-hover);
    }

    .spec-link svg {
      width: 14px;
      height: 14px;
    }

    /* ── Footer ────────────────────────────────────── */
    .footer {
      margin-top: auto;
      padding: 24px 0;
      border-top: 1px solid var(--border);
      width: 100%%;
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 8px;
      font-size: 12px;
      color: var(--text-muted);
    }

    .footer-tech {
      display: inline-flex;
      align-items: center;
      gap: 4px;
      padding: 2px 8px;
      background: var(--bg-card);
      border-radius: 4px;
      font-size: 11px;
      font-weight: 500;
    }

    /* ── Responsive ────────────────────────────────── */
    @media (max-width: 640px) {
      .hero { padding: 48px 0 32px; }
      .hero h1 { font-size: 32px; }
      .hero-avatar { width: 100px; height: 100px; border-radius: 20px; }
      .cards { grid-template-columns: 1fr; }
      .header-badge { display: none; }
    }
  </style>
</head>
<body>
  <div class="bg-grid"></div>
  <div class="bg-glow"></div>

  <div class="container">
    <header class="header">
      <div class="header-left">
        <img src="/capa.png" alt="Zé da API" class="header-logo" />
        <span class="header-title">Zé da API</span>
      </div>
      <div class="header-right">
        <div class="header-badge">
          <span class="status-dot" id="status-dot"></span>
          <span id="status-text">Checking...</span>
        </div>
        <div class="header-badge">v%s</div>
      </div>
    </header>

    <section class="hero">
      <img src="/capa.png" alt="Zé da API" class="hero-avatar" />
      <div class="hero-version">WhatsApp Business API</div>
      <h1>Zé da API</h1>
      <p class="hero-sub">Production-grade WhatsApp API built on Whatsmeow. Send messages, manage instances, configure webhooks, and handle events at scale.</p>
    </section>

    <div class="cards">
      <a href="/docs/" class="card">
        <div class="card-icon scalar">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z"/>
            <polyline points="14 2 14 8 20 8"/>
            <line x1="16" y1="13" x2="8" y2="13"/>
            <line x1="16" y1="17" x2="8" y2="17"/>
          </svg>
        </div>
        <div class="card-title">API Reference</div>
        <div class="card-desc">Modern interactive docs with code snippets, search, and environment switching.</div>
        <div class="card-arrow">
          Scalar
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="5" y1="12" x2="19" y2="12"/>
            <polyline points="12 5 19 12 12 19"/>
          </svg>
        </div>
      </a>
      <a href="/docs/swagger/" class="card">
        <div class="card-icon swagger">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="16 18 22 12 16 6"/>
            <polyline points="8 6 2 12 8 18"/>
          </svg>
        </div>
        <div class="card-title">Swagger UI</div>
        <div class="card-desc">Classic try-it-out explorer with editable examples and request testing.</div>
        <div class="card-arrow">
          Try it out
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="5" y1="12" x2="19" y2="12"/>
            <polyline points="12 5 19 12 12 19"/>
          </svg>
        </div>
      </a>
    </div>

    <div class="spec-links">
      <a href="/docs/openapi.yaml" class="spec-link" target="_blank">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
          <polyline points="7 10 12 15 17 10"/>
          <line x1="12" y1="15" x2="12" y2="3"/>
        </svg>
        OpenAPI YAML
      </a>
      <a href="/docs/openapi.json" class="spec-link" target="_blank">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
          <polyline points="7 10 12 15 17 10"/>
          <line x1="12" y1="15" x2="12" y2="3"/>
        </svg>
        OpenAPI JSON
      </a>
      <a href="/health" class="spec-link" target="_blank">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M22 12h-4l-3 9L9 3l-3 9H2"/>
        </svg>
        Health Check
      </a>
    </div>

    <footer class="footer">
      <span>Powered by</span>
      <span class="footer-tech">Go</span>
      <span class="footer-tech">Whatsmeow</span>
      <span class="footer-tech">OpenAPI 3.1</span>
    </footer>
  </div>

  <script>
    (function() {
      var dot = document.getElementById('status-dot');
      var text = document.getElementById('status-text');
      function check() {
        fetch('/health', { method: 'GET', cache: 'no-store' })
          .then(function(r) {
            if (r.ok) {
              dot.className = 'status-dot online';
              text.textContent = 'Online';
            } else {
              dot.className = 'status-dot offline';
              text.textContent = 'Degraded';
            }
          })
          .catch(function() {
            dot.className = 'status-dot offline';
            text.textContent = 'Offline';
          });
      }
      check();
      setInterval(check, 30000);
    })();
  </script>
</body>
</html>`, ver)

	return html, nil
}

// LandingHandler returns an HTTP handler that serves the landing page.
func LandingHandler(cfg Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		landingOnce.Do(func() {
			landingHTML, landingErr = generateLandingHTML(cfg.BaseURL)
		})
		if landingErr != nil {
			http.Error(w, fmt.Sprintf("failed to generate landing page: %v", landingErr), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(landingHTML))
	})
}
