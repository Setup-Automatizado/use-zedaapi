package queue

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/proto"

	wameow "go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"

	"go.mau.fi/whatsmeow/api/internal/events/echo"
)

// DocumentProcessor handles document message sending via WhatsApp
type DocumentProcessor struct {
	log             *slog.Logger
	mediaDownloader *MediaDownloader
	thumbGenerator  *ThumbnailGenerator
	presenceHelper  *PresenceHelper
	echoEmitter     *echo.Emitter
}

// NewDocumentProcessor creates a new document message processor
func NewDocumentProcessor(log *slog.Logger, echoEmitter *echo.Emitter) *DocumentProcessor {
	return &DocumentProcessor{
		log:             log.With(slog.String("processor", "document")),
		mediaDownloader: NewMediaDownloader(1000), // 1000MB max for documents
		presenceHelper:  NewPresenceHelper(),
		echoEmitter:     echoEmitter,
	}
}

// Process sends a document message via WhatsApp
func (p *DocumentProcessor) Process(ctx context.Context, client *wameow.Client, args *SendMessageArgs) error {
	if args.DocumentContent == nil {
		return fmt.Errorf("document_content is required for document messages")
	}

	// Parse recipient JID
	recipientJID, err := types.ParseJID(args.Phone)
	if err != nil {
		return fmt.Errorf("invalid phone number: %w", err)
	}

	// Simulate typing indicator if DelayTyping is set
	if args.DelayTyping > 0 {
		if err := p.presenceHelper.SimulateTyping(client, recipientJID, args.DelayTyping); err != nil {
			p.log.Warn("failed to send typing indicator",
				slog.String("error", err.Error()),
				slog.String("phone", args.Phone))
		}
	}

	// Download or decode document data using helper
	documentData, mimeType, err := p.mediaDownloader.Download(args.DocumentContent.MediaURL)
	if err != nil {
		return fmt.Errorf("download document: %w", err)
	}

	// Validate media type
	if err := p.mediaDownloader.ValidateMediaType(mimeType, MediaTypeDocument); err != nil {
		return fmt.Errorf("invalid media type: %w", err)
	}

	// Determine filename
	fileName := p.determineFilename(args.DocumentContent, mimeType)

	// Generate and upload thumbnail (lazy initialization)
	// Note: Document thumbnail generation requires additional libraries and is not yet implemented
	var thumbnail *ThumbnailResult
	if p.thumbGenerator == nil {
		p.thumbGenerator = NewThumbnailGenerator(client, p.log)
	}
	thumbnail, err = p.thumbGenerator.GenerateAndUploadDocumentThumbnail(ctx, documentData, mimeType)
	if err != nil {
		p.log.Warn("failed to generate document thumbnail (feature not yet implemented)",
			slog.String("error", err.Error()),
			slog.String("phone", args.Phone))
	}

	// Upload document to WhatsApp servers
	uploaded, err := client.Upload(ctx, documentData, wameow.MediaDocument)
	if err != nil {
		return fmt.Errorf("upload document: %w", err)
	}

	// Build ContextInfo using helper
	contextBuilder := NewContextInfoBuilder(client, recipientJID, args, p.log)
	contextInfo, err := contextBuilder.Build(ctx)
	if err != nil {
		p.log.Warn("failed to build context info, sending without it",
			slog.String("error", err.Error()),
			slog.String("phone", args.Phone))
	}

	// Build document message
	msg := p.buildMessage(args, uploaded, mimeType, fileName, thumbnail, contextInfo)

	// Send message
	resp, err := client.SendMessage(ctx, recipientJID, msg, BuildSendExtra(args))
	if err != nil {
		return fmt.Errorf("send document message: %w", err)
	}

	p.log.Info("document message sent successfully",
		slog.String("zaap_id", args.ZaapID),
		slog.String("phone", args.Phone),
		slog.String("whatsapp_message_id", resp.ID),
		slog.String("filename", fileName),
		slog.Int64("file_size", int64(uploaded.FileLength)),
		slog.Bool("has_thumbnail", thumbnail != nil),
		slog.Time("timestamp", resp.Timestamp))

	args.WhatsAppMessageID = resp.ID

	// Emit API echo event for webhook notification
	if p.echoEmitter != nil {
		echoReq := &echo.EchoRequest{
			InstanceID:        args.InstanceID,
			WhatsAppMessageID: resp.ID,
			RecipientJID:      recipientJID,
			Message:           msg,
			Timestamp:         resp.Timestamp,
			MessageType:       "document",
			MediaType:         "document",
			ZaapID:            args.ZaapID,
			HasMedia:          true,
		}
		if err := p.echoEmitter.EmitEcho(ctx, echoReq); err != nil {
			p.log.Warn("failed to emit API echo",
				slog.String("error", err.Error()),
				slog.String("zaap_id", args.ZaapID))
		}
	}

	return nil
}

// determineFilename determines the filename for the document
func (p *DocumentProcessor) determineFilename(content *MediaMessage, mimeType string) string {
	// Use provided filename if available
	if content.FileName != nil && *content.FileName != "" {
		return *content.FileName
	}

	// Try to extract from URL
	if !strings.HasPrefix(content.MediaURL, "data:") {
		fileName := p.extractFilenameFromURL(content.MediaURL)
		if fileName != "" && fileName != "." && fileName != "/" {
			return fileName
		}
	}

	// Generate from MIME type
	return p.filenameFromMimeType(mimeType)
}

// extractFilenameFromURL extracts filename from URL path
func (p *DocumentProcessor) extractFilenameFromURL(url string) string {
	// Remove query parameters
	if idx := strings.Index(url, "?"); idx != -1 {
		url = url[:idx]
	}

	// Get base filename
	return filepath.Base(url)
}

// filenameFromMimeType generates a filename based on MIME type
// Supports 58+ document formats including Microsoft Office, LibreOffice, archives, ebooks, and more
func (p *DocumentProcessor) filenameFromMimeType(mimeType string) string {
	extensions := map[string]string{
		// Microsoft Office (legacy)
		"application/msword":            "document.doc",
		"application/vnd.ms-excel":      "spreadsheet.xls",
		"application/vnd.ms-powerpoint": "presentation.ppt",
		"application/vnd.ms-access":     "database.mdb",

		// Microsoft Office (modern - Office Open XML)
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   "document.docx",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         "spreadsheet.xlsx",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": "presentation.pptx",

		// LibreOffice/OpenOffice (ODF)
		"application/vnd.oasis.opendocument.text":         "document.odt",
		"application/vnd.oasis.opendocument.spreadsheet":  "spreadsheet.ods",
		"application/vnd.oasis.opendocument.presentation": "presentation.odp",
		"application/vnd.oasis.opendocument.graphics":     "drawing.odg",
		"application/vnd.oasis.opendocument.formula":      "formula.odf",

		// Apple iWork
		"application/x-iwork-pages-sffpages":     "document.pages",
		"application/x-iwork-numbers-sffnumbers": "spreadsheet.numbers",
		"application/x-iwork-keynote-sffkey":     "presentation.key",

		// Adobe
		"application/pdf": "document.pdf",

		// Archives and Compression
		"application/zip":              "archive.zip",
		"application/x-rar-compressed": "archive.rar",
		"application/x-rar":            "archive.rar",
		"application/vnd.rar":          "archive.rar",
		"application/x-7z-compressed":  "archive.7z",
		"application/gzip":             "archive.gz",
		"application/x-gzip":           "archive.gz",
		"application/x-tar":            "archive.tar",
		"application/x-bzip2":          "archive.bz2",
		"application/x-xz":             "archive.xz",

		// Text formats
		"text/plain":            "document.txt",
		"text/csv":              "data.csv",
		"text/markdown":         "document.md",
		"text/html":             "document.html",
		"application/xhtml+xml": "document.xhtml",
		"application/rtf":       "document.rtf",

		// Data formats
		"application/json": "data.json",
		"application/xml":  "document.xml",
		"text/xml":         "document.xml",

		// eBooks
		"application/epub+zip":           "ebook.epub",
		"application/x-mobipocket-ebook": "ebook.mobi",

		// Database
		"application/x-sqlite3": "database.sqlite",

		// CAD and Design
		"application/vnd.visio":            "diagram.vsd",
		"application/vnd.ms-visio.drawing": "diagram.vsdx",
		"image/vnd.dwg":                    "drawing.dwg",
		"image/vnd.dxf":                    "drawing.dxf",
		"application/vnd.corel-draw":       "drawing.cdr",
		"application/vnd.adobe.photoshop":  "drawing.psd",
		"application/vnd.cad":              "drawing.cad",
		"application/vnd.dxf":              "drawing.dxf",

		// Source Code
		"text/x-python":          "code.py",
		"text/x-java":            "code.java",
		"text/x-c":               "code.c",
		"text/x-c++":             "code.cpp",
		"text/x-csharp":          "code.cs",
		"application/javascript": "code.js",
		"application/typescript": "code.ts",
		"text/x-go":              "code.go",
		"text/x-rust":            "code.rs",
		"text/x-php":             "code.php",
		"text/x-ruby":            "code.rb",
		"text/x-swift":           "code.swift",
		"text/x-kotlin":          "code.kt",

		// Configuration files
		"application/x-yaml": "config.yaml",
		"application/toml":   "config.toml",
		"application/x-sh":   "script.sh",
		"application/x-perl": "script.pl",
	}

	// Return mapped filename if found
	if filename, ok := extensions[mimeType]; ok {
		return filename
	}

	// Fallback to generic binary file
	return "document.bin"
}

// buildMessage constructs the document message proto
func (p *DocumentProcessor) buildMessage(
	args *SendMessageArgs,
	uploaded wameow.UploadResponse,
	mimeType string,
	fileName string,
	thumbnail *ThumbnailResult,
	contextInfo *waProto.ContextInfo,
) *waProto.Message {
	docMsg := &waProto.DocumentMessage{
		URL:           proto.String(uploaded.URL),
		DirectPath:    proto.String(uploaded.DirectPath),
		MediaKey:      uploaded.MediaKey,
		Mimetype:      proto.String(mimeType),
		FileEncSHA256: uploaded.FileEncSHA256,
		FileSHA256:    uploaded.FileSHA256,
		FileLength:    proto.Uint64(uploaded.FileLength),
		FileName:      proto.String(fileName),
		ContextInfo:   contextInfo,
	}

	// Add caption if provided
	if args.DocumentContent.Caption != nil && *args.DocumentContent.Caption != "" {
		docMsg.Caption = args.DocumentContent.Caption
	}

	// Add thumbnail if generated successfully
	if thumbnail != nil {
		docMsg.JPEGThumbnail = thumbnail.Data
		if thumbnail.DirectPath != "" {
			docMsg.ThumbnailDirectPath = proto.String(thumbnail.DirectPath)
		}
		if len(thumbnail.FileSha256) > 0 {
			docMsg.ThumbnailSHA256 = thumbnail.FileSha256
		}
		if len(thumbnail.FileEncSha256) > 0 {
			docMsg.ThumbnailEncSHA256 = thumbnail.FileEncSha256
		}
		if thumbnail.Width > 0 {
			docMsg.ThumbnailWidth = proto.Uint32(thumbnail.Width)
		}
		if thumbnail.Height > 0 {
			docMsg.ThumbnailHeight = proto.Uint32(thumbnail.Height)
		}
		if thumbnail.PageCount > 0 {
			docMsg.PageCount = proto.Uint32(thumbnail.PageCount)
		}
	}

	return &waProto.Message{
		DocumentMessage: docMsg,
	}
}
