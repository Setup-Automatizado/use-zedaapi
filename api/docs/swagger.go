package docs

import (
	"fmt"
	"net/http"
	"sync"
)

const swaggerUIVersion = "5.21.0"

var (
	swaggerOnce sync.Once
	swaggerHTML string
	swaggerErr  error
)

func generateSwaggerHTML(baseURL string) (string, error) {
	specURL := "/docs/openapi.json"

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Swagger UI - Zé da API</title>
  <link rel="icon" type="image/x-icon" href="/favicon.ico">
  <link rel="icon" type="image/png" sizes="32x32" href="/favicon-32x32.png">
  <link rel="icon" type="image/png" sizes="16x16" href="/favicon-16x16.png">
  <link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png">
  <link rel="manifest" href="/manifest.json">
  <meta name="theme-color" content="#1a1a2e">
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@%[1]s/swagger-ui.css">
  <style>
    *,
    *::before,
    *::after {
      box-sizing: border-box;
    }

    body {
      margin: 0;
      padding: 0;
      background: #1a1a2e;
      font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    }

    /* ── Navigation Bar ────────────────────────────────── */
    .docs-nav {
      position: sticky;
      top: 0;
      z-index: 1000;
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0 24px;
      height: 48px;
      background: #0f0f1a;
      border-bottom: 1px solid rgba(255, 255, 255, 0.08);
      backdrop-filter: blur(12px);
      -webkit-backdrop-filter: blur(12px);
    }

    .docs-nav-left {
      display: flex;
      align-items: center;
      gap: 20px;
    }

    .docs-nav-brand {
      font-size: 13px;
      font-weight: 600;
      color: #e2e8f0;
      letter-spacing: -0.01em;
      text-decoration: none;
      white-space: nowrap;
    }

    .docs-nav-tabs {
      display: flex;
      align-items: center;
      gap: 2px;
      background: rgba(255, 255, 255, 0.04);
      border-radius: 8px;
      padding: 3px;
    }

    .docs-nav-tab {
      display: flex;
      align-items: center;
      gap: 6px;
      padding: 5px 14px;
      border-radius: 6px;
      font-size: 12px;
      font-weight: 500;
      color: #94a3b8;
      text-decoration: none;
      transition: all 0.15s ease;
      white-space: nowrap;
    }

    .docs-nav-tab:hover {
      color: #e2e8f0;
      background: rgba(255, 255, 255, 0.06);
    }

    .docs-nav-tab.active {
      color: #fff;
      background: rgba(255, 255, 255, 0.1);
      box-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
    }

    .docs-nav-tab svg {
      width: 14px;
      height: 14px;
      flex-shrink: 0;
    }

    .docs-nav-right {
      display: flex;
      align-items: center;
      gap: 12px;
    }

    .docs-nav-spec {
      display: flex;
      align-items: center;
      gap: 6px;
      padding: 5px 12px;
      border-radius: 6px;
      font-size: 11px;
      font-weight: 500;
      color: #64748b;
      text-decoration: none;
      transition: all 0.15s ease;
      border: 1px solid rgba(255, 255, 255, 0.06);
    }

    .docs-nav-spec:hover {
      color: #94a3b8;
      border-color: rgba(255, 255, 255, 0.12);
      background: rgba(255, 255, 255, 0.04);
    }

    .docs-nav-spec svg {
      width: 12px;
      height: 12px;
    }

    /* ── Swagger UI Overrides (Dark Theme) ─────────────── */
    .swagger-ui {
      background: #1a1a2e;
      color: #e2e8f0;
    }

    .swagger-ui .topbar {
      display: none;
    }

    .swagger-ui .wrapper {
      max-width: 1400px;
      padding: 24px 32px;
    }

    .swagger-ui .info {
      margin: 16px 0 32px;
    }

    .swagger-ui .info .title {
      color: #f1f5f9 !important;
      font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
      font-weight: 700;
    }

    .swagger-ui .info .description,
    .swagger-ui .info .description p {
      color: #94a3b8 !important;
      font-size: 14px;
      line-height: 1.6;
    }

    .swagger-ui .info a {
      color: #818cf8;
    }

    .swagger-ui .scheme-container {
      background: #0f0f1a;
      box-shadow: none;
      border: 1px solid rgba(255, 255, 255, 0.06);
      border-radius: 12px;
      padding: 16px 20px;
    }

    .swagger-ui .opblock-tag {
      color: #e2e8f0;
      border-bottom: 1px solid rgba(255, 255, 255, 0.06);
      font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
    }

    .swagger-ui .opblock-tag:hover {
      background: rgba(255, 255, 255, 0.02);
    }

    .swagger-ui .opblock-tag small {
      color: #64748b;
    }

    /* Method blocks */
    .swagger-ui .opblock {
      border-radius: 10px;
      border: 1px solid rgba(255, 255, 255, 0.06);
      box-shadow: none;
      margin-bottom: 8px;
      background: rgba(255, 255, 255, 0.02);
    }

    .swagger-ui .opblock .opblock-summary {
      border-bottom: none;
      padding: 10px 16px;
    }

    .swagger-ui .opblock .opblock-summary-method {
      border-radius: 6px;
      font-size: 11px;
      font-weight: 700;
      font-family: 'Inter', monospace;
      min-width: 64px;
      padding: 6px 0;
      text-align: center;
    }

    .swagger-ui .opblock .opblock-summary-path {
      color: #e2e8f0 !important;
      font-family: 'JetBrains Mono', 'SF Mono', 'Fira Code', monospace;
      font-size: 13px;
    }

    .swagger-ui .opblock .opblock-summary-description {
      color: #94a3b8;
      font-size: 13px;
    }

    .swagger-ui .opblock.opblock-get {
      background: rgba(56, 189, 248, 0.04);
      border-color: rgba(56, 189, 248, 0.15);
    }

    .swagger-ui .opblock.opblock-get .opblock-summary-method {
      background: #0ea5e9;
    }

    .swagger-ui .opblock.opblock-post {
      background: rgba(74, 222, 128, 0.04);
      border-color: rgba(74, 222, 128, 0.15);
    }

    .swagger-ui .opblock.opblock-post .opblock-summary-method {
      background: #22c55e;
    }

    .swagger-ui .opblock.opblock-put {
      background: rgba(251, 146, 60, 0.04);
      border-color: rgba(251, 146, 60, 0.15);
    }

    .swagger-ui .opblock.opblock-put .opblock-summary-method {
      background: #f97316;
    }

    .swagger-ui .opblock.opblock-delete {
      background: rgba(248, 113, 113, 0.04);
      border-color: rgba(248, 113, 113, 0.15);
    }

    .swagger-ui .opblock.opblock-delete .opblock-summary-method {
      background: #ef4444;
    }

    .swagger-ui .opblock.opblock-patch {
      background: rgba(192, 132, 252, 0.04);
      border-color: rgba(192, 132, 252, 0.15);
    }

    .swagger-ui .opblock.opblock-patch .opblock-summary-method {
      background: #a855f7;
    }

    /* Expanded operation */
    .swagger-ui .opblock-body {
      background: rgba(0, 0, 0, 0.15);
    }

    .swagger-ui .opblock-body pre {
      background: #0f0f1a !important;
      color: #e2e8f0 !important;
      border: 1px solid rgba(255, 255, 255, 0.06);
      border-radius: 8px;
      font-family: 'JetBrains Mono', 'SF Mono', 'Fira Code', monospace;
      font-size: 12px;
      padding: 16px;
    }

    .swagger-ui .opblock-description-wrapper p,
    .swagger-ui .opblock-external-docs-wrapper p {
      color: #94a3b8;
      font-size: 13px;
    }

    /* Parameters */
    .swagger-ui .parameters-col_description p {
      color: #94a3b8;
      font-size: 13px;
    }

    .swagger-ui .parameter__name {
      color: #e2e8f0;
      font-family: 'JetBrains Mono', 'SF Mono', monospace;
      font-size: 13px;
    }

    .swagger-ui .parameter__name.required::after {
      color: #ef4444;
    }

    .swagger-ui .parameter__type {
      color: #818cf8;
      font-family: 'JetBrains Mono', 'SF Mono', monospace;
      font-size: 12px;
    }

    .swagger-ui .parameter__in {
      color: #64748b;
      font-size: 11px;
    }

    .swagger-ui table thead tr th,
    .swagger-ui table thead tr td {
      color: #94a3b8;
      border-bottom: 1px solid rgba(255, 255, 255, 0.06);
    }

    .swagger-ui table tbody tr td {
      color: #e2e8f0;
      border-bottom: 1px solid rgba(255, 255, 255, 0.04);
    }

    /* Inputs */
    .swagger-ui input[type="text"],
    .swagger-ui textarea,
    .swagger-ui select {
      background: #0f0f1a;
      color: #e2e8f0;
      border: 1px solid rgba(255, 255, 255, 0.1);
      border-radius: 6px;
      padding: 8px 12px;
      font-family: 'JetBrains Mono', 'SF Mono', monospace;
      font-size: 13px;
      transition: border-color 0.15s ease;
    }

    .swagger-ui input[type="text"]:focus,
    .swagger-ui textarea:focus,
    .swagger-ui select:focus {
      border-color: #818cf8;
      outline: none;
      box-shadow: 0 0 0 2px rgba(129, 140, 248, 0.15);
    }

    .swagger-ui select {
      appearance: none;
      background-image: url("data:image/svg+xml;charset=utf-8,%%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' fill='%%2394a3b8'%%3E%%3Cpath d='M6 8.5L1 3.5h10z'%%3E%%3C/path%%3E%%3C/svg%%3E");
      background-repeat: no-repeat;
      background-position: right 12px center;
      padding-right: 32px;
    }

    /* Buttons */
    .swagger-ui .btn {
      border-radius: 6px;
      font-weight: 500;
      font-size: 12px;
      transition: all 0.15s ease;
    }

    .swagger-ui .btn.execute {
      background: #818cf8;
      border-color: #818cf8;
      color: #fff;
      font-weight: 600;
      border-radius: 8px;
      padding: 8px 24px;
    }

    .swagger-ui .btn.execute:hover {
      background: #6366f1;
      border-color: #6366f1;
    }

    .swagger-ui .btn.cancel {
      border-color: rgba(255, 255, 255, 0.15);
      color: #94a3b8;
    }

    .swagger-ui .btn.authorize {
      color: #22c55e;
      border-color: #22c55e;
    }

    .swagger-ui .btn.authorize svg {
      fill: #22c55e;
    }

    /* Try it out button */
    .swagger-ui .try-out__btn {
      border-color: #818cf8;
      color: #818cf8;
      border-radius: 6px;
      font-weight: 500;
    }

    .swagger-ui .try-out__btn:hover {
      background: rgba(129, 140, 248, 0.1);
    }

    /* Responses */
    .swagger-ui .responses-inner {
      padding: 12px;
    }

    .swagger-ui .responses-table {
      padding: 0;
    }

    .swagger-ui .response-col_status {
      color: #e2e8f0;
      font-family: 'JetBrains Mono', 'SF Mono', monospace;
      font-weight: 600;
    }

    .swagger-ui .response-col_description {
      color: #94a3b8;
    }

    .swagger-ui .response-col_links {
      color: #64748b;
    }

    /* Models */
    .swagger-ui section.models {
      border: 1px solid rgba(255, 255, 255, 0.06);
      border-radius: 12px;
      background: rgba(255, 255, 255, 0.02);
    }

    .swagger-ui section.models h4 {
      color: #e2e8f0;
    }

    .swagger-ui .model-container {
      background: rgba(0, 0, 0, 0.15);
      border-radius: 8px;
      margin: 4px 0;
    }

    .swagger-ui .model {
      color: #e2e8f0;
      font-family: 'JetBrains Mono', 'SF Mono', monospace;
      font-size: 12px;
    }

    .swagger-ui .model .property {
      color: #94a3b8;
    }

    .swagger-ui .model .property.primitive {
      color: #818cf8;
    }

    .swagger-ui .model-title {
      color: #e2e8f0;
      font-family: 'JetBrains Mono', 'SF Mono', monospace;
    }

    /* Auth modal */
    .swagger-ui .dialog-ux .modal-ux {
      background: #1a1a2e;
      border: 1px solid rgba(255, 255, 255, 0.1);
      border-radius: 16px;
      box-shadow: 0 25px 50px rgba(0, 0, 0, 0.5);
    }

    .swagger-ui .dialog-ux .modal-ux-header {
      border-bottom: 1px solid rgba(255, 255, 255, 0.06);
    }

    .swagger-ui .dialog-ux .modal-ux-header h3 {
      color: #f1f5f9;
    }

    .swagger-ui .dialog-ux .modal-ux-content {
      color: #94a3b8;
    }

    .swagger-ui .dialog-ux .modal-ux-content p {
      color: #94a3b8;
    }

    .swagger-ui .dialog-ux .modal-ux-content h4 {
      color: #e2e8f0;
    }

    /* Scrollbars */
    .swagger-ui ::-webkit-scrollbar {
      width: 6px;
      height: 6px;
    }

    .swagger-ui ::-webkit-scrollbar-track {
      background: transparent;
    }

    .swagger-ui ::-webkit-scrollbar-thumb {
      background: rgba(255, 255, 255, 0.1);
      border-radius: 3px;
    }

    .swagger-ui ::-webkit-scrollbar-thumb:hover {
      background: rgba(255, 255, 255, 0.2);
    }

    /* Markdown content */
    .swagger-ui .markdown p,
    .swagger-ui .renderedMarkdown p {
      color: #94a3b8;
    }

    .swagger-ui .markdown code,
    .swagger-ui .renderedMarkdown code {
      background: rgba(129, 140, 248, 0.1);
      color: #818cf8;
      padding: 2px 6px;
      border-radius: 4px;
      font-family: 'JetBrains Mono', 'SF Mono', monospace;
      font-size: 12px;
    }

    .swagger-ui .markdown a,
    .swagger-ui .renderedMarkdown a {
      color: #818cf8;
    }

    /* Copy button */
    .swagger-ui .copy-to-clipboard {
      bottom: 8px;
      right: 8px;
    }

    .swagger-ui .copy-to-clipboard button {
      background: rgba(255, 255, 255, 0.06);
      border: 1px solid rgba(255, 255, 255, 0.08);
      border-radius: 6px;
      padding: 4px 8px;
    }

    /* Loading */
    .swagger-ui .loading-container {
      padding: 80px 0;
    }

    .swagger-ui .loading-container .loading::after {
      color: #818cf8;
    }

    /* Response codes */
    .swagger-ui .responses-wrapper .responses-inner .response .response-col_status {
      font-size: 13px;
    }

    /* Tab headers for response content types */
    .swagger-ui .tab li {
      color: #94a3b8;
    }

    .swagger-ui .tab li.active {
      color: #e2e8f0;
    }

    .swagger-ui .tab li button.tablinks {
      color: inherit;
    }

    /* Server dropdown */
    .swagger-ui .servers > label {
      color: #94a3b8;
    }

    .swagger-ui .servers > label select {
      border: 1px solid rgba(255, 255, 255, 0.1);
    }

    /* Links */
    .swagger-ui a {
      color: #818cf8;
    }

    /* Expand/collapse toggle */
    .swagger-ui .expand-operation svg,
    .swagger-ui .models-control svg {
      fill: #64748b;
    }

    .swagger-ui .expand-operation:hover svg,
    .swagger-ui .models-control:hover svg {
      fill: #94a3b8;
    }

    /* JSON schema */
    .swagger-ui .json-schema-form-item input {
      background: #0f0f1a;
      color: #e2e8f0;
      border-color: rgba(255, 255, 255, 0.1);
    }

    /* Highlighted code */
    .swagger-ui .highlight-code {
      background: #0f0f1a !important;
      border-radius: 8px;
    }

    .swagger-ui .highlight-code .microlight {
      background: transparent !important;
      color: #e2e8f0 !important;
      font-family: 'JetBrains Mono', 'SF Mono', 'Fira Code', monospace;
      font-size: 12px;
      line-height: 1.5;
    }

    /* Authorization lock icons */
    .swagger-ui .authorization__btn svg {
      fill: #64748b;
    }

    .swagger-ui .authorization__btn.locked svg {
      fill: #22c55e;
    }

    .swagger-ui .authorization__btn.unlocked svg {
      fill: #94a3b8;
    }

    /* Filter */
    .swagger-ui .filter-container {
      background: #0f0f1a;
      border-bottom: 1px solid rgba(255, 255, 255, 0.06);
      padding: 12px 20px;
    }

    .swagger-ui .filter-container input[type="text"] {
      background: rgba(255, 255, 255, 0.04);
      border: 1px solid rgba(255, 255, 255, 0.08);
    }

    /* ── Hero Cover ─────────────────────────────────── */
    .docs-hero {
      position: relative;
      overflow: hidden;
      padding: 48px 32px 40px;
      background: linear-gradient(135deg, #0f0f1a 0%%, #1e1b4b 40%%, #312e81 70%%, #4338ca 100%%);
    }

    .docs-hero::before {
      content: '';
      position: absolute;
      inset: 0;
      background: radial-gradient(ellipse 80%% 50%% at 50%% -20%%, rgba(129, 140, 248, 0.15), transparent);
      pointer-events: none;
    }

    .docs-hero::after {
      content: '';
      position: absolute;
      bottom: 0;
      left: 0;
      right: 0;
      height: 1px;
      background: linear-gradient(90deg, transparent, rgba(129, 140, 248, 0.3), transparent);
    }

    .docs-hero-inner {
      position: relative;
      max-width: 1400px;
      margin: 0 auto;
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 32px;
    }

    .docs-hero-content {
      flex: 1;
      min-width: 0;
    }

    .docs-hero-badge {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      padding: 4px 12px 4px 8px;
      background: rgba(129, 140, 248, 0.12);
      border: 1px solid rgba(129, 140, 248, 0.2);
      border-radius: 20px;
      font-size: 11px;
      font-weight: 500;
      color: #a5b4fc;
      margin-bottom: 16px;
    }

    .docs-hero-badge-dot {
      width: 6px;
      height: 6px;
      border-radius: 50%%;
      background: #22c55e;
      box-shadow: 0 0 6px rgba(34, 197, 94, 0.5);
    }

    .docs-hero h1 {
      font-size: 28px;
      font-weight: 700;
      color: #fff;
      margin: 0 0 8px;
      letter-spacing: -0.02em;
      line-height: 1.2;
    }

    .docs-hero p {
      font-size: 14px;
      color: #a5b4fc;
      margin: 0;
      line-height: 1.5;
      max-width: 520px;
    }

    .docs-hero-meta {
      display: flex;
      align-items: center;
      gap: 16px;
      margin-top: 20px;
    }

    .docs-hero-tag {
      display: inline-flex;
      align-items: center;
      gap: 5px;
      padding: 4px 10px;
      background: rgba(255, 255, 255, 0.06);
      border: 1px solid rgba(255, 255, 255, 0.08);
      border-radius: 6px;
      font-size: 11px;
      font-weight: 500;
      color: #cbd5e1;
    }

    .docs-hero-tag svg {
      width: 12px;
      height: 12px;
      opacity: 0.7;
    }

    .docs-hero-visual {
      flex-shrink: 0;
      width: 200px;
      height: 200px;
      display: flex;
      align-items: center;
      justify-content: center;
    }

    .docs-hero-img {
      width: 180px;
      height: 180px;
      border-radius: 20px;
      object-fit: cover;
      border: 2px solid rgba(129, 140, 248, 0.2);
      box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4), 0 0 60px rgba(129, 140, 248, 0.08);
      transition: transform 0.3s ease, box-shadow 0.3s ease;
    }

    .docs-hero-img:hover {
      transform: scale(1.04);
      box-shadow: 0 12px 40px rgba(0, 0, 0, 0.5), 0 0 80px rgba(129, 140, 248, 0.12);
    }

    /* ── Swagger topbar override ─────────────────────── */
    .swagger-ui .topbar-wrapper .topbar__inner {
      display: none;
    }

    /* Hide default Swagger info since hero replaces it */
    .swagger-ui .info hgroup.main {
      display: none;
    }

    /* Responsive */
    @media (max-width: 768px) {
      .docs-nav {
        padding: 0 12px;
        height: 44px;
      }

      .docs-nav-brand {
        display: none;
      }

      .docs-nav-right {
        display: none;
      }

      .swagger-ui .wrapper {
        padding: 16px;
      }

      .docs-hero {
        padding: 32px 16px 28px;
      }

      .docs-hero h1 {
        font-size: 22px;
      }

      .docs-hero-visual {
        display: none;
      }

      .docs-hero-meta {
        flex-wrap: wrap;
        gap: 8px;
      }
    }
  </style>
</head>
<body>
  <nav class="docs-nav">
    <div class="docs-nav-left">
      <a href="/docs/" class="docs-nav-brand">Zé da API</a>
      <div class="docs-nav-tabs">
        <a href="/docs/" class="docs-nav-tab">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z"/>
            <polyline points="14 2 14 8 20 8"/>
            <line x1="16" y1="13" x2="8" y2="13"/>
            <line x1="16" y1="17" x2="8" y2="17"/>
            <line x1="10" y1="9" x2="8" y2="9"/>
          </svg>
          Scalar
        </a>
        <a href="/docs/swagger/" class="docs-nav-tab active">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="16 18 22 12 16 6"/>
            <polyline points="8 6 2 12 8 18"/>
          </svg>
          Swagger
        </a>
      </div>
    </div>
    <div class="docs-nav-right">
      <a href="/docs/openapi.yaml" class="docs-nav-spec" target="_blank">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
          <polyline points="7 10 12 15 17 10"/>
          <line x1="12" y1="15" x2="12" y2="3"/>
        </svg>
        YAML
      </a>
      <a href="/docs/openapi.json" class="docs-nav-spec" target="_blank">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
          <polyline points="7 10 12 15 17 10"/>
          <line x1="12" y1="15" x2="12" y2="3"/>
        </svg>
        JSON
      </a>
    </div>
  </nav>

  <section class="docs-hero">
    <div class="docs-hero-inner">
      <div class="docs-hero-content">
        <div class="docs-hero-badge">
          <span class="docs-hero-badge-dot"></span>
          Swagger UI
        </div>
        <h1>Zé da API</h1>
        <p>WhatsApp Business API powered by Whatsmeow. Interactive endpoint explorer with live request testing.</p>
        <div class="docs-hero-meta">
          <span class="docs-hero-tag">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
            </svg>
            OpenAPI 3.1
          </span>
          <span class="docs-hero-tag">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <polyline points="12 6 12 12 16 14"/>
            </svg>
            Try It Out
          </span>
          <span class="docs-hero-tag">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
              <path d="M7 11V7a5 5 0 0 1 10 0v4"/>
            </svg>
            Auth Ready
          </span>
        </div>
      </div>
      <div class="docs-hero-visual">
        <img src="/capa.png" alt="Zé da API" class="docs-hero-img" />
      </div>
    </div>
  </section>

  <div id="swagger-ui"></div>

  <script src="https://unpkg.com/swagger-ui-dist@%[1]s/swagger-ui-bundle.js"></script>
  <script src="https://unpkg.com/swagger-ui-dist@%[1]s/swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function() {
      SwaggerUIBundle({
        url: "%[2]s",
        dom_id: "#swagger-ui",
        deepLinking: true,
        filter: true,
        showExtensions: true,
        showCommonExtensions: true,
        tryItOutEnabled: true,
        requestSnippetsEnabled: true,
        persistAuthorization: true,
        withCredentials: false,
        displayRequestDuration: true,
        defaultModelsExpandDepth: 2,
        defaultModelExpandDepth: 3,
        docExpansion: "list",
        syntaxHighlight: {
          activated: true,
          theme: "monokai"
        },
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout",
        validatorUrl: null
      });
    };
  </script>
</body>
</html>`, swaggerUIVersion, specURL)

	return html, nil
}

// SwaggerUIHandler returns an HTTP handler that serves the Swagger UI.
func SwaggerUIHandler(cfg Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		swaggerOnce.Do(func() {
			swaggerHTML, swaggerErr = generateSwaggerHTML(cfg.BaseURL)
		})
		if swaggerErr != nil {
			http.Error(w, fmt.Sprintf("failed to generate Swagger UI: %v", swaggerErr), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(swaggerHTML))
	})
}

// navHeaderCSS returns the CSS for the navigation header used by both UIs.
func navHeaderCSS() string {
	return `
    <style>
      .docs-nav {
        position: sticky;
        top: 0;
        z-index: 1000;
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0 24px;
        height: 48px;
        background: #0f0f1a;
        border-bottom: 1px solid rgba(255, 255, 255, 0.08);
        backdrop-filter: blur(12px);
        -webkit-backdrop-filter: blur(12px);
        font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      }
      .docs-nav-left {
        display: flex;
        align-items: center;
        gap: 20px;
      }
      .docs-nav-brand {
        font-size: 13px;
        font-weight: 600;
        color: #e2e8f0;
        letter-spacing: -0.01em;
        text-decoration: none;
        white-space: nowrap;
      }
      .docs-nav-tabs {
        display: flex;
        align-items: center;
        gap: 2px;
        background: rgba(255, 255, 255, 0.04);
        border-radius: 8px;
        padding: 3px;
      }
      .docs-nav-tab {
        display: flex;
        align-items: center;
        gap: 6px;
        padding: 5px 14px;
        border-radius: 6px;
        font-size: 12px;
        font-weight: 500;
        color: #94a3b8;
        text-decoration: none;
        transition: all 0.15s ease;
        white-space: nowrap;
      }
      .docs-nav-tab:hover {
        color: #e2e8f0;
        background: rgba(255, 255, 255, 0.06);
      }
      .docs-nav-tab.active {
        color: #fff;
        background: rgba(255, 255, 255, 0.1);
        box-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
      }
      .docs-nav-tab svg {
        width: 14px;
        height: 14px;
        flex-shrink: 0;
      }
      .docs-nav-right {
        display: flex;
        align-items: center;
        gap: 12px;
      }
      .docs-nav-spec {
        display: flex;
        align-items: center;
        gap: 6px;
        padding: 5px 12px;
        border-radius: 6px;
        font-size: 11px;
        font-weight: 500;
        color: #64748b;
        text-decoration: none;
        transition: all 0.15s ease;
        border: 1px solid rgba(255, 255, 255, 0.06);
      }
      .docs-nav-spec:hover {
        color: #94a3b8;
        border-color: rgba(255, 255, 255, 0.12);
        background: rgba(255, 255, 255, 0.04);
      }
      .docs-nav-spec svg {
        width: 12px;
        height: 12px;
      }
      @media (max-width: 768px) {
        .docs-nav { padding: 0 12px; height: 44px; }
        .docs-nav-brand { display: none; }
        .docs-nav-right { display: none; }
      }
    </style>`
}

// navHeaderHTML returns the HTML for the navigation header.
// activeTab should be "scalar" or "swagger".
func navHeaderHTML(activeTab string) string {
	scalarActive := ""
	swaggerActive := ""
	if activeTab == "scalar" {
		scalarActive = " active"
	} else {
		swaggerActive = " active"
	}

	return fmt.Sprintf(`<nav class="docs-nav">
    <div class="docs-nav-left">
      <a href="/docs/" class="docs-nav-brand">Zé da API</a>
      <div class="docs-nav-tabs">
        <a href="/docs/" class="docs-nav-tab%[1]s">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z"/>
            <polyline points="14 2 14 8 20 8"/>
            <line x1="16" y1="13" x2="8" y2="13"/>
            <line x1="16" y1="17" x2="8" y2="17"/>
            <line x1="10" y1="9" x2="8" y2="9"/>
          </svg>
          Scalar
        </a>
        <a href="/docs/swagger/" class="docs-nav-tab%[2]s">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="16 18 22 12 16 6"/>
            <polyline points="8 6 2 12 8 18"/>
          </svg>
          Swagger
        </a>
      </div>
    </div>
    <div class="docs-nav-right">
      <a href="/docs/openapi.yaml" class="docs-nav-spec" target="_blank">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
          <polyline points="7 10 12 15 17 10"/>
          <line x1="12" y1="15" x2="12" y2="3"/>
        </svg>
        YAML
      </a>
      <a href="/docs/openapi.json" class="docs-nav-spec" target="_blank">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
          <polyline points="7 10 12 15 17 10"/>
          <line x1="12" y1="15" x2="12" y2="3"/>
        </svg>
        JSON
      </a>
    </div>
  </nav>`, scalarActive, swaggerActive)

}
