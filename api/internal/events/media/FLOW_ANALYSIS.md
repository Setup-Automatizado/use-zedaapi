# 🔍 Análise Completa: Fluxo de Processamento de Mídia para S3

## 📋 Status da Implementação

### ✅ COMPLETO - Componentes Implementados

1. **MediaDownloader** ([downloader.go:353](downloader.go))
   - Suporte genérico para TODOS os 8 tipos de mídia WhatsApp
   - Auto-detecção via `whatsmeow.GetMediaType()`
   - Extração via reflection (Mimetype, FileName)
   - 100+ MIME types via `mime.ExtensionsByType()`

2. **S3Uploader** ([uploader.go:323](uploader.go))
   - AWS SDK v2 com presigned URLs
   - ACL opcional (modern pattern: bucket policy)
   - Upload multipart (5MB chunks, 100 concurrency)
   - Estrutura organizada: `{instance_id}/{year}/{month}/{day}/{event_id}.{ext}`

3. **MediaProcessor** ([processor.go:285](processor.go))
   - Pipeline completo: Download → Upload → Update metadata
   - Retry automático com exponential backoff
   - Classificação de erros (retryable vs fatal)

4. **MediaWorker** ([worker.go:301](worker.go))
   - Per-instance processing com polling loop
   - Message reconstruction from event_outbox
   - Distributed locking via PostgreSQL

5. **MediaCoordinator** ([coordinator.go:230](coordinator.go))
   - Worker pool management
   - Graceful shutdown com timeout
   - Thread-safe registration/unregistration

---

## 🔄 Fluxo Completo: Do Evento WhatsApp até URL Pública no Webhook

### **Fase 1: Captura de Evento com Mídia**

```
WhatsApp Client (whatsmeow)
    ↓ [*events.Message com ImageMessage/VideoMessage/etc]

EventHandler.Handle()
    ├─→ Extrai metadata de mídia (MediaKey, DirectPath, SHA256, MimeType, etc)
    ├─→ Cria InternalEvent com HasMedia=true
    └─→ EventBuffer.Add(event)

EventWriter.ProcessBatch()
    ├─→ Transform: NoOpTransformer serializa RawPayload → JSON
    ├─→ Persiste em event_outbox:
    │   ├─ payload: JSON serializado do evento original (events.Message)
    │   ├─ has_media: true
    │   ├─ media_processed: false
    │   ├─ media_url: NULL
    │   └─ media_error: NULL
    └─→ Persiste em media_metadata:
        ├─ event_id: UUID do evento
        ├─ instance_id: UUID da instância
        ├─ media_key: Chave de criptografia WhatsApp
        ├─ direct_path: Caminho no servidor WhatsApp
        ├─ file_sha256: Hash do arquivo
        ├─ media_type: image/video/audio/document/sticker
        ├─ mime_type: image/jpeg, video/mp4, etc
        ├─ download_status: 'pending'
        └─ download_attempts: 0
```

**Arquivo**: [capture/writer.go](../capture/writer.go)

**Tabelas**:
- `event_outbox`: Evento original + metadata
- `media_metadata`: Informações para download/upload

---

### **Fase 2: Processamento Assíncrono de Mídia**

```
MediaWorker.run() [loop a cada MEDIA_POLL_INTERVAL]
    ↓
┌─→ 1. MediaRepo.PollPendingDownloads(limit=batch_size)
│       └─→ SELECT * FROM media_metadata
│           WHERE download_status IN ('pending', 'failed')
│             AND download_attempts < max_retries
│             AND processing_worker_id IS NULL
│           ORDER BY created_at ASC
│           LIMIT $1
│
├─→ 2. MediaRepo.AcquireForProcessing(event_id, worker_id)
│       └─→ UPDATE media_metadata
│           SET processing_worker_id = $worker_id,
│               processing_started_at = NOW(),
│               download_status = 'downloading'
│           WHERE event_id = $event_id
│             AND processing_worker_id IS NULL
│           [Distributed lock via PostgreSQL]
│
├─→ 3. MediaWorker.reconstructMessage(media_metadata)
│       ├─→ OutboxRepo.GetEventByID(event_id)
│       │   └─→ SELECT payload FROM event_outbox WHERE event_id = $1
│       ├─→ json.Unmarshal(payload) → map[string]interface{}
│       ├─→ Extrai campo "Message" ou "message"
│       ├─→ Re-marshal para JSON
│       └─→ json.Unmarshal → proto message específico:
│           ├─ MediaTypeImage → waE2E.ImageMessage
│           ├─ MediaTypeVideo → waE2E.VideoMessage
│           ├─ MediaTypeAudio/Voice → waE2E.AudioMessage
│           ├─ MediaTypeDocument → waE2E.DocumentMessage
│           └─ MediaTypeSticker → waE2E.StickerMessage
│
├─→ 4. MediaProcessor.ProcessWithRetry(client, msg)
│       │
│       ├─→ 4.1 MediaDownloader.Download(client, msg)
│       │       ├─→ extractMediaInfoGeneric(msg)
│       │       │   ├─→ whatsmeow.GetMediaType(msg) → MediaType
│       │       │   ├─→ extractContentTypeReflection(msg) → MimeType
│       │       │   └─→ extractFileNameReflection(msg) → FileName
│       │       ├─→ client.Download(downloadable, nil)
│       │       │   └─→ WhatsApp CDN download + decrypt
│       │       ├─→ Metrics: MediaDownloadsTotal, MediaDownloadDuration
│       │       └─→ Return: DownloadResult{Data, ContentType, FileName, FileSize, SHA256}
│       │
│       ├─→ 4.2 MediaRepo.UpdateDownloadStatus(event_id, 'downloaded')
│       │
│       ├─→ 4.3 S3Uploader.Upload(data, metadata)
│       │       ├─→ Generate S3 key: {instance_id}/{year}/{month}/{day}/{event_id}.{ext}
│       │       ├─→ Build PutObjectInput:
│       │       │   ├─ Bucket: from config (S3_BUCKET)
│       │       │   ├─ Key: generated key
│       │       │   ├─ Body: io.Reader from download
│       │       │   ├─ ContentType: detected MIME
│       │       │   ├─ Metadata: instance_id, event_id, media_type
│       │       │   └─ ACL: optional (S3_ACL env var)
│       │       ├─→ manager.Uploader.Upload(ctx, input)
│       │       │   └─→ Multipart upload (5MB chunks, 100 concurrency)
│       │       ├─→ S3.PresignClient.PresignGetObject(key, expiry)
│       │       │   └─→ Presigned URL válida por S3_URL_EXPIRY (default: 7 dias)
│       │       ├─→ Metrics: MediaUploadAttempts, MediaUploadDuration, MediaUploadSizeBytes
│       │       └─→ Return: (s3_key, presigned_url)
│       │
│       ├─→ 4.4 MediaRepo.UpdateUploadInfo(event_id, bucket, key, url, expiry)
│       │       └─→ UPDATE media_metadata
│       │           SET s3_bucket = $bucket,
│       │               s3_key = $key,
│       │               s3_url = $presigned_url,
│       │               url_expires_at = NOW() + INTERVAL '30 days'
│       │
│       └─→ 4.5 MediaRepo.MarkComplete(event_id)
│               └─→ UPDATE media_metadata
│                   SET download_status = 'completed',
│                       completed_at = NOW()
│
└─→ 5. OutboxRepo.UpdateMediaInfo(event_id, media_url, NULL, true)
        └─→ UPDATE event_outbox
            SET media_url = $presigned_url,
                media_processed = true,
                media_error = NULL
            WHERE event_id = $event_id
```

**Arquivos**:
- [worker.go](worker.go) - Orchestration
- [processor.go](processor.go) - Pipeline
- [downloader.go](downloader.go) - WhatsApp download
- [uploader.go](uploader.go) - S3 upload

---

### **Fase 3: Webhook Delivery com URL Pública**

```
DispatchCoordinator → InstanceWorker → EventProcessor
    ↓
OutboxRepo.PollPendingEvents(instance_id)
    └─→ SELECT * FROM event_outbox
        WHERE instance_id = $1
          AND status IN ('pending', 'retrying')
          AND media_processed = true  [✅ Apenas eventos com mídia processada]
        ORDER BY sequence_number ASC

EventProcessor.Process(event)
    ├─→ Transform event:
    │   ├─→ TargetTransformer.Transform(event)
    │   │   └─→ ZAPITransformer injects media_url from event.Metadata:
    │   │       ├─ ImageMessage: callback.Image.ImageURL = event.Metadata["media_url"]
    │   │       ├─ VideoMessage: callback.Video.VideoURL = event.Metadata["media_url"]
    │   │       ├─ AudioMessage: callback.Audio.AudioURL = event.Metadata["media_url"]
    │   │       ├─ DocumentMessage: callback.Document.DocumentURL = event.Metadata["media_url"]
    │   │       └─ StickerMessage: callback.Sticker.StickerURL = event.Metadata["media_url"]
    │   └─→ Webhook payload now contains public S3 URL!
    │
    ├─→ Deliver via HTTPTransport.Deliver(webhook_url, payload)
    │   └─→ POST {webhook_url}
    │       Content-Type: application/json
    │       Body: {
    │         "event": "message",
    │         "instanceId": "...",
    │         "data": {
    │           "image": {
    │             "imageUrl": "https://s3.amazonaws.com/bucket/instance/.../event.jpg?X-Amz-...",
    │             "caption": "...",
    │             "mimeType": "image/jpeg"
    │           }
    │         }
    │       }
    │
    └─→ OutboxRepo.MarkDelivered(event_id)
        └─→ UPDATE event_outbox
            SET status = 'delivered',
                delivered_at = NOW()
```

**Arquivos**:
- [dispatch/processor.go](../dispatch/processor.go)
- [transform/zapi/transformer.go](../transform/zapi/transformer.go)
- [transport/http.go](../transport/http.go)

---

## ⚠️ Tratamento de Falhas: O Que Acontece Quando o S3 Falha?

### **Cenário 1: Falha no Download do WhatsApp**

```
MediaDownloader.Download() → ERROR
    ↓
classifyDownloadError(err) determina tipo:
    ├─ Retryable Errors:
    │   ├─ timeout → Retry
    │   ├─ connection → Retry
    │   ├─ network → Retry
    │   └─ media_conn_refresh_failed → Retry
    └─ Non-Retryable Errors:
        ├─ not_logged_in → FAIL PERMANENTE
        ├─ no_url → FAIL PERMANENTE
        ├─ http_403/404/410 → FAIL PERMANENTE
        └─ invalid_hmac/hash → FAIL PERMANENTE

MediaProcessor.ProcessWithRetry():
    ├─ If retryable: Exponential backoff (2s, 4s, 8s)
    ├─ Max retries: MEDIA_MAX_RETRIES (default: 3)
    └─ If max exceeded:
        ├─→ MediaRepo.UpdateDownloadStatus(event_id, 'failed', error_msg)
        ├─→ OutboxRepo.UpdateMediaInfo(event_id, NULL, error_msg, false)
        │   └─→ event_outbox.media_error = "download failed: timeout"
        │       event_outbox.media_processed = false
        └─→ Event PERMANECE no outbox mas NÃO é entregue (media_processed=false)
```

**Resultado**: Webhook **NÃO é enviado** até mídia ser processada com sucesso.

---

### **Cenário 2: Falha no Upload para S3**

```
S3Uploader.Upload() → ERROR
    ↓
classifyS3Error(err) determina tipo:
    ├─ timeout → Retryable
    ├─ connection → Retryable
    ├─ network → Retryable
    ├─ access_denied → Non-Retryable (credencial inválida)
    ├─ bucket_not_found → Non-Retryable (bucket não existe)
    └─ file_too_large → Non-Retryable (excede limite)

MediaProcessor.ProcessWithRetry():
    ├─ If retryable: Exponential backoff (2s, 4s, 8s)
    ├─ Max retries: MEDIA_MAX_RETRIES (default: 3)
    └─ If max exceeded:
        ├─→ Metrics: MediaFailures.WithLabels(instance_id, media_type, "upload").Inc()
        ├─→ OutboxRepo.UpdateMediaInfo(event_id, NULL, error_msg, false)
        │   └─→ event_outbox.media_error = "upload failed: access_denied"
        │       event_outbox.media_processed = false
        └─→ Event PERMANECE no outbox mas NÃO é entregue
```

**Resultado**: Webhook **NÃO é enviado** até S3 upload ter sucesso.

---

### **Cenário 3: Falha na Geração de Presigned URL**

```
S3Uploader.GeneratePresignedURL() → ERROR
    ↓
MediaProcessor.Process():
    └─→ Return error: "failed to generate presigned URL: ..."

MediaProcessor.ProcessWithRetry():
    ├─ Retry (presigned URL é idempotent)
    └─ If max exceeded:
        ├─→ OutboxRepo.UpdateMediaInfo(event_id, NULL, error_msg, false)
        └─→ Event NÃO é entregue
```

**Resultado**: Webhook **NÃO é enviado** sem URL válida.

---

### **Cenário 4: S3 Upload Sucesso mas UpdateMediaInfo Falha**

```
S3Uploader.Upload() → SUCCESS (file uploaded to S3)
    ↓
MediaRepo.UpdateUploadInfo() → ERROR (PostgreSQL failure)
    ↓
MediaProcessor.Process() → Return error: "failed to update upload info: ..."
    ↓
MediaProcessor.ProcessWithRetry():
    └─ Retry ENTIRE process:
        ├─→ Re-download from WhatsApp (idempotent)
        ├─→ Re-upload to S3 (overwrites existing file)
        └─→ Retry UpdateUploadInfo()
```

**Resultado**: File pode ser duplicado no S3, mas metadata será eventualmente consistente.

**Otimização Possível**: Usar `S3Uploader.ObjectExists(key)` antes de re-upload.

---

### **Cenário 5: Timeout no Processamento**

```
MediaProcessor.Process(ctx) com timeout configurável
    ↓
Context deadline exceeded:
    ├─ MediaDownloadTimeout: 30s (default)
    └─ MediaUploadTimeout: 60s (default)

If timeout:
    ├─→ ctx.Done() detected in loop
    ├─→ Return context.DeadlineExceeded
    ├─→ classifyError() → retryable=true
    └─→ ProcessWithRetry() schedules retry
```

**Resultado**: Retry com timeout incrementado (exponential backoff).

---

### **Cenário 6: Worker Crash Durante Processamento**

```
MediaWorker crashes while processing media
    ↓
PostgreSQL distributed lock prevents duplicate processing:
    └─→ media_metadata.processing_worker_id = crashed_worker_id
        media_metadata.processing_started_at = timestamp

Other workers CANNOT acquire lock (processing_worker_id IS NOT NULL)
    ↓
Manual intervention OR automated cleanup job:
    └─→ Release stuck locks after timeout (e.g., 10 minutes):
        UPDATE media_metadata
        SET processing_worker_id = NULL,
            processing_started_at = NULL
        WHERE processing_worker_id IS NOT NULL
          AND processing_started_at < NOW() - INTERVAL '10 minutes'
```

**Resultado**: Mídia será reprocessada após cleanup job liberar lock.

**TODO**: Implementar cleanup job em Phase 7 (Background Jobs).

---

## 📊 Estados Finais Possíveis

### ✅ **Sucesso Completo**

```sql
-- media_metadata:
download_status = 'completed'
s3_url = 'https://s3.amazonaws.com/bucket/instance/.../event.jpg?X-Amz-...'
completed_at = NOW()

-- event_outbox:
media_processed = true
media_url = 'https://s3.amazonaws.com/bucket/instance/.../event.jpg?X-Amz-...'
media_error = NULL
status = 'delivered'
delivered_at = NOW()
```

**Webhook enviado** com URL pública válida por 30 dias.

---

### ❌ **Falha Permanente (Download)**

```sql
-- media_metadata:
download_status = 'failed'
download_error = 'not_logged_in' (ou outro erro fatal)
download_attempts = 3 (max_retries)

-- event_outbox:
media_processed = false
media_url = NULL
media_error = 'download failed: not_logged_in'
status = 'pending' (nunca será delivered)
```

**Webhook NÃO será enviado**. Requer intervenção manual ou reconnect do cliente.

---

### ❌ **Falha Permanente (Upload)**

```sql
-- media_metadata:
download_status = 'downloaded' (download OK)
download_attempts = 3
upload_error = 'access_denied' (S3 credentials inválidas)

-- event_outbox:
media_processed = false
media_url = NULL
media_error = 'upload failed: access_denied'
status = 'pending'
```

**Webhook NÃO será enviado**. Requer correção de credenciais S3.

---

### ⏸️ **Processamento Pendente**

```sql
-- media_metadata:
download_status = 'pending'
download_attempts = 0
processing_worker_id = NULL

-- event_outbox:
media_processed = false
media_url = NULL
media_error = NULL
status = 'pending'
```

**Aguardando** MediaWorker processar.

---

## 🔍 Gaps Identificados e Soluções

### ❗ Gap 1: Falta integração entre MediaCoordinator e ClientRegistry

**Problema**: Workers não são auto-registrados quando cliente conecta.

**Solução (Phase 6.9)**:
```go
// ClientRegistry.wrapEventHandler()
case *events.Connected:
    mediaCoordinator.RegisterInstance(instanceID, client)

case *events.LoggedOut:
    mediaCoordinator.UnregisterInstance(instanceID)
```

---

### ❗ Gap 2: Cleanup de locks travados

**Problema**: Worker crashes deixam locks permanentes.

**Solução (Phase 7)**:
```go
// Background job: cleanup_stuck_locks.go
func (j *CleanupJob) releaseStuckLocks(ctx context.Context) {
    query := `
        UPDATE media_metadata
        SET processing_worker_id = NULL,
            processing_started_at = NULL
        WHERE processing_worker_id IS NOT NULL
          AND processing_started_at < NOW() - INTERVAL '10 minutes'`

    result, _ := j.pool.Exec(ctx, query)
    log.Info("released stuck locks", "count", result.RowsAffected())
}
```

---

### ❗ Gap 3: Retry de eventos com mídia falhada

**Problema**: Eventos com `media_processed=false` ficam presos no outbox.

**Solução (Phase 7)**:
```go
// Background job: retry_failed_media.go
func (j *RetryJob) retryFailedMedia(ctx context.Context) {
    // 1. Busca media_metadata com status=failed
    medias := mediaRepo.GetFailedMedia(ctx, limit)

    // 2. Para cada mídia:
    for _, media := range medias {
        // 2.1 Verifica se erro é retryable
        if !isRetryableError(media.DownloadError) {
            continue // Skip erros permanentes
        }

        // 2.2 Reseta status para pending
        mediaRepo.UpdateDownloadStatus(ctx, media.EventID, 'pending', 0, nil, nil)

        // Worker vai reprocessar automaticamente
    }
}
```

---

### ❗ Gap 4: Expiração de Presigned URLs

**Problema**: Presigned URLs expiram após 30 dias, mas eventos permanecem no sistema.

**Solução (Phase 7)**:
```go
// Background job: refresh_expired_urls.go
func (j *RefreshJob) refreshExpiredURLs(ctx context.Context) {
    // 1. Busca media_metadata com url_expires_at < NOW() + 1 day
    medias := mediaRepo.GetExpiringMedia(ctx)

    // 2. Para cada mídia:
    for _, media := range medias {
        // 2.1 Re-gera presigned URL (arquivo já existe no S3)
        newURL, _ := s3Uploader.GeneratePresignedURL(ctx, media.S3Key)

        // 2.2 Atualiza media_metadata e event_outbox
        newExpiry := time.Now().Add(30 * 24 * time.Hour)
        mediaRepo.UpdateUploadInfo(ctx, media.EventID, media.S3Bucket, media.S3Key, newURL, persistence.S3URLPresigned, &newExpiry)
        outboxRepo.UpdateMediaInfo(ctx, media.EventID, &newURL, nil, true)
    }
}
```

---

## 🎯 Conclusão

### ✅ **O Sistema ESTÁ Completo Para**:

1. ✅ Download de mídia do WhatsApp (todos os 8 tipos)
2. ✅ Upload para S3 com presigned URLs
3. ✅ Injeção de URLs públicas nos webhooks
4. ✅ Retry automático com exponential backoff
5. ✅ Distributed locking para prevenir duplicação
6. ✅ Métricas Prometheus em todas as etapas
7. ✅ Structured logging com contextual fields

### ⚠️ **Comportamento em Caso de Falha S3**:

- **Erros Temporários**: Retry automático (3x) com backoff exponencial
- **Erros Permanentes**: Evento **NÃO é entregue** no webhook
- **Crash de Worker**: Lock liberado após timeout (manual ou background job)
- **URL Expirada**: Precisa de background job para refresh (Phase 7)

### 🚀 **Próximos Passos**:

1. **Phase 6.9**: Integrar MediaCoordinator com ClientRegistry
2. **Phase 6.10**: Testing com todos os 8 media types
3. **Phase 7**: Background Jobs (cleanup, retry, refresh)
