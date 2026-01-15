import { Database } from "bun:sqlite";

// ═══════════════════════════════════════════════════════════════════════════════
// DATABASE - SQLite with bun:sqlite
// ═══════════════════════════════════════════════════════════════════════════════

const db = new Database("webhooks.db", { create: true });
db.exec("PRAGMA journal_mode = WAL");
db.exec("PRAGMA synchronous = NORMAL");

db.exec(`
  CREATE TABLE IF NOT EXISTS events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_type TEXT NOT NULL,
    content_type TEXT,
    notification TEXT,
    instance_id TEXT,
    phone TEXT,
    valid INTEGER NOT NULL,
    errors TEXT,
    payload TEXT NOT NULL,
    created_at TEXT DEFAULT (datetime('now'))
  )
`);

db.exec(`
  CREATE TABLE IF NOT EXISTS stats (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    total INTEGER DEFAULT 0,
    valid INTEGER DEFAULT 0,
    invalid INTEGER DEFAULT 0,
    started_at TEXT DEFAULT (datetime('now'))
  )
`);

db.exec(`INSERT OR IGNORE INTO stats (id) VALUES (1)`);

// ═══════════════════════════════════════════════════════════════════════════════
// EVENT TYPES & VALIDATION - Based on Z-API schemas
// ═══════════════════════════════════════════════════════════════════════════════

const EventTypes = [
	"ReceivedCallback",
	"DeliveryCallback",
	"MessageStatusCallback",
	"PresenceChatCallback",
	"ConnectedCallback",
	"DisconnectedCallback",
] as const;
type EventType = (typeof EventTypes)[number];

const RequiredFields: Record<EventType, string[]> = {
	ReceivedCallback: [
		"type",
		"instanceId",
		"messageId",
		"phone",
		"momment",
		"status",
		"fromMe",
		"isGroup",
	],
	DeliveryCallback: ["type", "instanceId", "messageId", "phone", "zaapId"],
	MessageStatusCallback: [
		"type",
		"instanceId",
		"status",
		"ids",
		"phone",
		"momment",
		"isGroup",
	],
	PresenceChatCallback: ["type", "instanceId", "phone", "status"],
	ConnectedCallback: ["type", "instanceId", "connected", "momment"],
	DisconnectedCallback: ["type", "instanceId", "disconnected", "momment"],
};

const ContentTypes = [
	"text",
	"image",
	"audio",
	"video",
	"document",
	"sticker",
	"location",
	"contact",
	"reaction",
	"poll",
	"pollVote",
	"buttonsResponseMessage",
	"listResponseMessage",
	"hydratedTemplate",
	"buttonsMessage",
	"pixKeyMessage",
	"carouselMessage",
	"product",
	"order",
	"reviewAndPay",
	"reviewOrder",
	"requestPayment",
	"sendPayment",
	"newsletterAdminInvite",
	"pinMessage",
	"event",
	"eventResponse",
	"externalAdReply",
] as const;

const ContentRequiredFields: Record<string, string[]> = {
	text: ["message"],
	image: ["mimeType", "imageUrl"],
	audio: ["mimeType", "audioUrl"],
	video: ["videoUrl", "mimeType"],
	document: ["documentUrl", "mimeType"],
	sticker: ["stickerUrl", "mimeType"],
	location: ["longitude", "latitude"],
	contact: ["displayName", "vCard"],
	reaction: ["value", "time", "reactionBy", "referencedMessage"],
	poll: ["question", "options"],
};

const MessageStatuses = [
	"PENDING",
	"SENT",
	"RECEIVED",
	"READ",
	"READ_BY_ME",
	"PLAYED",
	"PLAYED_BY_ME",
] as const;
const PresenceStatuses = [
	"UNAVAILABLE",
	"AVAILABLE",
	"COMPOSING",
	"PAUSED",
	"RECORDING",
] as const;

interface ValidationResult {
	valid: boolean;
	eventType: string;
	contentType?: string;
	notification?: string;
	errors: string[];
}

function validateEvent(payload: Record<string, unknown>): ValidationResult {
	const errors: string[] = [];
	const eventType = payload.type as string;

	if (!eventType || !EventTypes.includes(eventType as EventType)) {
		return {
			valid: false,
			eventType: eventType || "Unknown",
			errors: [`Invalid event type: ${eventType}`],
		};
	}

	for (const field of RequiredFields[eventType as EventType]) {
		if (payload[field] === undefined || payload[field] === null)
			errors.push(`Missing: ${field}`);
	}

	let contentType: string | undefined;
	let notification: string | undefined;

	if (eventType === "ReceivedCallback") {
		contentType = ContentTypes.find((ct) => payload[ct] !== undefined);
		notification = payload.notification as string;
		if (contentType) {
			const content = payload[contentType] as Record<string, unknown>;
			const required = ContentRequiredFields[contentType];
			if (required && content) {
				for (const f of required) {
					if (content[f] === undefined)
						errors.push(`${contentType}.${f} required`);
				}
			}
		}
	} else if (eventType === "MessageStatusCallback") {
		if (
			!MessageStatuses.includes(
				payload.status as (typeof MessageStatuses)[number],
			)
		) {
			errors.push(`Invalid status: ${payload.status}`);
		}
	} else if (eventType === "PresenceChatCallback") {
		if (
			!PresenceStatuses.includes(
				payload.status as (typeof PresenceStatuses)[number],
			)
		) {
			errors.push(`Invalid presence: ${payload.status}`);
		}
	}

	return {
		valid: errors.length === 0,
		eventType,
		contentType,
		notification,
		errors,
	};
}

// ═══════════════════════════════════════════════════════════════════════════════
// DATABASE OPERATIONS
// ═══════════════════════════════════════════════════════════════════════════════

const insertEvent = db.prepare(`
  INSERT INTO events (event_type, content_type, notification, instance_id, phone, valid, errors, payload)
  VALUES (?, ?, ?, ?, ?, ?, ?, ?)
`);

const updateStats = db.prepare(
	`UPDATE stats SET total = total + 1, valid = valid + ?, invalid = invalid + ? WHERE id = 1`,
);

function saveEvent(
	payload: Record<string, unknown>,
	validation: ValidationResult,
): number {
	const result = insertEvent.run(
		validation.eventType,
		validation.contentType || null,
		validation.notification || null,
		(payload.instanceId as string) || null,
		(payload.phone as string) || null,
		validation.valid ? 1 : 0,
		validation.errors.length > 0 ? JSON.stringify(validation.errors) : null,
		JSON.stringify(payload),
	);
	updateStats.run(validation.valid ? 1 : 0, validation.valid ? 0 : 1);
	return Number(result.lastInsertRowid);
}

function getStats() {
	const stats = db.query(`SELECT * FROM stats WHERE id = 1`).get() as Record<
		string,
		unknown
	>;
	const byType = db
		.query(
			`SELECT event_type, COUNT(*) as total, SUM(valid) as valid, COUNT(*) - SUM(valid) as invalid FROM events GROUP BY event_type`,
		)
		.all();
	const byContent = db
		.query(
			`SELECT content_type, COUNT(*) as count FROM events WHERE content_type IS NOT NULL GROUP BY content_type ORDER BY count DESC`,
		)
		.all();
	const recent = db
		.query(
			`SELECT id, event_type, content_type, notification, instance_id, phone, valid, errors, payload, created_at FROM events ORDER BY id DESC LIMIT 50`,
		)
		.all();
	return { ...stats, byType, byContent, recent };
}

// ═══════════════════════════════════════════════════════════════════════════════
// WEBSOCKET CONNECTIONS
// ═══════════════════════════════════════════════════════════════════════════════

const clients = new Set<{ send: (data: string) => void }>();

function broadcast(data: unknown) {
	const msg = JSON.stringify(data);
	for (const client of clients) {
		try {
			client.send(msg);
		} catch {
			clients.delete(client);
		}
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// FRONTEND HTML
// ═══════════════════════════════════════════════════════════════════════════════

const HTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Webhook Monitor</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body { font-family: system-ui, -apple-system, sans-serif; background: #0a0a0a; color: #e5e5e5; min-height: 100vh; }
    .container { max-width: 1600px; margin: 0 auto; padding: 20px; }
    header { background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%); padding: 24px; border-radius: 12px; margin-bottom: 20px; border: 1px solid #2a2a4a; }
    header h1 { font-size: 1.5rem; color: #00d9ff; margin-bottom: 8px; }
    header p { color: #888; font-size: 0.9rem; }
    .status { display: inline-flex; align-items: center; gap: 8px; padding: 6px 12px; border-radius: 20px; font-size: 0.8rem; }
    .status.connected { background: #0d3320; color: #22c55e; }
    .status.disconnected { background: #3d1515; color: #ef4444; }
    .status::before { content: ''; width: 8px; height: 8px; border-radius: 50%; background: currentColor; animation: pulse 2s infinite; }
    @keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.5; } }
    .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 16px; margin-bottom: 20px; }
    .card { background: #141414; border: 1px solid #2a2a2a; border-radius: 12px; padding: 20px; }
    .card h3 { font-size: 0.75rem; text-transform: uppercase; color: #666; margin-bottom: 8px; letter-spacing: 0.5px; }
    .card .value { font-size: 2rem; font-weight: 700; }
    .card .value.total { color: #00d9ff; }
    .card .value.valid { color: #22c55e; }
    .card .value.invalid { color: #ef4444; }
    .card .value.rate { color: #f59e0b; }
    .section { background: #141414; border: 1px solid #2a2a2a; border-radius: 12px; margin-bottom: 20px; overflow: hidden; }
    .section-header { padding: 16px 20px; border-bottom: 1px solid #2a2a2a; display: flex; justify-content: space-between; align-items: center; }
    .section-header h2 { font-size: 1rem; color: #e5e5e5; }
    .badges { display: flex; flex-wrap: wrap; gap: 8px; padding: 16px 20px; }
    .badge { padding: 8px 14px; border-radius: 8px; font-size: 0.8rem; font-weight: 500; }
    .badge.type { background: #1e3a5f; color: #60a5fa; }
    .badge.content { background: #3d1f5c; color: #c084fc; }
    .badge .count { margin-left: 6px; opacity: 0.7; }
    .events { max-height: 600px; overflow-y: auto; overflow-x: auto; }
    table { width: 100%; border-collapse: collapse; min-width: 800px; }
    th { background: #1a1a1a; padding: 12px 16px; text-align: left; font-size: 0.7rem; text-transform: uppercase; color: #666; letter-spacing: 0.5px; position: sticky; top: 0; }
    td { padding: 12px 16px; border-bottom: 1px solid #1a1a1a; font-size: 0.85rem; vertical-align: middle; }
    tr:hover td { background: #1a1a1a; }
    tr.new td { animation: highlight 2s ease-out; }
    @keyframes highlight { from { background: #1e3a5f; } to { background: transparent; } }
    .id { color: #666; font-family: monospace; font-size: 0.8rem; }
    .type { color: #60a5fa; font-weight: 500; }
    .content { color: #c084fc; }
    .status-val { color: #f59e0b; font-weight: 500; font-size: 0.8rem; }
    .instance { color: #888; font-family: monospace; font-size: 0.75rem; max-width: 180px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .phone { color: #22d3ee; font-family: monospace; font-size: 0.8rem; }
    .notif { color: #a78bfa; font-size: 0.8rem; }
    .st { padding: 4px 10px; border-radius: 4px; font-size: 0.75rem; font-weight: 600; display: inline-block; }
    .st.valid { background: #0d3320; color: #22c55e; }
    .st.invalid { background: #3d1515; color: #ef4444; }
    .errors { color: #ef4444; font-size: 0.75rem; max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .time { color: #666; font-size: 0.75rem; white-space: nowrap; }
    .empty { padding: 40px; text-align: center; color: #666; }
    ::-webkit-scrollbar { width: 8px; height: 8px; }
    ::-webkit-scrollbar-track { background: #1a1a1a; }
    ::-webkit-scrollbar-thumb { background: #333; border-radius: 4px; }
    .reset-btn { background: #3d1515; color: #ef4444; border: none; padding: 8px 16px; border-radius: 6px; cursor: pointer; font-size: 0.8rem; }
    .reset-btn:hover { background: #4d1f1f; }
    /* Detail badges and tags */
    .details { display: flex; flex-wrap: wrap; gap: 6px; align-items: center; min-width: 250px; }
    .detail-badge { padding: 4px 10px; border-radius: 4px; font-size: 0.75rem; font-weight: 600; white-space: nowrap; }
    .content-badge { background: #3d1f5c; color: #c084fc; border: 1px solid #5b2d8c; }
    .detail-text { color: #a3e635; font-size: 0.8rem; font-style: italic; max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .detail-media { background: #1e3a5f; color: #60a5fa; padding: 3px 8px; border-radius: 4px; font-size: 0.75rem; }
    .detail-tag { padding: 2px 6px; border-radius: 3px; font-size: 0.65rem; font-weight: 600; text-transform: uppercase; }
    .detail-tag.fromme { background: #0d3320; color: #22c55e; }
    .detail-tag.group { background: #1e3a5f; color: #60a5fa; }
    .detail-tag.notif { background: #3d1f5c; color: #a78bfa; }
    .detail-info { color: #888; font-size: 0.75rem; }
    .detail-none { color: #444; }
    /* Modal */
    .modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.8); display: none; justify-content: center; align-items: center; z-index: 1000; padding: 20px; }
    .modal-overlay.active { display: flex; }
    .modal { background: #141414; border: 1px solid #2a2a2a; border-radius: 12px; max-width: 800px; width: 100%; max-height: 80vh; display: flex; flex-direction: column; }
    .modal-header { padding: 16px 20px; border-bottom: 1px solid #2a2a2a; display: flex; justify-content: space-between; align-items: center; }
    .modal-header h3 { color: #00d9ff; font-size: 1rem; }
    .modal-close { background: none; border: none; color: #666; font-size: 1.5rem; cursor: pointer; padding: 0 8px; }
    .modal-close:hover { color: #fff; }
    .modal-body { padding: 20px; overflow-y: auto; flex: 1; }
    .modal-footer { padding: 16px 20px; border-top: 1px solid #2a2a2a; display: flex; gap: 12px; justify-content: flex-end; }
    .btn { padding: 10px 20px; border-radius: 6px; font-size: 0.85rem; cursor: pointer; border: none; font-weight: 500; }
    .btn-primary { background: #3b82f6; color: white; }
    .btn-primary:hover { background: #2563eb; }
    .btn-success { background: #22c55e; color: white; }
    .btn-success:hover { background: #16a34a; }
    .json-view { background: #0a0a0a; border: 1px solid #2a2a2a; border-radius: 8px; padding: 16px; font-family: 'Monaco', 'Menlo', monospace; font-size: 0.8rem; line-height: 1.6; overflow-x: auto; white-space: pre-wrap; word-break: break-all; }
    .json-key { color: #60a5fa; }
    .json-string { color: #a3e635; }
    .json-number { color: #f59e0b; }
    .json-bool { color: #c084fc; }
    .json-null { color: #6b7280; }
    .copied { position: fixed; top: 20px; right: 20px; background: #22c55e; color: white; padding: 12px 24px; border-radius: 8px; font-weight: 500; z-index: 1001; animation: fadeOut 2s forwards; }
    @keyframes fadeOut { 0%, 70% { opacity: 1; } 100% { opacity: 0; } }
    tr { cursor: pointer; }
    .view-btn { background: #1e3a5f; color: #60a5fa; border: none; padding: 4px 8px; border-radius: 4px; font-size: 0.7rem; cursor: pointer; }
    .view-btn:hover { background: #2563eb; color: white; }
  </style>
</head>
<body>
  <div class="container">
    <header>
      <div style="display: flex; justify-content: space-between; align-items: start;">
        <div>
          <h1>Webhook Monitor</h1>
          <p>WhatsApp API - Z-API Compatible Event Interceptor</p>
        </div>
        <div style="display:flex;gap:12px;align-items:center;">
          <button class="reset-btn" onclick="resetStats()">Reset</button>
          <div id="connection" class="status disconnected">Connecting...</div>
        </div>
      </div>
    </header>
    <div class="grid">
      <div class="card"><h3>Total Events</h3><div class="value total" id="total">0</div></div>
      <div class="card"><h3>Valid</h3><div class="value valid" id="valid">0</div></div>
      <div class="card"><h3>Invalid</h3><div class="value invalid" id="invalid">0</div></div>
      <div class="card"><h3>Success Rate</h3><div class="value rate" id="rate">0%</div></div>
    </div>
    <div class="section">
      <div class="section-header"><h2>Event Types</h2></div>
      <div class="badges" id="types"></div>
    </div>
    <div class="section">
      <div class="section-header"><h2>Content Types</h2></div>
      <div class="badges" id="contents"></div>
    </div>
    <div class="section">
      <div class="section-header"><h2>Recent Events</h2><span id="eventCount" style="color:#666;font-size:0.8rem;">0 events</span></div>
      <div class="events" id="events"><div class="empty">Waiting for webhooks...</div></div>
    </div>
  </div>
  <!-- Modal -->
  <div class="modal-overlay" id="modal" onclick="closeModal(event)">
    <div class="modal" onclick="event.stopPropagation()">
      <div class="modal-header">
        <h3 id="modal-title">Event Payload</h3>
        <button class="modal-close" onclick="closeModal()">&times;</button>
      </div>
      <div class="modal-body">
        <div class="json-view" id="json-content"></div>
      </div>
      <div class="modal-footer">
        <button class="btn btn-primary" onclick="copyPayload()">Copy JSON</button>
        <button class="btn btn-success" onclick="downloadPayload()">Download</button>
      </div>
    </div>
  </div>
  <script>
    let ws, events = [];
    const $ = id => document.getElementById(id);

    function connect() {
      ws = new WebSocket('ws://' + location.host + '/ws');
      ws.onopen = () => { $('connection').className = 'status connected'; $('connection').textContent = 'Live'; };
      ws.onclose = () => { $('connection').className = 'status disconnected'; $('connection').textContent = 'Reconnecting...'; setTimeout(connect, 2000); };
      ws.onmessage = e => {
        const data = JSON.parse(e.data);
        if (data.type === 'init' || data.type === 'stats') updateStats(data);
        if (data.type === 'event') addEvent(data.event);
      };
    }

    function resetStats() {
      fetch('/api/reset', { method: 'POST' }).then(() => {
        events = [];
        renderEvents();
      });
    }

    function updateStats(data) {
      $('total').textContent = data.total || 0;
      $('valid').textContent = data.valid || 0;
      $('invalid').textContent = data.invalid || 0;
      const rate = data.total > 0 ? ((data.valid / data.total) * 100).toFixed(1) : 0;
      $('rate').textContent = rate + '%';

      $('types').innerHTML = (data.byType || []).map(t =>
        '<div class="badge type">' + t.event_type + '<span class="count">' + t.total + '</span></div>'
      ).join('') || '<span style="color:#666;padding:8px;">No events yet</span>';

      $('contents').innerHTML = (data.byContent || []).map(c =>
        '<div class="badge content">' + c.content_type + '<span class="count">' + c.count + '</span></div>'
      ).join('') || '<span style="color:#666;padding:8px;">No content types yet</span>';

      if (data.recent) { events = data.recent; renderEvents(); }
    }

    function addEvent(event) {
      events.unshift(event);
      if (events.length > 100) events.pop();
      renderEvents(true);
      fetch('/api/stats').then(r => r.json()).then(updateStats);
    }

    function getPayload(e) {
      if (!e.payload) return {};
      try { return typeof e.payload === 'string' ? JSON.parse(e.payload) : e.payload; }
      catch { return {}; }
    }

    // Status badges with colors
    const statusColors = {
      // Presence
      COMPOSING: '#f59e0b', RECORDING: '#ef4444', PAUSED: '#6b7280', AVAILABLE: '#22c55e', UNAVAILABLE: '#6b7280',
      // Message
      PENDING: '#6b7280', SENT: '#3b82f6', RECEIVED: '#22c55e', READ: '#8b5cf6', READ_BY_ME: '#a78bfa', PLAYED: '#ec4899', PLAYED_BY_ME: '#f472b6',
      // Connection
      connected: '#22c55e', disconnected: '#ef4444'
    };

    function getDetails(e) {
      const p = getPayload(e);
      const type = e.event_type;
      let html = '';

      // Status badge for presence/message events
      if (p.status) {
        const color = statusColors[p.status] || '#888';
        html += '<span class="detail-badge" style="background:' + color + '20;color:' + color + ';border:1px solid ' + color + '40">' + p.status + '</span>';
      }

      // Content type badge
      if (e.content_type) {
        html += '<span class="detail-badge content-badge">' + e.content_type + '</span>';
      }

      // Event-specific details
      if (type === 'ReceivedCallback') {
        if (p.text?.message) html += '<span class="detail-text">"' + truncate(p.text.message, 30) + '"</span>';
        else if (p.image) html += '<span class="detail-media">Image</span>';
        else if (p.audio) html += '<span class="detail-media">Audio ' + (p.audio.seconds || '') + 's</span>';
        else if (p.video) html += '<span class="detail-media">Video</span>';
        else if (p.document) html += '<span class="detail-media">Doc: ' + (p.document.fileName || 'file') + '</span>';
        else if (p.sticker) html += '<span class="detail-media">Sticker</span>';
        else if (p.location) html += '<span class="detail-media">Location</span>';
        else if (p.contact) html += '<span class="detail-media">Contact: ' + (p.contact.displayName || '') + '</span>';
        else if (p.reaction) html += '<span class="detail-media">Reaction: ' + (p.reaction.value || '') + '</span>';
        if (p.fromMe) html += '<span class="detail-tag fromme">FROM ME</span>';
        if (p.isGroup) html += '<span class="detail-tag group">GROUP</span>';
        if (p.notification) html += '<span class="detail-tag notif">' + p.notification + '</span>';
      }

      if (type === 'MessageStatusCallback') {
        if (p.ids?.length) html += '<span class="detail-info">' + p.ids.length + ' msg(s)</span>';
        if (p.isGroup) html += '<span class="detail-tag group">GROUP</span>';
      }

      if (type === 'PresenceChatCallback') {
        if (p.participant) html += '<span class="detail-info">by ' + p.participant + '</span>';
      }

      if (type === 'ConnectedCallback') {
        html += '<span class="detail-badge" style="background:#22c55e20;color:#22c55e;border:1px solid #22c55e40">CONNECTED</span>';
        if (p.phone) html += '<span class="detail-info">' + p.phone + '</span>';
      }

      if (type === 'DisconnectedCallback') {
        html += '<span class="detail-badge" style="background:#ef444420;color:#ef4444;border:1px solid #ef444440">DISCONNECTED</span>';
        if (p.reason) html += '<span class="detail-info">' + p.reason + '</span>';
      }

      if (type === 'DeliveryCallback') {
        if (p.zaapId) html += '<span class="detail-info">zaapId: ' + truncate(p.zaapId, 12) + '</span>';
      }

      return html || '<span class="detail-none">-</span>';
    }

    function truncate(str, len) {
      if (!str) return '';
      return str.length > len ? str.slice(0, len) + '...' : str;
    }

    function renderEvents(isNew = false) {
      $('eventCount').textContent = events.length + ' events';
      if (events.length === 0) {
        $('events').innerHTML = '<div class="empty">Waiting for webhooks...</div>';
        return;
      }
      $('events').innerHTML = '<table>' +
        '<thead><tr>' +
        '<th>ID</th><th>Event Type</th><th>Details</th><th>Instance</th><th>Phone</th><th>Valid</th><th>Time</th><th></th>' +
        '</tr></thead><tbody>' +
        events.map((e, i) => {
          const errors = e.errors ? (typeof e.errors === 'string' ? e.errors : JSON.stringify(e.errors)) : '';
          const shortInstance = e.instance_id ? e.instance_id.slice(0, 8) + '...' : '-';
          return '<tr class="' + (i === 0 && isNew ? 'new' : '') + '" onclick="showPayload(' + i + ')">' +
            '<td class="id">#' + e.id + '</td>' +
            '<td class="type">' + e.event_type.replace('Callback', '') + '</td>' +
            '<td class="details">' + getDetails(e) + '</td>' +
            '<td class="instance" title="' + (e.instance_id || '') + '">' + shortInstance + '</td>' +
            '<td class="phone">' + (e.phone || '-') + '</td>' +
            '<td><span class="st ' + (e.valid ? 'valid' : 'invalid') + '" title="' + errors.replace(/"/g, '&quot;') + '">' + (e.valid ? 'OK' : 'ERR') + '</span></td>' +
            '<td class="time">' + (e.created_at ? new Date(e.created_at).toLocaleTimeString() : '-') + '</td>' +
            '<td><button class="view-btn" onclick="event.stopPropagation(); showPayload(' + i + ')">View</button></td>' +
            '</tr>';
        }).join('') +
        '</tbody></table>';
    }

    // Modal functions
    let currentPayload = null;
    let currentEventId = null;

    function showPayload(index) {
      const e = events[index];
      if (!e) return;
      currentEventId = e.id;
      const payload = getPayload(e);
      currentPayload = JSON.stringify(payload, null, 2);
      $('modal-title').textContent = 'Event #' + e.id + ' - ' + e.event_type;
      $('json-content').innerHTML = syntaxHighlight(currentPayload);
      $('modal').classList.add('active');
    }

    function closeModal(e) {
      if (e && e.target !== $('modal')) return;
      $('modal').classList.remove('active');
    }

    function copyPayload() {
      navigator.clipboard.writeText(currentPayload).then(() => {
        const toast = document.createElement('div');
        toast.className = 'copied';
        toast.textContent = 'Copied to clipboard!';
        document.body.appendChild(toast);
        setTimeout(() => toast.remove(), 2000);
      });
    }

    function downloadPayload() {
      const blob = new Blob([currentPayload], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'event-' + currentEventId + '.json';
      a.click();
      URL.revokeObjectURL(url);
    }

    function syntaxHighlight(json) {
      return json
        .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
        .replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, (match) => {
          let cls = 'json-number';
          if (/^"/.test(match)) {
            cls = /:$/.test(match) ? 'json-key' : 'json-string';
          } else if (/true|false/.test(match)) {
            cls = 'json-bool';
          } else if (/null/.test(match)) {
            cls = 'json-null';
          }
          return '<span class="' + cls + '">' + match + '</span>';
        });
    }

    // Keyboard shortcut to close modal
    document.addEventListener('keydown', (e) => {
      if (e.key === 'Escape') closeModal();
    });

    connect();
  </script>
</body>
</html>`;

// ═══════════════════════════════════════════════════════════════════════════════
// SERVER - Native Bun.serve with WebSocket
// ═══════════════════════════════════════════════════════════════════════════════

const PORT = Number(Bun.env.PORT) || 3333;

const C = {
	reset: "\x1b[0m",
	green: "\x1b[32m",
	red: "\x1b[31m",
	cyan: "\x1b[36m",
	magenta: "\x1b[35m",
	dim: "\x1b[2m",
	bold: "\x1b[1m",
};

function logEvent(v: ValidationResult, id: number, p: Record<string, unknown>) {
	const valid = v.valid
		? `${C.green}VALID${C.reset}`
		: `${C.red}INVALID${C.reset}`;
	const content = v.contentType
		? `${C.magenta}[${v.contentType}]${C.reset}`
		: "";
	const status = p.status ? `${C.bold}${p.status}${C.reset}` : "";
	const instance = p.instanceId
		? `${C.dim}${String(p.instanceId).slice(0, 8)}...${C.reset}`
		: "";
	console.log(
		`${C.bold}#${id}${C.reset} ${valid} ${C.cyan}${v.eventType}${C.reset} ${content} ${status} ${instance} ${C.dim}${p.phone || ""}${C.reset}`,
	);
	if (v.errors.length > 0)
		console.log(`   ${C.red}${v.errors.join(", ")}${C.reset}`);
}

Bun.serve({
	port: PORT,
	async fetch(req, server) {
		const url = new URL(req.url);

		// WebSocket upgrade
		if (url.pathname === "/ws") {
			if (server.upgrade(req)) return;
			return new Response("WebSocket upgrade failed", { status: 400 });
		}

		// Dashboard
		if (
			req.method === "GET" &&
			(url.pathname === "/" || url.pathname === "/dashboard")
		) {
			return new Response(HTML, {
				headers: { "Content-Type": "text/html" },
			});
		}

		// Stats API
		if (req.method === "GET" && url.pathname === "/api/stats") {
			return Response.json(getStats());
		}

		// Reset API
		if (req.method === "POST" && url.pathname === "/api/reset") {
			db.exec(`DELETE FROM events`);
			db.exec(
				`UPDATE stats SET total = 0, valid = 0, invalid = 0, started_at = datetime('now')`,
			);
			broadcast({ type: "stats", ...getStats() });
			return Response.json({ reset: true });
		}

		// Webhook handler - accepts POST on any path
		if (req.method === "POST") {
			try {
				const payload = (await req.json()) as Record<string, unknown>;
				const validation = validateEvent(payload);
				const id = saveEvent(payload, validation);

				const event = {
					id,
					event_type: validation.eventType,
					content_type: validation.contentType,
					notification: validation.notification,
					instance_id: payload.instanceId,
					phone: payload.phone,
					valid: validation.valid ? 1 : 0,
					errors:
						validation.errors.length > 0
							? JSON.stringify(validation.errors)
							: null,
					payload: JSON.stringify(payload),
					created_at: new Date().toISOString(),
				};

				broadcast({ type: "event", event });
				logEvent(validation, id, payload);

				return Response.json({
					received: true,
					id,
					valid: validation.valid,
					errors: validation.errors,
				});
			} catch (e: any) {
				return Response.json({ error: e.message }, { status: 400 });
			}
		}

		return new Response("Not Found", { status: 404 });
	},
	websocket: {
		open(ws) {
			clients.add(ws);
			ws.send(JSON.stringify({ type: "init", ...getStats() }));
		},
		close(ws) {
			clients.delete(ws);
		},
		message() {},
	},
});

console.log(`
${C.bold}${C.cyan}╔═══════════════════════════════════════════════════════════════╗
║           WEBHOOK MONITOR - WhatsApp API                      ║
╚═══════════════════════════════════════════════════════════════╝${C.reset}

  ${C.green}Dashboard:${C.reset}  http://localhost:${PORT}
  ${C.cyan}Webhook:${C.reset}    POST http://localhost:${PORT}/*
  ${C.dim}Database:${C.reset}   webhooks.db (SQLite)

  ${C.dim}Waiting for webhooks...${C.reset}
`);
