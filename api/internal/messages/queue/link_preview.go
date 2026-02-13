package queue

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"google.golang.org/protobuf/proto"

	wameow "go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
)

// LinkPreviewGenerator handles link preview generation with metadata and thumbnails
type LinkPreviewGenerator struct {
	client         *wameow.Client
	httpClient     *http.Client
	thumbGenerator *ThumbnailGenerator
	log            *slog.Logger
}

// LinkPreviewMetadata contains metadata extracted from URL
type LinkPreviewMetadata struct {
	Title       string
	Description string
	Image       string
	ImageThumb  []byte
	Height      *uint32
	Width       *uint32
}

// NewLinkPreviewGenerator creates a new link preview generator
func NewLinkPreviewGenerator(client *wameow.Client, log *slog.Logger) *LinkPreviewGenerator {
	return &LinkPreviewGenerator{
		client: client,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
		thumbGenerator: NewThumbnailGenerator(client, log),
		log:            log.With(slog.String("component", "link_preview")),
	}
}

// Generate generates complete link preview for text message
// If override is provided, it uses the custom metadata instead of fetching from URL
func (l *LinkPreviewGenerator) Generate(ctx context.Context, messageText string, linkPreview *bool, override *LinkPreviewOverride) (*waProto.ExtendedTextMessage, error) {
	// Determine the link URL to use
	var linkURL string

	// If override is provided with URL, use it
	if override != nil && override.URL != "" {
		linkURL = override.URL
	} else {
		// Extract URLs from text
		urls := extractLinksFromText(messageText)
		if len(urls) == 0 {
			return nil, nil
		}
		linkURL = urls[0] // Use first URL
	}

	// Check if link preview is disabled
	if linkPreview != nil && !*linkPreview {
		// Disable link preview by setting ForwardingScore
		return &waProto.ExtendedTextMessage{
			Text: proto.String(messageText),
			ContextInfo: &waProto.ContextInfo{
				ForwardingScore: proto.Uint32(1),
			},
		}, nil
	}

	var metadata *LinkPreviewMetadata

	// Check if we have custom override values
	if override != nil && (override.Title != "" || override.Description != "" || override.Image != "") {
		// Use custom override metadata
		metadata = &LinkPreviewMetadata{
			Title:       override.Title,
			Description: override.Description,
			Image:       override.Image,
		}

		// Download custom image if provided
		if override.Image != "" {
			l.log.Debug("downloading custom thumbnail",
				slog.String("url", override.Image))

			imgReq, err := http.NewRequestWithContext(ctx, "GET", override.Image, nil)
			if err == nil {
				imgReq.Header.Set("User-Agent", "WhatsApp/2.23.24.76")
				imgResp, err := l.httpClient.Do(imgReq)
				if err == nil && imgResp.StatusCode == http.StatusOK {
					defer imgResp.Body.Close()

					contentType := imgResp.Header.Get("Content-Type")
					if strings.HasPrefix(contentType, "image/") {
						imageData, err := io.ReadAll(io.LimitReader(imgResp.Body, 5*1024*1024))
						if err == nil && len(imageData) > 0 {
							metadata.ImageThumb = imageData

							// Get dimensions
							if img, _, err := image.Decode(bytes.NewReader(imageData)); err == nil {
								bounds := img.Bounds()
								width := uint32(bounds.Max.X - bounds.Min.X)
								height := uint32(bounds.Max.Y - bounds.Min.Y)
								metadata.Width = &width
								metadata.Height = &height
							}
						}
					}
				}
			}
		}

		// If no title was provided but we need to fetch it, get from URL
		if metadata.Title == "" || metadata.Description == "" {
			fetchedMeta, err := l.getMetadataFromURL(linkURL)
			if err == nil {
				if metadata.Title == "" {
					metadata.Title = fetchedMeta.Title
				}
				if metadata.Description == "" {
					metadata.Description = fetchedMeta.Description
				}
				// If no custom image was provided, use the fetched one
				if len(metadata.ImageThumb) == 0 && len(fetchedMeta.ImageThumb) > 0 {
					metadata.ImageThumb = fetchedMeta.ImageThumb
					metadata.Width = fetchedMeta.Width
					metadata.Height = fetchedMeta.Height
				}
			}
		}
	} else {
		// Get metadata from URL (standard behavior)
		var err error
		metadata, err = l.getMetadataFromURL(linkURL)
		if err != nil {
			l.log.Warn("failed to get metadata",
				slog.String("url", linkURL),
				slog.String("error", err.Error()))
			return nil, nil
		}
	}

	// Build extended text message with link preview
	extMsg := &waProto.ExtendedTextMessage{
		Text:        proto.String(messageText),
		Title:       proto.String(metadata.Title),
		MatchedText: proto.String(linkURL),
		Description: proto.String(metadata.Description),
	}

	// Determine preview type (video or image)
	if isVideoLink(linkURL) {
		previewType := waProto.ExtendedTextMessage_VIDEO
		extMsg.PreviewType = &previewType
		extMsg.DoNotPlayInline = proto.Bool(false)
	} else {
		previewType := waProto.ExtendedTextMessage_IMAGE
		extMsg.PreviewType = &previewType
		extMsg.DoNotPlayInline = proto.Bool(true)
	}

	// Upload thumbnail if available
	if len(metadata.ImageThumb) > 0 {
		uploaded, err := l.client.Upload(context.Background(), metadata.ImageThumb, wameow.MediaLinkThumbnail)
		if err == nil {
			extMsg.ThumbnailDirectPath = proto.String(uploaded.DirectPath)
			extMsg.ThumbnailSHA256 = uploaded.FileSHA256
			extMsg.ThumbnailEncSHA256 = uploaded.FileEncSHA256
			extMsg.MediaKey = uploaded.MediaKey

			if metadata.Height != nil {
				extMsg.ThumbnailHeight = metadata.Height
			}
			if metadata.Width != nil {
				extMsg.ThumbnailWidth = metadata.Width
			}
		} else {
			// Fallback to inline JPEG thumbnail
			extMsg.JPEGThumbnail = metadata.ImageThumb
		}
	}

	return extMsg, nil
}

// getMetadataFromURL fetches metadata from URL (YouTube, Instagram, etc.)
func (l *LinkPreviewGenerator) getMetadataFromURL(urlStr string) (*LinkPreviewMetadata, error) {
	// Check for YouTube
	if strings.Contains(urlStr, "youtube.com") || strings.Contains(urlStr, "youtu.be") {
		return l.extractYouTubeMetadata(urlStr)
	}

	// Generic metadata extraction
	return l.extractGenericMetadata(urlStr)
}

// extractYouTubeMetadata extracts metadata from YouTube URLs
func (l *LinkPreviewGenerator) extractYouTubeMetadata(videoURL string) (*LinkPreviewMetadata, error) {
	videoID := extractYouTubeVideoID(videoURL)
	if videoID == "" {
		return l.extractGenericMetadata(videoURL)
	}

	// Try oEmbed API first
	oembedURL := fmt.Sprintf("https://www.youtube.com/oembed?url=%s&format=json", url.QueryEscape(videoURL))

	req, _ := http.NewRequest("GET", oembedURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json, */*")
	req.Header.Set("Referer", "https://www.youtube.com/")

	resp, err := l.httpClient.Do(req)
	if err == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()

		var oembedData map[string]interface{}
		if json.NewDecoder(resp.Body).Decode(&oembedData) == nil {
			metadata := &LinkPreviewMetadata{}

			if title, ok := oembedData["title"].(string); ok {
				metadata.Title = title
			}
			if authorName, ok := oembedData["author_name"].(string); ok {
				metadata.Description = fmt.Sprintf("Por %s", authorName)
			}
			if thumbnailURL, ok := oembedData["thumbnail_url"].(string); ok {
				// Download thumbnail
				thumbReq, _ := http.NewRequest("GET", thumbnailURL, nil)
				thumbReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
				thumbReq.Header.Set("Referer", "https://www.youtube.com/")

				imgResp, err := l.httpClient.Do(thumbReq)
				if err == nil && imgResp.StatusCode == http.StatusOK {
					defer imgResp.Body.Close()
					imageData, _ := io.ReadAll(io.LimitReader(imgResp.Body, 3*1024*1024))
					if len(imageData) > 0 {
						metadata.ImageThumb = imageData
						metadata.Image = thumbnailURL

						// Set dimensions
						if width, ok := oembedData["width"].(float64); ok {
							w := uint32(width)
							metadata.Width = &w
						} else {
							defaultWidth := uint32(480)
							metadata.Width = &defaultWidth
						}
						if height, ok := oembedData["height"].(float64); ok {
							h := uint32(height)
							metadata.Height = &h
						} else {
							defaultHeight := uint32(360)
							metadata.Height = &defaultHeight
						}
					}
				}
			}

			return metadata, nil
		}
	}

	// Fallback to generic extraction
	return l.extractGenericMetadata(videoURL)
}

// extractGenericMetadata extracts metadata from generic URLs
func (l *LinkPreviewGenerator) extractGenericMetadata(urlStr string) (*LinkPreviewMetadata, error) {
	baseURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %v", err)
	}

	req, _ := http.NewRequest("GET", urlStr, nil)
	req.Header.Set("User-Agent", "WhatsApp/2.23.24.76")

	response, err := l.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status: %s", response.Status)
	}

	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, err
	}

	metadata := &LinkPreviewMetadata{}

	// Extract og:title
	document.Find("meta[property='og:title']").Each(func(i int, s *goquery.Selection) {
		if content, exists := s.Attr("content"); exists {
			metadata.Title = content
		}
	})

	// Fallback to <title>
	if metadata.Title == "" {
		document.Find("title").Each(func(i int, s *goquery.Selection) {
			metadata.Title = s.Text()
		})
	}

	// Extract description
	document.Find("meta[name='description']").Each(func(i int, s *goquery.Selection) {
		metadata.Description, _ = s.Attr("content")
	})

	// Extract og:image
	document.Find("meta[property='og:image']").Each(func(i int, s *goquery.Selection) {
		if content, exists := s.Attr("content"); exists {
			metadata.Image = content
		}
	})

	// Fallback to twitter:image
	if metadata.Image == "" {
		document.Find("meta[name='twitter:image']").Each(func(i int, s *goquery.Selection) {
			if content, exists := s.Attr("content"); exists {
				metadata.Image = content
			}
		})
	}

	// Download image thumbnail
	if metadata.Image != "" {
		imgURL, err := url.Parse(metadata.Image)
		if err == nil {
			metadata.Image = baseURL.ResolveReference(imgURL).String()
		}

		imgResponse, err := l.httpClient.Get(metadata.Image)
		if err == nil && imgResponse.StatusCode == http.StatusOK {
			defer imgResponse.Body.Close()

			contentType := imgResponse.Header.Get("Content-Type")
			if strings.HasPrefix(contentType, "image/") {
				imageData, err := io.ReadAll(io.LimitReader(imgResponse.Body, 5*1024*1024))
				if err == nil && len(imageData) > 0 {
					metadata.ImageThumb = imageData

					// Get dimensions
					if img, _, err := image.Decode(bytes.NewReader(imageData)); err == nil {
						bounds := img.Bounds()
						width := uint32(bounds.Max.X - bounds.Min.X)
						height := uint32(bounds.Max.Y - bounds.Min.Y)
						metadata.Width = &width
						metadata.Height = &height
					}
				}
			}
		}
	}

	return metadata, nil
}

// extractLinksFromText extracts all URLs from text
func extractLinksFromText(text string) []string {
	urlRegex := regexp.MustCompile(`https?://[^\s]+`)
	return urlRegex.FindAllString(text, -1)
}

// extractYouTubeVideoID extracts video ID from YouTube URL
func extractYouTubeVideoID(videoURL string) string {
	patterns := []string{
		`(?:youtube\.com/watch\?v=)([^&\n?#]+)`,
		`(?:youtu\.be/)([^&\n?#]+)`,
		`(?:youtube\.com/embed/)([^&\n?#]+)`,
		`(?:youtube\.com/v/)([^&\n?#]+)`,
		`(?:youtube\.com/shorts/)([^&\n?#]+)`,
	}

	for _, pattern := range patterns {
		if matches := regexp.MustCompile(pattern).FindStringSubmatch(videoURL); len(matches) > 1 {
			videoID := strings.TrimSpace(matches[1])
			if len(videoID) == 11 {
				return videoID
			}
		}
	}
	return ""
}

// isVideoLink checks if URL is a video link
func isVideoLink(urlStr string) bool {
	videoSites := []string{
		"youtube.com", "youtu.be",
		"vimeo.com",
		"dailymotion.com",
		"twitch.tv",
	}

	videoPaths := []string{
		"/reel/", "/reels/",
		"/video/", "/videos/",
		"/watch?v=",
	}

	urlLower := strings.ToLower(urlStr)

	for _, site := range videoSites {
		if strings.Contains(urlLower, site) {
			return true
		}
	}

	for _, path := range videoPaths {
		if strings.Contains(urlLower, path) {
			return true
		}
	}

	return false
}
