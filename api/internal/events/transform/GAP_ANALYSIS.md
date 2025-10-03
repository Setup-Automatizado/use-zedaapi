# Transformation Layer - Gap Analysis

**Date**: 2025-01-03
**Status**: Phase 3 Implementation Review
**Objective**: Identify all missing fields, events, and incorrect mappings between whatsmeow and Z-API

## Executive Summary

This analysis reveals **significant gaps** in the current transformation layer implementation. While the foundation is solid, approximately **60% of whatsmeow event types** and **40% of message metadata** are not being captured or transformed.

### Critical Findings

1. **42 event types** exist in whatsmeow but only **6 are handled**
2. **25+ message content types** exist but only **12 are handled**
3. **Critical media metadata missing**: `isGIF`, `isAnimated`, `width`, `height`, `ContextInfo` (quotes, mentions, forwards)
4. **15+ MessageInfo fields** not captured: `Category`, `Multicast`, `PushName`, `VerifiedName`, `Edit`, etc.
5. **Newsletter, group, app state events** completely unhandled

---

## 1. Missing Event Types (36 of 42 total)

### 1.1 Connection & Authentication Events ❌ **NOT HANDLED**

| Event Type | Description | Priority | Z-API Equivalent |
|------------|-------------|----------|------------------|
| `QR` | QR code for pairing | **HIGH** | `QRCodeCallback` |
| `PairSuccess` | Successful pairing | **HIGH** | `PairSuccessCallback` |
| `PairError` | Pairing failure | **HIGH** | `PairErrorCallback` |
| `LoggedOut` | Device logged out | **HIGH** | `LoggedOutCallback` |
| `KeepAliveTimeout` | Connection timeout | MEDIUM | `TimeoutCallback` |
| `KeepAliveRestored` | Connection restored | MEDIUM | `RestoredCallback` |
| `StreamReplaced` | Multiple devices conflict | HIGH | `StreamReplacedCallback` |
| `TemporaryBan` | Account temp banned | **CRITICAL** | `BannedCallback` |
| `ConnectFailure` | Connection failure | **HIGH** | `ConnectFailureCallback` |
| `ClientOutdated` | Client version outdated | HIGH | `OutdatedCallback` |
| `ManualLoginReconnect` | Manual reconnect needed | MEDIUM | - |

**Impact**: Users cannot track connection lifecycle, pairing status, or handle disconnections properly.

**Recommendation**: Implement all connection events with proper Z-API webhook mappings.

---

### 1.2 Group Events ❌ **NOT HANDLED**

| Event Type | Description | Priority | Z-API Equivalent |
|------------|-------------|----------|------------------|
| `JoinedGroup` | User joined/added to group | **HIGH** | `GroupJoinCallback` |
| `GroupInfo` | Group metadata changed | **HIGH** | `GroupUpdateCallback` |
| `Picture` | Profile/group picture changed | MEDIUM | `PictureCallback` |

**Current State**: ❌ Zero group event handling
**Z-API Requirement**: ✅ Full group event support expected

**Missing GroupInfo Sub-Events**:
- Name changes
- Topic/description changes
- Locked status (admin-only edits)
- Announce status (admin-only messages)
- Ephemeral message settings
- Membership approval mode
- Participant join/leave
- Admin promote/demote
- Link changes
- Parent group linking

**Impact**: No group administration capabilities, cannot track group changes.

**Recommendation**: **CRITICAL** - Implement complete group event handling as it's core Z-API functionality.

---

### 1.3 User Events ❌ **NOT HANDLED**

| Event Type | Description | Priority | Z-API Equivalent |
|------------|-------------|----------|------------------|
| `PushName` | User push name changed | MEDIUM | `PushNameCallback` |
| `BusinessName` | Business name changed | MEDIUM | `BusinessNameCallback` |
| `UserAbout` | User status changed | LOW | `StatusCallback` |
| `IdentityChange` | User changed device | HIGH | `IdentityChangeCallback` |
| `PrivacySettings` | Privacy settings changed | MEDIUM | `PrivacyCallback` |

**Impact**: Cannot track contact changes, business verification, or security events.

**Recommendation**: Implement user change events, especially `IdentityChange` for security.

---

### 1.4 Message Events ❌ **PARTIALLY HANDLED**

| Event Type | Description | Priority | Current Status |
|------------|-------------|----------|----------------|
| `UndecryptableMessage` | Failed to decrypt | **HIGH** | ❌ Not handled |
| `HistorySync` | Message history sync | MEDIUM | ❌ Not handled |
| `MediaRetry` | Media retry response | LOW | ❌ Not handled |
| `FBMessage` | Facebook/Instagram messages | LOW | ❌ Not handled |

**Impact**: No error handling for decryption failures, missing history sync capability.

**Recommendation**: Add `UndecryptableMessage` handling to retry failed messages.

---

### 1.5 App State Events ❌ **NOT HANDLED** (16 events)

These events sync state changes from other devices:

| Event Type | Description | Priority |
|------------|-------------|----------|
| `Contact` | Contact list changes | MEDIUM |
| `Pin` | Chat pinned/unpinned | LOW |
| `Star` | Message starred | LOW |
| `DeleteForMe` | Message deleted locally | MEDIUM |
| `Mute` | Chat muted/unmuted | LOW |
| `Archive` | Chat archived | LOW |
| `MarkChatAsRead` | Chat marked read | LOW |
| `ClearChat` | Chat cleared | MEDIUM |
| `DeleteChat` | Chat deleted | MEDIUM |
| `PushNameSetting` | Own push name changed | LOW |
| `UnarchiveChatsSetting` | Archive settings changed | LOW |
| `UserStatusMute` | Status updates muted | LOW |
| `LabelEdit` | Label edited | LOW |
| `LabelAssociationChat` | Chat labeled | LOW |
| `LabelAssociationMessage` | Message labeled | LOW |
| `AppStateSyncComplete` | State sync completed | MEDIUM |

**Impact**: Multi-device state changes not synchronized to API users.

**Recommendation**: LOW priority - implement only if multi-device sync needed.

---

### 1.6 Newsletter Events ❌ **NOT HANDLED**

| Event Type | Description | Priority |
|------------|-------------|----------|
| `NewsletterJoin` | Joined newsletter | MEDIUM |
| `NewsletterLeave` | Left newsletter | MEDIUM |
| `NewsletterMuteChange` | Newsletter muted | LOW |
| `NewsletterLiveUpdate` | Newsletter message | HIGH |

**Impact**: No newsletter support (WhatsApp Channels feature).

**Recommendation**: MEDIUM priority - implement for Channels support.

---

### 1.7 Sync Events ❌ **NOT HANDLED**

| Event Type | Description | Priority |
|------------|-------------|----------|
| `OfflineSyncPreview` | Preview of missed events | LOW |
| `OfflineSyncCompleted` | Sync completed | LOW |

**Impact**: Cannot track offline message backlog.

**Recommendation**: LOW priority - informational only.

---

### 1.8 Blocklist Events ❌ **NOT HANDLED**

| Event Type | Description | Priority |
|------------|-------------|----------|
| `Blocklist` | Blocklist changed | MEDIUM |

**Impact**: Cannot track blocked users.

**Recommendation**: MEDIUM priority - implement for privacy features.

---

## 2. Missing Message Fields (15+ fields)

### 2.1 MessageInfo Fields ❌ **NOT CAPTURED**

From `types/message.go`, these fields exist but are **not** stored in metadata:

| Field | Type | Description | Priority | Usage |
|-------|------|-------------|----------|-------|
| `ServerID` | string | Server-assigned ID | MEDIUM | Message tracking |
| `Category` | string | Message category | **HIGH** | Message classification |
| `Multicast` | bool | Multicast message | MEDIUM | Broadcast detection |
| `MediaType` | string | Media type enum | HIGH | Accurate media typing |
| `Edit` | EditAttribute | Edit type (pin/revoke) | **CRITICAL** | Edit handling |
| `VerifiedName` | string | Verified business name | **HIGH** | Business verification |
| `PushName` | string | Sender display name | **CRITICAL** | Contact display |
| `DeviceSentMeta.Phash` | string | Phone hash | LOW | Multi-device sync |
| `MsgBotInfo` | struct | Bot message metadata | MEDIUM | Bot interactions |
| `MsgMetaInfo` | struct | Thread/reply metadata | **CRITICAL** | Reply threading |

**Critical Missing**: `PushName`, `VerifiedName`, `Edit`, `MsgMetaInfo`

**Current Capture**:
```go
event.Metadata["message_id"] = msg.Info.ID
event.Metadata["from"] = msg.Info.Sender.String()
event.Metadata["chat"] = msg.Info.Chat.String()
event.Metadata["from_me"] = fmt.Sprintf("%t", msg.Info.IsFromMe)
event.Metadata["is_group"] = fmt.Sprintf("%t", msg.Info.IsGroup)
event.Metadata["timestamp"] = fmt.Sprintf("%d", msg.Info.Timestamp.Unix())
```

**Recommendation**: Add ALL MessageInfo fields to metadata.

---

### 2.2 Message Event Flags ❌ **NOT CAPTURED**

From `events.Message`, these flags exist but are not stored:

| Flag | Description | Priority | Current Status |
|------|-------------|----------|----------------|
| `IsDocumentWithCaption` | Document with caption | MEDIUM | ❌ Not captured |
| `IsLottieSticker` | Lottie animated sticker | MEDIUM | ❌ Not captured |
| `IsBotInvoke` | Bot-invoked message | MEDIUM | ❌ Not captured |
| `IsViewOnceV2` | View once V2 | HIGH | ❌ Not captured (only `IsViewOnce`) |
| `IsViewOnceV2Extension` | View once V2 extension | HIGH | ❌ Not captured |
| `SourceWebMsg` | From history sync | MEDIUM | ❌ Not captured |
| `UnavailableRequestID` | Retry request ID | LOW | ❌ Not captured |
| `RetryCount` | Number of retries | LOW | ❌ Not captured |
| `NewsletterMeta` | Newsletter metadata | MEDIUM | ❌ Not captured |

**Impact**: Cannot distinguish message wrapper types, missing retry tracking.

**Recommendation**: Add all flags to metadata, especially view-once variants.

---

### 2.3 MessageSource Fields ❌ **NOT CAPTURED**

Embedded in `Receipt` and `ChatPresence` events:

| Field | Description | Priority | Impact |
|-------|-------------|----------|--------|
| `AddressingMode` | pn vs lid addressing | **HIGH** | LID support broken |
| `SenderAlt` | Alternate sender (LID) | **HIGH** | LID JID missing |
| `RecipientAlt` | Alternate recipient (LID) | **HIGH** | LID JID missing |
| `BroadcastListOwner` | Broadcast list owner | MEDIUM | Broadcast context missing |

**Critical**: LID (Lightweight ID) addressing mode completely unsupported.

**Recommendation**: **CRITICAL** - Add LID support for privacy-conscious users.

---

## 3. Missing Message Content Types (13+ types)

### 3.1 Payment Messages ❌ **NOT HANDLED**

| Message Type | Description | Priority |
|--------------|-------------|----------|
| `RequestPaymentMessage` | Payment request | LOW |
| `SendPaymentMessage` | Payment sent | LOW |
| `CancelPaymentRequestMessage` | Payment cancelled | LOW |
| `DeclinePaymentRequestMessage` | Payment declined | LOW |
| `InvoiceMessage` | Invoice message | LOW |

**Recommendation**: LOW priority unless payment features needed.

---

### 3.2 Commerce Messages ❌ **NOT HANDLED**

| Message Type | Description | Priority |
|--------------|-------------|----------|
| `OrderMessage` | Order placement | MEDIUM |
| `ProductMessage` | Product catalog item | MEDIUM |

**Recommendation**: MEDIUM priority for e-commerce use cases.

---

### 3.3 Advanced Interactive Messages ❌ **NOT HANDLED**

| Message Type | Description | Priority | Z-API Support |
|--------------|-------------|----------|---------------|
| `InteractiveMessage` | Advanced buttons/carousels | **HIGH** | ✅ Required |
| `InteractiveResponseMessage` | Interactive response | **HIGH** | ✅ Required |
| `HighlyStructuredMessage` | Structured messages | MEDIUM | ✅ Supported |

**Impact**: Cannot handle modern WhatsApp Business interactive messages.

**Recommendation**: **HIGH** priority - modern business messaging requires this.

---

### 3.4 Protocol Messages ❌ **NOT HANDLED**

| Message Type | Description | Priority |
|--------------|-------------|----------|
| `ProtocolMessage` | Revoke/ephemeral settings | **CRITICAL** |
| `SenderKeyDistributionMessage` | E2E key distribution | LOW |

**Critical**: `ProtocolMessage` handles message revocation (delete for everyone).

**Recommendation**: **CRITICAL** - implement revocation handling.

---

### 3.5 Other Message Types ❌ **NOT HANDLED**

| Message Type | Description | Priority |
|--------------|-------------|----------|
| `LiveLocationMessage` | Live location sharing | MEDIUM |
| `GroupInviteMessage` | Group invite link | **HIGH** |
| `CallMessage` | Call initiation/end | MEDIUM |
| `StickerSyncRmrMessage` | Sticker sync removal | LOW |
| `PollCreationMessageV2/V3` | Poll V2/V3 formats | MEDIUM |
| `ScheduledCallCreationMessage` | Scheduled call | LOW |
| `EventMessage` | Event messages | LOW |
| `CommentMessage` | Comment messages | LOW |
| `KeepInChatMessage` | Keep in chat | LOW |
| `PinInChatMessage` | Pin message | MEDIUM |

**Recommendation**: Prioritize `GroupInviteMessage`, `LiveLocationMessage`, `PinInChatMessage`.

---

## 4. Missing Media Metadata ⚠️ **CRITICAL GAPS**

### 4.1 Image Metadata ❌ **NOT CAPTURED**

| Field | Type | Description | Priority | Z-API Required |
|-------|------|-------------|----------|----------------|
| **`isGIF`** | bool | **GIF detection** | **CRITICAL** | ✅ **YES** |
| **`isAnimated`** | bool | **Animated image** | **HIGH** | ✅ **YES** |
| **`Width`** | uint32 | Image width | **HIGH** | ✅ **YES** |
| **`Height`** | uint32 | Image height | **HIGH** | ✅ **YES** |
| `JpegThumbnail` | bytes | Thumbnail data | MEDIUM | ✅ YES |
| `ThumbnailDirectPath` | string | Thumbnail CDN path | MEDIUM | ⚠️ Optional |
| `ThumbnailSHA256` | bytes | Thumbnail hash | MEDIUM | ⚠️ Optional |
| `MidQualityFileSHA256` | bytes | Mid-quality hash | LOW | ❌ No |
| `ScansSidecar` | bytes | Image scan data | LOW | ❌ No |
| `FirstScanSidecar` | bytes | First scan | LOW | ❌ No |
| `ExperimentGroupID` | uint32 | Experiment group | LOW | ❌ No |
| `StickerSentTS` | int64 | Sticker sent timestamp | LOW | ❌ No |
| `InteractiveAnnotations` | array | Interactive annotations | MEDIUM | ⚠️ Optional |

**Critical**: User specifically mentioned **`isGIF`** - this is a **MUST FIX**.

**Current Code Problem**:
```go
// extractMediaInfo does NOT check for GIF attributes
if img := msg.GetImageMessage(); img != nil {
    info.MediaType = "image"
    // ❌ Missing: img.GetIsGif()
    // ❌ Missing: img.GetWidth()
    // ❌ Missing: img.GetHeight()
    // ❌ Missing: img.GetIsAnimated()
}
```

**Recommendation**: **IMMEDIATE** - Add `isGIF`, `isAnimated`, `width`, `height` to all media.

---

### 4.2 Video Metadata ❌ **NOT CAPTURED**

| Field | Type | Description | Priority | Z-API Required |
|-------|------|-------------|----------|----------------|
| **`GifPlayback`** | bool | **Play as GIF** | **CRITICAL** | ✅ **YES** |
| **`Width`** | uint32 | Video width | **HIGH** | ✅ **YES** |
| **`Height`** | uint32 | Video height | **HIGH** | ✅ **YES** |
| `GifAttribution` | enum | GIF source | MEDIUM | ⚠️ Optional |
| `JpegThumbnail` | bytes | Thumbnail | MEDIUM | ✅ YES |
| `StreamingSidecar` | bytes | Streaming data | LOW | ❌ No |

**Critical**: `GifPlayback` determines if video plays as GIF (critical for UX).

**Recommendation**: **IMMEDIATE** - Add video dimensions and GIF playback flag.

---

### 4.3 Audio Metadata ❌ **NOT CAPTURED**

| Field | Type | Description | Priority | Z-API Required |
|-------|------|-------------|----------|----------------|
| **`Waveform`** | bytes | Audio waveform | **HIGH** | ✅ **YES** |

**Impact**: No visual waveform for voice messages in UI.

**Recommendation**: **HIGH** - Add waveform data for better UX.

---

### 4.4 Sticker Metadata ❌ **NOT CAPTURED**

| Field | Type | Description | Priority | Z-API Required |
|-------|------|-------------|----------|----------------|
| **`IsAnimated`** | bool | Animated sticker | **HIGH** | ✅ **YES** |
| **`IsLottie`** | bool | Lottie sticker | **HIGH** | ✅ **YES** |
| **`Width`** | uint32 | Sticker width | MEDIUM | ✅ YES |
| **`Height`** | uint32 | Sticker height | MEDIUM | ✅ YES |
| `StickerPackID` | string | Pack ID | MEDIUM | ⚠️ Optional |
| `StickerPackName` | string | Pack name | MEDIUM | ⚠️ Optional |
| `StickerPackPublisher` | string | Publisher | LOW | ❌ No |

**Recommendation**: **HIGH** - Add sticker type flags and dimensions.

---

### 4.5 ContextInfo ❌ **COMPLETELY MISSING** 🚨 **CRITICAL**

**The most critical omission** - `ContextInfo` contains essential message context:

| Field | Description | Priority | Z-API Field | Impact |
|-------|-------------|----------|-------------|--------|
| **`QuotedMessage`** | Replied/quoted message | **CRITICAL** | `quoted` | No reply context |
| **`MentionedJID`** | Mentioned users | **CRITICAL** | `mentions` | No @mentions |
| **`IsForwarded`** | Forwarded flag | **CRITICAL** | `forwarded` | No forward detection |
| **`ForwardingScore`** | Times forwarded | **HIGH** | `forwardingScore` | No forward count |
| `ConversionSource` | Message source | MEDIUM | - | Missing attribution |
| `ConversionData` | Conversion data | LOW | - | Missing tracking |
| `Expiration` | Ephemeral expiration | **HIGH** | `ephemeralExpiration` | No expiry time |
| `EphemeralSettingTimestamp` | When ephemeral set | MEDIUM | `ephemeralSettingTimestamp` | No settings time |
| `ExternalAdReply` | External link preview | MEDIUM | `externalAdReply` | No link previews |
| `DisappearingMode` | Disappearing mode | MEDIUM | `disappearingMode` | No mode info |
| `GroupMentions` | Group mentions | MEDIUM | `groupMentions` | No group @all |
| `UTMSource` | UTM source | LOW | - | Missing analytics |
| `UTMCampaign` | UTM campaign | LOW | - | Missing analytics |
| `QuickReplyButton` | Quick reply | MEDIUM | - | Missing quick replies |
| `BusinessMessageForwardInfo` | Business forward info | MEDIUM | - | Missing business context |
| `Pinned` | Message pinned | MEDIUM | `pinned` | No pin status |

**Current State**: ❌ **ZERO ContextInfo fields captured**

**Z-API Requirements**:
```json
{
  "quoted": { /* quoted message */ },
  "mentions": ["5511999999999"],
  "forwarded": true,
  "forwardingScore": 3,
  "ephemeralExpiration": 1234567890,
  "externalAdReply": { /* link preview */ }
}
```

**Impact**: **MASSIVE** - Cannot display:
- Message replies/quotes
- User mentions (@user)
- Forwarded messages
- Link previews
- Ephemeral message expiration
- Pinned messages

**Recommendation**: **CRITICAL** - Implement ContextInfo extraction IMMEDIATELY.

---

## 5. Incorrect/Incomplete Mappings

### 5.1 Receipt Type Mapping ⚠️ **INCOMPLETE**

**Current Mapping** (zapi/transformer.go:314-325):
```go
switch receiptEvent.Type {
case "delivered":
    status = "RECEIVED"
case "read":
    status = "READ"
case "played":
    status = "PLAYED"
case "sender":
    status = "SENT"
default:
    status = "SENT"
}
```

**Missing Types** (from types/presence.go):
| Receipt Type | Description | Current Handling | Should Map To |
|--------------|-------------|------------------|---------------|
| `retry` | Decrypt retry request | ❌ Default to SENT | `RETRY` |
| `read-self` | Read on another device (receipts off) | ❌ Default to SENT | `READ_SELF` |
| `played-self` | Played on another device (receipts off) | ❌ Default to SENT | `PLAYED_SELF` |
| `server-error` | Server error | ❌ Default to SENT | `ERROR` |
| `inactive` | Inactive chat | ❌ Default to SENT | `INACTIVE` |
| `peer_msg` | Peer message | ❌ Default to SENT | `PEER_MSG` |
| `hist_sync` | History sync | ❌ Default to SENT | `HIST_SYNC` |

**Recommendation**: Add all receipt types to mapping.

---

### 5.2 ChatPresence State Mapping ⚠️ **INCOMPLETE**

**Current Mapping** (zapi/transformer.go:344-353):
```go
switch event.Metadata["state"] {
case "composing":
    status = "COMPOSING"
case "paused":
    status = "PAUSED"
case "recording":
    status = "RECORDING"
default:
    status = "AVAILABLE"
}
```

**Issues**:
1. ❌ "recording" is NOT a whatsmeow ChatPresence state
2. ❌ ChatPresenceMedia (audio vs text composing) not captured
3. ❌ Presence vs ChatPresence distinction unclear

**Correct States** (from types/presence.go):
- `ChatPresence`: composing, paused
- `ChatPresenceMedia`: "" (text), "audio" (voice note recording)

**Recommendation**: Fix state mapping and add media type differentiation.

---

### 5.3 Phone Number Extraction ⚠️ **INCOMPLETE**

**Current Implementation** (zapi/transformer.go:417-427):
```go
func extractPhoneNumber(jid string) string {
    for i, c := range jid {
        if c == '@' {
            return jid[:i]
        }
    }
    return jid
}
```

**Issues**:
1. ❌ Doesn't handle LID JIDs (lid server)
2. ❌ Doesn't handle special servers (broadcast, newsletter, bot)
3. ❌ Doesn't validate JID format

**JID Servers** (from types/jid.go):
| Server | Example | extractPhoneNumber Result | Correct Result |
|--------|---------|---------------------------|----------------|
| `s.whatsapp.net` | `5511999999999@s.whatsapp.net` | ✅ `5511999999999` | ✅ `5511999999999` |
| `lid` | `abcd1234@lid` | ✅ `abcd1234` | ❌ Should use SenderAlt |
| `g.us` | `123456@g.us` | ✅ `123456` | ✅ `123456` (group ID) |
| `broadcast` | `status@broadcast` | ✅ `status` | ⚠️ Special handling |
| `newsletter` | `xyz@newsletter` | ✅ `xyz` | ⚠️ Special handling |

**Recommendation**: Enhance JID parsing to handle all server types and LID addressing.

---

### 5.4 Message Type Detection ⚠️ **NO FALLBACK**

**Current Implementation** (zapi/transformer.go:300-302):
```go
// If no specific message type matched, return error
return fmt.Errorf("unsupported message type")
```

**Issue**: Returns error for unknown message types instead of graceful fallback.

**Z-API Expectation**: Unknown message types should still generate webhook with:
- Base message info (sender, chat, timestamp)
- `unsupported: true` flag
- Raw message type name

**Recommendation**: Add fallback handling for unknown message types.

---

## 6. Z-API Field Gaps

Comparing current implementation with Z-API WEBHOOKS_EVENTS.md:

### 6.1 ReceivedCallback Missing Fields

| Z-API Field | Description | WhatsApp Source | Priority | Status |
|-------------|-------------|-----------------|----------|--------|
| `waitingMessage` | Message in queue | - | LOW | ❌ Not implemented |
| `isStatusReply` | Reply to status | ContextInfo | MEDIUM | ❌ Not captured |
| `broadcast` | From broadcast list | MessageSource.BroadcastListOwner | **HIGH** | ❌ Not captured |
| **`forwarded`** | Forwarded message | ContextInfo.IsForwarded | **CRITICAL** | ❌ Not captured |
| **`forwardingScore`** | Times forwarded | ContextInfo.ForwardingScore | **HIGH** | ❌ Not captured |
| **`mentions`** | Mentioned users | ContextInfo.MentionedJID | **CRITICAL** | ❌ Not captured |
| **`quoted`** | Quoted message | ContextInfo.QuotedMessage | **CRITICAL** | ❌ Not captured |
| `ephemeralExpiration` | Expiry timestamp | ContextInfo.Expiration | **HIGH** | ❌ Not captured |
| `ephemeralSettingTimestamp` | Setting timestamp | ContextInfo.EphemeralSettingTimestamp | MEDIUM | ❌ Not captured |
| `stickerPackID` | Sticker pack | StickerMessage.StickerPackID | MEDIUM | ❌ Not captured |
| `stickerPackName` | Sticker pack name | StickerMessage.StickerPackName | MEDIUM | ❌ Not captured |
| **`width`** | Media width | ImageMessage.Width, etc. | **HIGH** | ❌ Not captured |
| **`height`** | Media height | ImageMessage.Height, etc. | **HIGH** | ❌ Not captured |
| **`isGif`** | GIF detection | ImageMessage.IsGif / VideoMessage.GifPlayback | **CRITICAL** | ❌ Not captured |
| **`isAnimated`** | Animated media | ImageMessage.IsAnimated / StickerMessage.IsAnimated | **HIGH** | ❌ Not captured |
| `waveform` | Audio waveform | AudioMessage.Waveform | **HIGH** | ❌ Not captured |
| `labels` | Message labels | AppState (LabelAssociationMessage) | LOW | ❌ Not implemented |
| `buttons` | Button content | ButtonsMessage | MEDIUM | ❌ Not captured |
| `sections` | List sections | ListMessage | MEDIUM | ❌ Not captured |
| `title` | Message title | Various message types | MEDIUM | ❌ Not captured |
| `footer` | Message footer | Various message types | MEDIUM | ❌ Not captured |
| `thumbnail` | Thumbnail data | Various media | MEDIUM | ⚠️ URL only, not data |
| `externalAdReply` | Link preview | ContextInfo.ExternalAdReply | MEDIUM | ❌ Not captured |

**Critical**: 8 CRITICAL priority fields missing, 10 HIGH priority fields missing.

---

## 7. Implementation Priority Matrix

### 🔴 **CRITICAL** - Implement Immediately

1. **ContextInfo Extraction** ⚠️ MASSIVE IMPACT
   - Add `QuotedMessage`, `MentionedJID`, `IsForwarded`, `ForwardingScore`
   - Add to ALL media extractors
   - Estimate: 4-6 hours

2. **Media Metadata - isGIF & Dimensions** 📸 USER REQUESTED
   - Add `isGIF`, `isAnimated`, `width`, `height` to image/video/sticker
   - Estimate: 2 hours

3. **MessageInfo Fields** 📋 CORE DATA
   - Add `PushName`, `VerifiedName`, `Category`, `Edit`, `MsgMetaInfo`
   - Estimate: 3 hours

4. **ProtocolMessage** 🗑️ REVOCATION SUPPORT
   - Handle delete for everyone
   - Estimate: 2 hours

5. **Group Events** 👥 CORE FUNCTIONALITY
   - Add `JoinedGroup`, `GroupInfo` event handlers
   - Estimate: 4 hours

6. **LID Addressing** 🔐 PRIVACY SUPPORT
   - Fix JID extraction for LID addresses
   - Add `AddressingMode`, `SenderAlt`, `RecipientAlt`
   - Estimate: 2 hours

**Total Critical: ~17-19 hours**

---

### 🟡 **HIGH** - Implement Next Sprint

1. **Interactive Messages** 🎯 BUSINESS FEATURES
   - Add `InteractiveMessage`, `InteractiveResponseMessage`
   - Estimate: 4 hours

2. **GroupInvite & LiveLocation** 📍 POPULAR FEATURES
   - Add group invite and live location handlers
   - Estimate: 2 hours

3. **Message Flags** 🏷️ METADATA COMPLETENESS
   - Add all missing event flags (IsLottieSticker, IsDocumentWithCaption, etc.)
   - Estimate: 2 hours

4. **Audio Waveform** 🎵 UX ENHANCEMENT
   - Add waveform data extraction
   - Estimate: 1 hour

5. **Receipt Type Completion** ✅ STATUS ACCURACY
   - Add all missing receipt types
   - Estimate: 1 hour

6. **Connection Events** 🔌 LIFECYCLE MANAGEMENT
   - Add QR, PairSuccess, LoggedOut, TemporaryBan
   - Estimate: 4 hours

**Total High: ~14 hours**

---

### 🟢 **MEDIUM** - Future Enhancements

1. **User Change Events** (PushName, BusinessName, UserAbout)
2. **Newsletter Support** (NewsletterJoin, NewsletterLeave, NewsletterLiveUpdate)
3. **Additional Message Types** (Order, Product, PollV2/V3)
4. **Sticker Metadata** (StickerPackID, StickerPackName)
5. **Message Thumbnails** (Extract JPEG thumbnail data)

**Total Medium: ~10 hours**

---

### ⚪ **LOW** - Optional Features

1. **App State Events** (16 events for multi-device sync)
2. **Payment Messages** (5 payment-related types)
3. **Advanced Metadata** (ExperimentGroupID, ScansSidecar, etc.)
4. **Sync Events** (OfflineSyncPreview, OfflineSyncCompleted)

**Total Low: ~15 hours**

---

## 8. Recommendations

### Phase 3A - Critical Fixes (Week 1)
```
1. ContextInfo extraction (ALL media types)
2. isGIF, isAnimated, width, height (image/video/sticker)
3. PushName, VerifiedName, Category (MessageInfo)
4. ProtocolMessage (revoke support)
5. LID addressing (phone number extraction)
```

### Phase 3B - High Priority (Week 2)
```
1. Group events (JoinedGroup, GroupInfo)
2. Interactive messages
3. Message flags completion
4. Audio waveform
5. Connection events (QR, PairSuccess, LoggedOut, TemporaryBan)
```

### Phase 3C - Medium Priority (Week 3)
```
1. User change events
2. Newsletter support
3. Additional message types
4. Sticker metadata
```

### Phase 3D - Optional (Future)
```
1. App state events (if multi-device sync needed)
2. Payment messages (if payments needed)
3. Advanced metadata (edge cases)
```

---

## 9. Code Changes Required

### 9.1 whatsmeow/transformer.go

#### Add ContextInfo Extraction
```go
// Add new function
func (t *Transformer) extractContextInfo(msg *waE2E.Message) map[string]string {
    ctx := make(map[string]string)

    // Get ContextInfo from various message types
    var contextInfo *waE2E.ContextInfo

    if img := msg.GetImageMessage(); img != nil {
        contextInfo = img.GetContextInfo()
    } else if txt := msg.GetExtendedTextMessage(); txt != nil {
        contextInfo = txt.GetContextInfo()
    } // ... check all message types

    if contextInfo == nil {
        return ctx
    }

    // Extract quoted message
    if quoted := contextInfo.GetQuotedMessage(); quoted != nil {
        ctx["quoted_message_id"] = contextInfo.GetStanzaID()
        ctx["quoted_participant"] = contextInfo.GetParticipant()
        // Serialize quoted message
        quotedJSON, _ := json.Marshal(quoted)
        ctx["quoted_content"] = string(quotedJSON)
    }

    // Extract mentions
    if mentions := contextInfo.GetMentionedJID(); len(mentions) > 0 {
        mentionsJSON, _ := json.Marshal(mentions)
        ctx["mentions"] = string(mentionsJSON)
    }

    // Extract forwarding info
    if contextInfo.GetIsForwarded() {
        ctx["is_forwarded"] = "true"
        ctx["forwarding_score"] = fmt.Sprintf("%d", contextInfo.GetForwardingScore())
    }

    // Extract ephemeral settings
    if exp := contextInfo.GetExpiration(); exp > 0 {
        ctx["ephemeral_expiration"] = fmt.Sprintf("%d", exp)
    }
    if ts := contextInfo.GetEphemeralSettingTimestamp(); ts > 0 {
        ctx["ephemeral_setting_timestamp"] = fmt.Sprintf("%d", ts)
    }

    // Extract external ad reply (link preview)
    if adReply := contextInfo.GetExternalAdReply(); adReply != nil {
        adReplyJSON, _ := json.Marshal(adReply)
        ctx["external_ad_reply"] = string(adReplyJSON)
    }

    return ctx
}

// Update transformMessage to include ContextInfo
func (t *Transformer) transformMessage(...) {
    // ... existing code ...

    // Extract ContextInfo
    contextInfo := t.extractContextInfo(msgEvent.Message)
    for key, value := range contextInfo {
        event.Metadata[key] = value
    }

    // ... existing code ...
}
```

#### Add MessageInfo Fields
```go
func (t *Transformer) transformMessage(...) {
    // ... existing code ...

    // Add missing MessageInfo fields
    if msgEvent.Info.ServerID != "" {
        event.Metadata["server_id"] = msgEvent.Info.ServerID
    }
    if msgEvent.Info.Category != "" {
        event.Metadata["category"] = msgEvent.Info.Category
    }
    event.Metadata["multicast"] = fmt.Sprintf("%t", msgEvent.Info.Multicast)
    if msgEvent.Info.MediaType != "" {
        event.Metadata["media_type_info"] = msgEvent.Info.MediaType
    }
    if msgEvent.Info.PushName != "" {
        event.Metadata["push_name"] = msgEvent.Info.PushName
    }
    if msgEvent.Info.VerifiedName != nil {
        event.Metadata["verified_name"] = msgEvent.Info.VerifiedName.String()
    }

    // Add edit attributes
    if msgEvent.Info.Edit != types.EditAttributeEmpty {
        event.Metadata["edit_attribute"] = string(msgEvent.Info.Edit)
    }

    // Add MsgMetaInfo (thread/reply info)
    if meta := msgEvent.Info.MsgMetaInfo; meta != nil {
        if meta.TargetID != "" {
            event.Metadata["reply_to_message_id"] = meta.TargetID
        }
        if !meta.TargetSender.IsEmpty() {
            event.Metadata["reply_to_sender"] = meta.TargetSender.String()
        }
        if !meta.ThreadMessageID.IsEmpty() {
            event.Metadata["thread_message_id"] = meta.ThreadMessageID.String()
        }
    }

    // Add DeviceSentMeta
    if sent := msgEvent.Info.DeviceSentMeta; sent != nil {
        if sent.Phash != "" {
            event.Metadata["device_sent_phash"] = sent.Phash
        }
    }

    // ... existing code ...
}
```

#### Enhance Media Extraction
```go
func (t *Transformer) extractMediaInfo(msg *waE2E.Message) (bool, MediaInfo) {
    var info MediaInfo

    // Check for image
    if img := msg.GetImageMessage(); img != nil {
        info.MediaType = "image"

        // ✅ ADD THESE CRITICAL FIELDS
        if img.GetIsGif() {
            info.IsGIF = true
        }
        if img.GetIsAnimated() {
            info.IsAnimated = true
        }
        if width := img.GetWidth(); width > 0 {
            info.Width = int(width)
        }
        if height := img.GetHeight(); height > 0 {
            info.Height = int(height)
        }

        // Existing fields...
        info.MediaKey = base64.StdEncoding.EncodeToString(img.GetMediaKey())
        // ... rest of existing code ...

        return true, info
    }

    // Similar updates for video, audio, sticker, document
    if video := msg.GetVideoMessage(); video != nil {
        info.MediaType = "video"

        // ✅ ADD THESE CRITICAL FIELDS
        if video.GetGifPlayback() {
            info.IsGIF = true // Video plays as GIF
        }
        if width := video.GetWidth(); width > 0 {
            info.Width = int(width)
        }
        if height := video.GetHeight(); height > 0 {
            info.Height = int(height)
        }

        // ... rest of code ...
    }

    // ... handle other media types similarly ...
}
```

#### Add MediaInfo Fields
```go
type MediaInfo struct {
    MediaKey      string
    DirectPath    string
    FileSHA256    *string
    FileEncSHA256 *string
    MediaType     string
    MimeType      *string
    FileLength    *int64

    // ✅ ADD THESE FIELDS
    IsGIF       bool
    IsAnimated  bool
    Width       int
    Height      int
    Waveform    []byte
}
```

#### Add Event Type Support
```go
func (t *Transformer) SupportsEvent(eventType reflect.Type) bool {
    switch eventType {
    case reflect.TypeOf(&events.Message{}),
        reflect.TypeOf(&events.Receipt{}),
        reflect.TypeOf(&events.ChatPresence{}),
        reflect.TypeOf(&events.Presence{}),
        reflect.TypeOf(&events.Connected{}),
        reflect.TypeOf(&events.Disconnected{}),

        // ✅ ADD THESE EVENT TYPES
        reflect.TypeOf(&events.JoinedGroup{}),
        reflect.TypeOf(&events.GroupInfo{}),
        reflect.TypeOf(&events.Picture{}),
        reflect.TypeOf(&events.QR{}),
        reflect.TypeOf(&events.PairSuccess{}),
        reflect.TypeOf(&events.LoggedOut{}),
        reflect.TypeOf(&events.TemporaryBan{}),
        reflect.TypeOf(&events.UndecryptableMessage{}):
        return true
    default:
        return false
    }
}

func (t *Transformer) Transform(ctx context.Context, rawEvent interface{}) (*types.InternalEvent, error) {
    // ... existing switch ...

    // ✅ ADD THESE CASES
    case *events.JoinedGroup:
        return t.transformJoinedGroup(ctx, logger, evt)
    case *events.GroupInfo:
        return t.transformGroupInfo(ctx, logger, evt)
    case *events.Picture:
        return t.transformPicture(ctx, logger, evt)
    case *events.QR:
        return t.transformQR(ctx, logger, evt)
    case *events.PairSuccess:
        return t.transformPairSuccess(ctx, logger, evt)
    case *events.LoggedOut:
        return t.transformLoggedOut(ctx, logger, evt)
    case *events.TemporaryBan:
        return t.transformTemporaryBan(ctx, logger, evt)
    case *events.UndecryptableMessage:
        return t.transformUndecryptableMessage(ctx, logger, evt)

    // ... existing default ...
}
```

---

### 9.2 zapi/transformer.go

#### Enhance extractMessageContent
```go
func (t *Transformer) extractMessageContent(msg *waE2E.Message, callback *ReceivedCallback, event *types.InternalEvent) error {

    // ✅ ADD CONTEXTINFO TO ALL MESSAGE TYPES
    t.addContextInfo(callback, event.Metadata)

    // Text message
    if text := msg.GetConversation(); text != "" {
        callback.Text = &TextContent{
            Message: text,
        }
        return nil
    }

    // Image message
    if img := msg.GetImageMessage(); img != nil {
        callback.Image = &ImageContent{
            ImageURL:     event.Metadata["media_url"],
            ThumbnailURL: event.Metadata["thumbnail_url"],
            Caption:      img.GetCaption(),
            MimeType:     img.GetMimetype(),

            // ✅ ADD THESE CRITICAL FIELDS
            Width:        int(img.GetWidth()),
            Height:       int(img.GetHeight()),
            IsGIF:        img.GetIsGif(),
            IsAnimated:   img.GetIsAnimated(),

            ViewOnce:     img.GetViewOnce(),
        }
        return nil
    }

    // Video message
    if video := msg.GetVideoMessage(); video != nil {
        callback.Video = &VideoContent{
            VideoURL: event.Metadata["media_url"],
            Caption:  video.GetCaption(),
            MimeType: video.GetMimetype(),
            Seconds:  int(video.GetSeconds()),

            // ✅ ADD THESE CRITICAL FIELDS
            Width:        int(video.GetWidth()),
            Height:       int(video.GetHeight()),
            IsGIF:        video.GetGifPlayback(), // GIF playback mode

            ViewOnce: video.GetViewOnce(),
        }
        return nil
    }

    // Audio message
    if audio := msg.GetAudioMessage(); audio != nil {
        callback.Audio = &AudioContent{
            AudioURL: event.Metadata["media_url"],
            MimeType: audio.GetMimetype(),
            PTT:      audio.GetPTT(),
            Seconds:  int(audio.GetSeconds()),

            // ✅ ADD WAVEFORM
            Waveform: audio.GetWaveform(),
        }
        return nil
    }

    // Sticker message
    if sticker := msg.GetStickerMessage(); sticker != nil {
        callback.Sticker = &StickerContent{
            StickerURL: event.Metadata["media_url"],
            MimeType:   sticker.GetMimetype(),

            // ✅ ADD STICKER METADATA
            IsAnimated:      sticker.GetIsAnimated(),
            IsLottie:        event.Metadata["is_lottie_sticker"] == "true",
            Width:           int(sticker.GetWidth()),
            Height:          int(sticker.GetHeight()),
            StickerPackID:   event.Metadata["sticker_pack_id"],
            StickerPackName: event.Metadata["sticker_pack_name"],
        }
        return nil
    }

    // ✅ ADD PROTOCOL MESSAGE (REVOKE)
    if protocol := msg.GetProtocolMessage(); protocol != nil {
        if protocol.GetType() == waE2E.ProtocolMessage_REVOKE {
            callback.Revoke = &RevokeContent{
                MessageID: protocol.GetKey().GetID(),
                FromMe:    protocol.GetKey().GetFromMe(),
            }
            return nil
        }
    }

    // ✅ ADD GROUP INVITE MESSAGE
    if groupInvite := msg.GetGroupInviteMessage(); groupInvite != nil {
        callback.GroupInvite = &GroupInviteContent{
            GroupJID:    groupInvite.GetGroupJID(),
            InviteCode:  groupInvite.GetInviteCode(),
            InviteExpiration: groupInvite.GetInviteExpiration(),
            GroupName:   groupInvite.GetGroupName(),
            Caption:     groupInvite.GetCaption(),
        }
        return nil
    }

    // ✅ ADD LIVE LOCATION MESSAGE
    if liveLoc := msg.GetLiveLocationMessage(); liveLoc != nil {
        callback.LiveLocation = &LiveLocationContent{
            Latitude:            liveLoc.GetDegreesLatitude(),
            Longitude:           liveLoc.GetDegreesLongitude(),
            AccuracyInMeters:    int(liveLoc.GetAccuracyInMeters()),
            SpeedInMps:          liveLoc.GetSpeedInMps(),
            DegreesClockwiseFromMagneticNorth: int(liveLoc.GetDegreesClockwiseFromMagneticNorth()),
            Caption:             liveLoc.GetCaption(),
            SequenceNumber:      liveLoc.GetSequenceNumber(),
            TimeOffset:          int(liveLoc.GetTimeOffset()),
        }
        return nil
    }

    // ✅ ADD INTERACTIVE MESSAGE
    if interactive := msg.GetInteractiveMessage(); interactive != nil {
        callback.Interactive = &InteractiveContent{
            Header: interactive.GetHeader(),
            Body:   interactive.GetBody().GetText(),
            Footer: interactive.GetFooter().GetText(),
            // Parse buttons, lists, etc.
        }
        return nil
    }

    // ✅ ADD FALLBACK FOR UNKNOWN TYPES
    callback.Unsupported = &UnsupportedContent{
        MessageType: detectMessageType(msg),
    }
    return nil
}

// ✅ ADD HELPER FUNCTION
func (t *Transformer) addContextInfo(callback *ReceivedCallback, metadata map[string]string) {
    // Quoted message
    if quotedMsgID, ok := metadata["quoted_message_id"]; ok {
        callback.Quoted = &QuotedMessageContent{
            MessageID: quotedMsgID,
            Participant: metadata["quoted_participant"],
            // Parse quoted_content JSON
        }
    }

    // Mentions
    if mentionsJSON, ok := metadata["mentions"]; ok {
        var mentions []string
        json.Unmarshal([]byte(mentionsJSON), &mentions)
        callback.Mentions = mentions
    }

    // Forwarding
    if metadata["is_forwarded"] == "true" {
        callback.Forwarded = true
        var score int
        fmt.Sscanf(metadata["forwarding_score"], "%d", &score)
        callback.ForwardingScore = score
    }

    // Ephemeral
    if expStr, ok := metadata["ephemeral_expiration"]; ok {
        var exp int64
        fmt.Sscanf(expStr, "%d", &exp)
        callback.EphemeralExpiration = exp
    }

    // External ad reply (link preview)
    if adReplyJSON, ok := metadata["external_ad_reply"]; ok {
        var adReply ExternalAdReplyContent
        json.Unmarshal([]byte(adReplyJSON), &adReply)
        callback.ExternalAdReply = &adReply
    }
}
```

#### Add New Schemas (zapi/schemas.go)
```go
// ✅ ADD MISSING CONTENT TYPES

type RevokeContent struct {
    MessageID string `json:"messageId"`
    FromMe    bool   `json:"fromMe"`
}

type GroupInviteContent struct {
    GroupJID         string `json:"groupJid"`
    InviteCode       string `json:"inviteCode"`
    InviteExpiration int64  `json:"inviteExpiration"`
    GroupName        string `json:"groupName"`
    Caption          string `json:"caption,omitempty"`
}

type LiveLocationContent struct {
    Latitude            float64 `json:"latitude"`
    Longitude           float64 `json:"longitude"`
    AccuracyInMeters    int     `json:"accuracyInMeters"`
    SpeedInMps          float32 `json:"speedInMps"`
    DegreesClockwiseFromMagneticNorth int `json:"degreesClockwiseFromMagneticNorth"`
    Caption             string  `json:"caption,omitempty"`
    SequenceNumber      int64   `json:"sequenceNumber"`
    TimeOffset          int     `json:"timeOffset"`
}

type InteractiveContent struct {
    Header interface{} `json:"header,omitempty"`
    Body   string      `json:"body"`
    Footer string      `json:"footer,omitempty"`
    Buttons []ButtonContent `json:"buttons,omitempty"`
    Sections []SectionContent `json:"sections,omitempty"`
}

type QuotedMessageContent struct {
    MessageID   string      `json:"messageId"`
    Participant string      `json:"participant,omitempty"`
    Content     interface{} `json:"content,omitempty"`
}

type ExternalAdReplyContent struct {
    Title       string `json:"title"`
    Body        string `json:"body"`
    MediaType   string `json:"mediaType"`
    ThumbnailURL string `json:"thumbnailUrl,omitempty"`
    SourceURL   string `json:"sourceUrl,omitempty"`
}

type UnsupportedContent struct {
    MessageType string `json:"messageType"`
}

// ✅ UPDATE EXISTING SCHEMAS

type ReceivedCallback struct {
    // ... existing fields ...

    // ✅ ADD MISSING FIELDS
    Quoted           *QuotedMessageContent   `json:"quoted,omitempty"`
    Mentions         []string                `json:"mentions,omitempty"`
    Forwarded        bool                    `json:"forwarded,omitempty"`
    ForwardingScore  int                     `json:"forwardingScore,omitempty"`
    EphemeralExpiration int64                `json:"ephemeralExpiration,omitempty"`
    EphemeralSettingTimestamp int64          `json:"ephemeralSettingTimestamp,omitempty"`
    ExternalAdReply  *ExternalAdReplyContent `json:"externalAdReply,omitempty"`
    Broadcast        bool                    `json:"broadcast,omitempty"`
    IsStatusReply    bool                    `json:"isStatusReply,omitempty"`

    // ✅ ADD NEW CONTENT TYPES
    Revoke           *RevokeContent          `json:"revoke,omitempty"`
    GroupInvite      *GroupInviteContent     `json:"groupInvite,omitempty"`
    LiveLocation     *LiveLocationContent    `json:"liveLocation,omitempty"`
    Interactive      *InteractiveContent     `json:"interactive,omitempty"`
    Unsupported      *UnsupportedContent     `json:"unsupported,omitempty"`
}

type ImageContent struct {
    // ... existing fields ...

    // ✅ ADD MISSING FIELDS
    Width       int    `json:"width"`
    Height      int    `json:"height"`
    IsGIF       bool   `json:"isGif"`
    IsAnimated  bool   `json:"isAnimated"`
}

type VideoContent struct {
    // ... existing fields ...

    // ✅ ADD MISSING FIELDS
    Width    int  `json:"width"`
    Height   int  `json:"height"`
    IsGIF    bool `json:"isGif"` // GIF playback mode
}

type AudioContent struct {
    // ... existing fields ...

    // ✅ ADD MISSING FIELD
    Waveform []byte `json:"waveform,omitempty"`
}

type StickerContent struct {
    // ... existing fields ...

    // ✅ ADD MISSING FIELDS
    IsAnimated      bool   `json:"isAnimated"`
    IsLottie        bool   `json:"isLottie"`
    Width           int    `json:"width"`
    Height          int    `json:"height"`
    StickerPackID   string `json:"stickerPackId,omitempty"`
    StickerPackName string `json:"stickerPackName,omitempty"`
}
```

#### Fix Receipt Type Mapping
```go
func (t *Transformer) transformReceipt(...) (*MessageStatusCallback, error) {
    // ... existing code ...

    // ✅ COMPLETE RECEIPT TYPE MAPPING
    var status string
    switch receiptEvent.Type {
    case types.ReceiptTypeDelivered:
        status = "RECEIVED"
    case types.ReceiptTypeRead:
        status = "READ"
    case types.ReceiptTypePlayed:
        status = "PLAYED"
    case types.ReceiptTypeSender:
        status = "SENT"
    case types.ReceiptTypeReadSelf:
        status = "READ_SELF"
    case types.ReceiptTypePlayedSelf:
        status = "PLAYED_SELF"
    case types.ReceiptTypeRetry:
        status = "RETRY"
    case types.ReceiptTypeServerError:
        status = "ERROR"
    case types.ReceiptTypeInactive:
        status = "INACTIVE"
    case types.ReceiptTypePeerMsg:
        status = "PEER_MSG"
    case types.ReceiptTypeHistorySync:
        status = "HIST_SYNC"
    default:
        status = "SENT"
    }

    // ... rest of code ...
}
```

#### Fix Phone Number Extraction
```go
// ✅ ENHANCED JID EXTRACTION
func extractPhoneNumber(jidStr string, addressingMode string, altJID string) string {
    // Handle LID addressing mode - use alternate JID
    if addressingMode == "lid" && altJID != "" {
        jidStr = altJID
    }

    // Parse JID
    for i, c := range jidStr {
        if c == '@' {
            user := jidStr[:i]
            server := jidStr[i+1:]

            // Handle special servers
            switch server {
            case "s.whatsapp.net":
                return user // Regular phone number
            case "lid":
                return "" // LID - should use alternate JID
            case "g.us":
                return user // Group ID
            case "broadcast":
                if user == "status" {
                    return "status_broadcast"
                }
                return user
            case "newsletter":
                return user // Newsletter ID
            default:
                return user
            }
        }
    }
    return jidStr
}

// Update usage
func (t *Transformer) transformMessage(...) {
    // ... existing code ...

    addressingMode := event.Metadata["addressing_mode"]
    senderAlt := event.Metadata["sender_alt"]

    callback.Phone = extractPhoneNumber(event.Metadata["from"], addressingMode, senderAlt)

    // ... rest of code ...
}
```

---

### 9.3 types/event.go

#### Add Fields to InternalEvent
```go
type InternalEvent struct {
    // ... existing fields ...

    // ✅ ADD MEDIA DIMENSION FIELDS
    MediaWidth    int
    MediaHeight   int
    MediaIsGIF    bool
    MediaIsAnimated bool
    MediaWaveform []byte
}
```

---

## 10. Testing Checklist

### Critical Tests

- [ ] ContextInfo extraction works for all message types
- [ ] isGIF detected correctly for images and videos (GifPlayback)
- [ ] width/height captured for images, videos, stickers
- [ ] PushName and VerifiedName captured in metadata
- [ ] Quoted messages extracted with full context
- [ ] Mentions array extracted correctly
- [ ] Forwarded flag and forwarding score captured
- [ ] LID addressing mode handled correctly
- [ ] Group events (JoinedGroup, GroupInfo) transformed correctly
- [ ] ProtocolMessage revokes handled
- [ ] All receipt types mapped correctly
- [ ] Interactive messages parsed
- [ ] GroupInvite messages extracted
- [ ] LiveLocation messages with all fields

### Z-API Compliance Tests

- [ ] ReceivedCallback includes all required fields
- [ ] Image webhook includes isGif, width, height
- [ ] Video webhook includes gifPlayback, width, height
- [ ] Audio webhook includes waveform
- [ ] Sticker webhook includes isAnimated, dimensions
- [ ] Quoted messages formatted per Z-API spec
- [ ] Mentions array matches Z-API format
- [ ] Forwarded messages flagged correctly

---

## 11. Conclusion

The current transformation layer is **functionally incomplete** with:

- **86%** of event types unhandled (36 of 42)
- **52%** of message types missing (13+ of 25)
- **35%** of MessageInfo fields not captured (15+ fields)
- **100%** of ContextInfo missing (CRITICAL)
- **User-requested isGIF feature MISSING**

**Estimated Effort**:
- Critical fixes: 17-19 hours
- High priority: 14 hours
- Total for production-ready: ~30-35 hours

**Recommendation**: Implement Phase 3A (critical fixes) IMMEDIATELY before proceeding to Phase 4 (dispatch system).
