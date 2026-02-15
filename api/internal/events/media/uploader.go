package media

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"

	"go.mau.fi/whatsmeow/api/internal/config"
	"go.mau.fi/whatsmeow/api/internal/logging"
	"go.mau.fi/whatsmeow/api/internal/observability"
)

type S3Uploader struct {
	client        *s3.Client
	uploader      *manager.Uploader
	bucket        string
	urlExpiry     time.Duration
	acl           string
	usePresigned  bool
	publicBaseURL string
	metrics       *observability.Metrics
	logger        *slog.Logger
}

func NewS3Uploader(ctx context.Context, cfg *config.Config, metrics *observability.Metrics) (*S3Uploader, error) {
	logger := logging.ContextLogger(ctx, nil).With(
		slog.String("component", "s3_uploader"),
	)

	awsCfg := aws.Config{
		Region: cfg.S3.Region,
	}

	// Usar credenciais estaticas APENAS se fornecidas (dev/MinIO)
	// Em producao/homolog com ECS, IAM Role sera usado automaticamente
	if cfg.S3.AccessKey != "" && cfg.S3.SecretKey != "" {
		logger.Info("using static S3 credentials (MinIO/dev mode)",
			slog.String("credential_mode", "static"))
		awsCfg.Credentials = credentials.NewStaticCredentialsProvider(cfg.S3.AccessKey, cfg.S3.SecretKey, "")
	} else {
		logger.Info("using IAM Role credentials chain",
			slog.String("credential_mode", "iam_role"))
		// AWS SDK usara automaticamente: env vars -> IAM Role -> EC2 metadata
	}

	if cfg.S3.Endpoint != "" {
		awsCfg.BaseEndpoint = aws.String(cfg.S3.Endpoint)
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	uploader := manager.NewUploader(s3Client, func(u *manager.Uploader) {
		u.PartSize = cfg.Events.MediaChunkSize
		u.Concurrency = 100
	})

	logger.Info("s3 uploader initialized",
		slog.String("bucket", cfg.S3.Bucket),
		slog.String("region", cfg.S3.Region),
		slog.Duration("url_expiry", cfg.S3.URLExpiry),
		slog.Bool("use_presigned_urls", cfg.S3.UsePresignedURLs),
		slog.String("public_base_url", cfg.S3.PublicBaseURL),
		slog.String("acl", cfg.S3.ACL))

	if cfg.S3.ACL != "" {
		logger.Warn("using S3 ACL (deprecated AWS pattern - prefer bucket policies)",
			slog.String("acl", cfg.S3.ACL))
	}

	return &S3Uploader{
		client:        s3Client,
		uploader:      uploader,
		bucket:        cfg.S3.Bucket,
		urlExpiry:     cfg.S3.URLExpiry,
		acl:           cfg.S3.ACL,
		usePresigned:  cfg.S3.UsePresignedURLs,
		publicBaseURL: strings.TrimSuffix(cfg.S3.PublicBaseURL, "/"),
		metrics:       metrics,
		logger:        logger,
	}, nil
}

func (u *S3Uploader) Upload(ctx context.Context, instanceID uuid.UUID, eventID uuid.UUID, mediaType string, reader io.Reader, contentType string, fileSize int64) (key string, mediaURL string, err error) {
	logger := logging.ContextLogger(ctx, u.logger).With(
		slog.String("instance_id", instanceID.String()),
		slog.String("event_id", eventID.String()),
		slog.String("media_type", mediaType),
		slog.Int64("file_size", fileSize))

	start := time.Now()

	key = u.generateKey(instanceID, eventID, mediaType, contentType)

	logger.Debug("uploading media to s3", slog.String("s3_key", key))

	uploadInput := &s3.PutObjectInput{
		Bucket:      aws.String(u.bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(contentType),
		Metadata: map[string]string{
			"instance-id": instanceID.String(),
			"event-id":    eventID.String(),
			"media-type":  mediaType,
		},
	}

	if u.acl != "" {
		uploadInput.ACL = types.ObjectCannedACL(u.acl)
	}

	uploadResult, err := u.uploader.Upload(ctx, uploadInput)

	duration := time.Since(start)

	if err != nil {
		logger.Error("s3 upload failed",
			slog.String("error", err.Error()),
			slog.Duration("duration", duration))

		u.metrics.MediaUploadsTotal.WithLabelValues(instanceID.String(), mediaType, "failure").Inc()
		u.metrics.MediaUploadAttempts.WithLabelValues("failure").Inc()
		u.metrics.MediaUploadErrors.WithLabelValues(classifyS3Error(err)).Inc()

		return "", "", fmt.Errorf("s3 upload failed: %w", err)
	}

	logger.Info("s3 upload succeeded",
		slog.String("location", uploadResult.Location),
		slog.Duration("duration", duration))

	u.metrics.MediaUploadsTotal.WithLabelValues(instanceID.String(), mediaType, "success").Inc()
	u.metrics.MediaUploadAttempts.WithLabelValues("success").Inc()
	u.metrics.MediaUploadDuration.WithLabelValues(instanceID.String(), mediaType).Observe(duration.Seconds())
	u.metrics.MediaUploadSizeBytes.WithLabelValues(mediaType).Add(float64(fileSize))

	if u.usePresigned {
		mediaURL, err = u.GeneratePresignedURL(ctx, key)
		if err != nil {
			return key, "", fmt.Errorf("failed to generate presigned URL: %w", err)
		}
	} else {
		mediaURL = u.buildPublicURL(key, uploadResult.Location)
	}

	return key, mediaURL, nil
}

func (u *S3Uploader) GeneratePresignedURL(ctx context.Context, key string) (string, error) {
	if !u.usePresigned {
		return "", fmt.Errorf("presigned URLs are disabled")
	}

	logger := logging.ContextLogger(ctx, u.logger).With(
		slog.String("s3_key", key))

	presignClient := s3.NewPresignClient(u.client)

	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = u.urlExpiry
	})

	if err != nil {
		logger.Error("failed to generate presigned URL",
			slog.String("error", err.Error()))
		return "", fmt.Errorf("presign URL generation failed: %w", err)
	}

	logger.Debug("presigned URL generated",
		slog.String("url", req.URL),
		slog.Time("expires_at", time.Now().Add(u.urlExpiry)))

	u.metrics.MediaPresignedURLGenerated.Inc()

	return req.URL, nil
}

func (u *S3Uploader) buildPublicURL(key string, uploadLocation string) string {
	if u.publicBaseURL != "" {
		return fmt.Sprintf("%s/%s/%s", u.publicBaseURL, u.bucket, key)
	}
	return uploadLocation
}

func (u *S3Uploader) UsesPresignedURLs() bool {
	return u.usePresigned
}

func (u *S3Uploader) Delete(ctx context.Context, key string) error {
	return u.DeleteObject(ctx, u.bucket, key)
}

func (u *S3Uploader) DeleteObject(ctx context.Context, bucket, key string) error {
	if key == "" {
		return fmt.Errorf("s3 delete failed: empty key")
	}
	if bucket == "" {
		bucket = u.bucket
	}

	logger := logging.ContextLogger(ctx, u.logger).With(
		slog.String("s3_bucket", bucket),
		slog.String("s3_key", key))

	logger.Debug("deleting media from s3")

	_, err := u.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		logger.Error("s3 delete failed",
			slog.String("error", err.Error()))

		u.metrics.MediaDeleteAttempts.WithLabelValues("failure").Inc()
		return fmt.Errorf("s3 delete failed: %w", err)
	}

	logger.Info("s3 delete succeeded")
	u.metrics.MediaDeleteAttempts.WithLabelValues("success").Inc()

	return nil
}

func (u *S3Uploader) generateKey(instanceID uuid.UUID, eventID uuid.UUID, mediaType string, contentType string) string {
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")

	ext := u.getExtensionFromContentType(contentType, mediaType)

	return fmt.Sprintf("%s/%s/%s/%s/%s.%s",
		instanceID.String(), year, month, day, eventID.String(), ext)
}

// mimeExtensionMap is a comprehensive mapping of MIME types to preferred file extensions.
// This is deterministic and OS-independent, unlike mime.ExtensionsByType which reads
// from the host's MIME database and may return different results (e.g., "image/jpeg" → ".jfif"
// on some Linux systems). Covers all WhatsApp media types and common file formats.
var mimeExtensionMap = map[string]string{
	// ── Image ──────────────────────────────────────────────────────────
	"image/jpeg":                "jpg",
	"image/png":                 "png",
	"image/gif":                 "gif",
	"image/webp":                "webp",
	"image/bmp":                 "bmp",
	"image/x-bmp":               "bmp",
	"image/x-ms-bmp":            "bmp",
	"image/tiff":                "tiff",
	"image/x-tiff":              "tiff",
	"image/svg+xml":             "svg",
	"image/heic":                "heic",
	"image/heif":                "heif",
	"image/avif":                "avif",
	"image/x-icon":              "ico",
	"image/vnd.microsoft.icon":  "ico",
	"image/x-portable-pixmap":   "ppm",
	"image/x-portable-graymap":  "pgm",
	"image/x-portable-bitmap":   "pbm",
	"image/x-portable-anymap":   "pnm",
	"image/x-tga":               "tga",
	"image/vnd.adobe.photoshop": "psd",
	"image/x-photoshop":         "psd",
	"image/x-xcf":               "xcf",
	"image/x-raw":               "raw",
	"image/jp2":                 "jp2",
	"image/jxl":                 "jxl",

	// ── Video ──────────────────────────────────────────────────────────
	"video/mp4":               "mp4",
	"video/3gpp":              "3gp",
	"video/3gpp2":             "3g2",
	"video/mpeg":              "mpeg",
	"video/quicktime":         "mov",
	"video/x-msvideo":         "avi",
	"video/x-matroska":        "mkv",
	"video/webm":              "webm",
	"video/x-flv":             "flv",
	"video/ogg":               "ogv",
	"video/x-ms-wmv":          "wmv",
	"video/x-ms-asf":          "asf",
	"video/mp2t":              "ts",
	"video/x-m4v":             "m4v",
	"video/h264":              "h264",
	"video/h265":              "h265",
	"video/vnd.dlna.mpeg-tts": "ts",

	// ── Audio (including WhatsApp-specific MIME variants with parameters) ──
	"audio/ogg":              "ogg",
	"audio/ogg; codecs=opus": "ogg",
	"audio/ogg;codecs=opus":  "ogg",
	"audio/mpeg":             "mp3",
	"audio/mp4":              "m4a",
	"audio/aac":              "aac",
	"audio/amr":              "amr",
	"audio/amr-wb":           "amr",
	"audio/wav":              "wav",
	"audio/x-wav":            "wav",
	"audio/vnd.wave":         "wav",
	"audio/flac":             "flac",
	"audio/x-flac":           "flac",
	"audio/webm":             "weba",
	"audio/x-m4a":            "m4a",
	"audio/mp3":              "mp3",
	"audio/opus":             "opus",
	"audio/vorbis":           "ogg",
	"audio/x-aac":            "aac",
	"audio/x-ms-wma":         "wma",
	"audio/basic":            "au",
	"audio/midi":             "mid",
	"audio/x-midi":           "mid",
	"audio/x-aiff":           "aiff",
	"audio/aiff":             "aiff",
	"audio/x-pn-realaudio":   "ra",
	"audio/x-matroska":       "mka",
	"audio/ac3":              "ac3",
	"audio/x-ape":            "ape",

	// ── Document - Office ──────────────────────────────────────────────
	"application/pdf":    "pdf",
	"application/rtf":    "rtf",
	"application/msword": "doc",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   "docx",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.template":   "dotx",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         "xlsx",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.template":      "xltx",
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": "pptx",
	"application/vnd.openxmlformats-officedocument.presentationml.template":     "potx",
	"application/vnd.openxmlformats-officedocument.presentationml.slideshow":    "ppsx",
	"application/vnd.ms-excel":                              "xls",
	"application/vnd.ms-excel.sheet.macroEnabled.12":        "xlsm",
	"application/vnd.ms-excel.sheet.binary.macroEnabled.12": "xlsb",
	"application/vnd.ms-powerpoint":                         "ppt",
	"application/vnd.ms-word.document.macroEnabled.12":      "docm",
	"application/vnd.oasis.opendocument.text":               "odt",
	"application/vnd.oasis.opendocument.spreadsheet":        "ods",
	"application/vnd.oasis.opendocument.presentation":       "odp",
	"application/vnd.oasis.opendocument.graphics":           "odg",
	"application/vnd.visio":                                 "vsd",
	"application/vnd.ms-visio.drawing":                      "vsdx",
	"application/vnd.ms-access":                             "mdb",
	"application/x-mspublisher":                             "pub",
	"application/vnd.ms-project":                            "mpp",
	"application/vnd.ms-works":                              "wps",
	"application/vnd.ms-xpsdocument":                        "xps",

	// ── Document - Archive & Compression ───────────────────────────────
	"application/zip":                         "zip",
	"application/x-zip-compressed":            "zip",
	"application/x-rar-compressed":            "rar",
	"application/vnd.rar":                     "rar",
	"application/gzip":                        "gz",
	"application/x-gzip":                      "gz",
	"application/x-tar":                       "tar",
	"application/x-7z-compressed":             "7z",
	"application/x-bzip2":                     "bz2",
	"application/x-xz":                        "xz",
	"application/x-lzip":                      "lz",
	"application/x-lzma":                      "lzma",
	"application/x-compress":                  "Z",
	"application/zstd":                        "zst",
	"application/x-zstd":                      "zst",
	"application/x-lz4":                       "lz4",
	"application/x-iso9660-image":             "iso",
	"application/vnd.android.package-archive": "apk",
	"application/x-apple-diskimage":           "dmg",
	"application/x-debian-package":            "deb",
	"application/x-rpm":                       "rpm",

	// ── Document - Executables & Software ──────────────────────────────
	"application/x-msdownload":                      "exe",
	"application/x-msdos-program":                   "exe",
	"application/x-executable":                      "exe",
	"application/vnd.microsoft.portable-executable": "exe",
	"application/x-msi":                             "msi",
	"application/x-ms-installer":                    "msi",
	"application/x-dosexec":                         "exe",
	"application/x-elf":                             "elf",
	"application/x-sharedlib":                       "so",
	"application/x-mach-binary":                     "macho",
	"application/java-archive":                      "jar",
	"application/x-java-archive":                    "jar",
	"application/vnd.snap":                          "snap",
	"application/x-flatpak":                         "flatpak",
	"application/x-appimage":                        "appimage",
	"application/x-ms-shortcut":                     "lnk",
	"application/x-bat":                             "bat",
	"application/x-sh":                              "sh",
	"application/x-shellscript":                     "sh",
	"application/x-csh":                             "csh",
	"application/x-powershell":                      "ps1",

	// ── Document - Source Code & Development ────────────────────────────
	"text/x-python":            "py",
	"text/x-python3":           "py",
	"text/x-java-source":       "java",
	"text/x-csrc":              "c",
	"text/x-chdr":              "h",
	"text/x-c++src":            "cpp",
	"text/x-c++hdr":            "hpp",
	"text/x-csharp":            "cs",
	"text/x-go":                "go",
	"text/x-rust":              "rs",
	"text/x-ruby":              "rb",
	"text/x-perl":              "pl",
	"text/x-php":               "php",
	"text/x-lua":               "lua",
	"text/x-scala":             "scala",
	"text/x-kotlin":            "kt",
	"text/x-swift":             "swift",
	"text/x-objectivec":        "m",
	"text/x-haskell":           "hs",
	"text/x-erlang":            "erl",
	"text/x-clojure":           "clj",
	"text/x-diff":              "diff",
	"text/x-patch":             "patch",
	"text/x-makefile":          "makefile",
	"text/x-cmake":             "cmake",
	"text/x-dockerfile":        "dockerfile",
	"text/x-shellscript":       "sh",
	"text/x-sql":               "sql",
	"text/x-r":                 "r",
	"text/x-matlab":            "m",
	"text/x-fortran":           "f90",
	"text/x-assembly":          "asm",
	"application/javascript":   "js",
	"text/javascript":          "js",
	"application/typescript":   "ts",
	"text/typescript":          "ts",
	"application/x-typescript": "ts",
	"text/jsx":                 "jsx",
	"text/tsx":                 "tsx",
	"text/x-sass":              "sass",
	"text/x-scss":              "scss",
	"text/x-less":              "less",
	"text/css":                 "css",
	"application/wasm":         "wasm",
	"application/x-httpd-php":  "php",

	// ── Document - Data & Config ───────────────────────────────────────
	"application/json":        "json",
	"application/ld+json":     "jsonld",
	"application/xml":         "xml",
	"application/x-yaml":      "yaml",
	"application/yaml":        "yaml",
	"text/yaml":               "yaml",
	"text/x-yaml":             "yaml",
	"application/toml":        "toml",
	"application/x-protobuf":  "pb",
	"application/graphql":     "graphql",
	"application/x-ndjson":    "ndjson",
	"application/sql":         "sql",
	"application/x-sql":       "sql",
	"application/x-sqlite3":   "sqlite",
	"application/vnd.sqlite3": "sqlite",

	// ── Text & Markup ──────────────────────────────────────────────────
	"text/plain":                "txt",
	"text/csv":                  "csv",
	"text/tab-separated-values": "tsv",
	"text/html":                 "html",
	"text/xml":                  "xml",
	"text/vcard":                "vcf",
	"text/x-vcard":              "vcf",
	"text/calendar":             "ics",
	"text/markdown":             "md",
	"text/x-markdown":           "md",
	"text/x-rst":                "rst",
	"text/x-log":                "log",
	"text/x-ini":                "ini",
	"text/x-properties":         "properties",
	"text/rtf":                  "rtf",
	"application/xhtml+xml":     "xhtml",
	"application/atom+xml":      "atom",
	"application/rss+xml":       "rss",
	"application/mathml+xml":    "mml",

	// ── eBook & Publishing ─────────────────────────────────────────────
	"application/epub+zip":           "epub",
	"application/x-mobipocket-ebook": "mobi",
	"application/vnd.amazon.ebook":   "azw",
	"application/x-fictionbook+xml":  "fb2",
	"application/vnd.comicbook+zip":  "cbz",
	"application/vnd.comicbook-rar":  "cbr",
	"application/x-tex":              "tex",
	"application/x-latex":            "latex",
	"application/x-bibtex":           "bib",
	"application/postscript":         "ps",
	"application/x-dvi":              "dvi",

	// ── Professional Graphics & Design ────────────────────────────────
	"application/illustrator":                     "ai",
	"application/vnd.adobe.illustrator":           "ai",
	"application/x-illustrator":                   "ai",
	"image/x-eps":                                 "eps",
	"application/eps":                             "eps",
	"application/x-eps":                           "eps",
	"application/x-indesign":                      "indd",
	"application/vnd.adobe.indesign-idml-package": "idml",
	"application/x-coreldraw":                     "cdr",
	"application/vnd.corel-draw":                  "cdr",
	"application/cdr":                             "cdr",
	"application/x-cdr":                           "cdr",
	"application/x-sketch":                        "sketch",
	"application/vnd.sketch":                      "sketch",
	"application/x-figma":                         "fig",
	"application/vnd.figma":                       "fig",
	"application/x-affinity-designer":             "afdesign",
	"application/x-affinity-photo":                "afphoto",
	"application/x-affinity-publisher":            "afpub",
	"image/x-gimp-gbr":                            "gbr",
	"image/x-gimp-pat":                            "pat",
	"application/x-gimp":                          "xcf",
	"image/x-psd":                                 "psd",
	"image/vnd.zbrush.pcx":                        "pcx",
	"image/x-pcx":                                 "pcx",
	"image/x-dds":                                 "dds",
	"image/vnd.ms-dds":                            "dds",
	"image/x-exr":                                 "exr",
	"image/vnd.radiance":                          "hdr",
	"image/x-hdr":                                 "hdr",
	"image/vnd.fpx":                               "fpx",
	"image/x-xpixmap":                             "xpm",
	"image/x-xbitmap":                             "xbm",
	"image/x-xwindowdump":                         "xwd",
	"image/x-cmu-raster":                          "ras",
	"image/x-sun-raster":                          "ras",
	"image/x-freehand":                            "fh",
	"image/x-adobe-dng":                           "dng",
	"application/vnd.ms-paint":                    "msp",

	// ── Camera RAW ────────────────────────────────────────────────────
	"image/x-canon-cr2":      "cr2",
	"image/x-canon-cr3":      "cr3",
	"image/x-canon-crw":      "crw",
	"image/x-nikon-nef":      "nef",
	"image/x-nikon-nrw":      "nrw",
	"image/x-sony-arw":       "arw",
	"image/x-sony-sr2":       "sr2",
	"image/x-sony-srf":       "srf",
	"image/x-fuji-raf":       "raf",
	"image/x-olympus-orf":    "orf",
	"image/x-panasonic-rw2":  "rw2",
	"image/x-panasonic-raw":  "raw",
	"image/x-pentax-pef":     "pef",
	"image/x-samsung-srw":    "srw",
	"image/x-sigma-x3f":      "x3f",
	"image/x-leica-rwl":      "rwl",
	"image/x-hasselblad-3fr": "3fr",
	"image/x-kodak-dcr":      "dcr",
	"image/x-kodak-kdc":      "kdc",
	"image/x-minolta-mrw":    "mrw",
	"image/x-epson-erf":      "erf",
	"image/x-mamiya-mef":     "mef",
	"image/x-phaseone-iiq":   "iiq",

	// ── CAD, 3D & Design ──────────────────────────────────────────────
	"model/stl":                    "stl",
	"model/obj":                    "obj",
	"model/gltf+json":              "gltf",
	"model/gltf-binary":            "glb",
	"model/vnd.collada+xml":        "dae",
	"model/step":                   "step",
	"model/iges":                   "iges",
	"model/3mf":                    "3mf",
	"model/fbx":                    "fbx",
	"application/x-blender":        "blend",
	"image/x-3ds":                  "3ds",
	"application/x-autocad":        "dwg",
	"image/vnd.dxf":                "dxf",
	"application/vnd.sketchup.skp": "skp",
	"application/x-cinema4d":       "c4d",
	"application/x-maya":           "mb",
	"application/x-houdini":        "hip",
	"application/x-substance":      "sbs",

	// ── GIS & Maps ─────────────────────────────────────────────────────
	"application/geo+json":                 "geojson",
	"application/vnd.google-earth.kml+xml": "kml",
	"application/vnd.google-earth.kmz":     "kmz",
	"application/gpx+xml":                  "gpx",
	"application/x-shapefile":              "shp",

	// ── Crypto & Security ──────────────────────────────────────────────
	"application/pkix-cert":            "cer",
	"application/x-x509-ca-cert":       "crt",
	"application/x-pem-file":           "pem",
	"application/pkcs8":                "p8",
	"application/pkcs12":               "p12",
	"application/x-pkcs12":             "pfx",
	"application/x-pkcs7-certificates": "p7b",
	"application/pgp-keys":             "asc",
	"application/pgp-signature":        "sig",
	"application/pgp-encrypted":        "pgp",

	// ── Font ───────────────────────────────────────────────────────────
	"font/ttf":               "ttf",
	"font/otf":               "otf",
	"font/woff":              "woff",
	"font/woff2":             "woff2",
	"application/font-ttf":   "ttf",
	"application/font-otf":   "otf",
	"application/font-woff":  "woff",
	"application/font-woff2": "woff2",
	"application/x-font-ttf": "ttf",
	"application/x-font-otf": "otf",

	// ── Database & Backup ──────────────────────────────────────────────
	"application/x-mysql":    "sql",
	"application/x-pgsql":    "sql",
	"application/vnd.dbf":    "dbf",
	"application/x-msaccess": "mdb",
	"application/x-hdf5":     "hdf5",
	"application/x-parquet":  "parquet",
	"application/x-avro":     "avro",

	// ── Misc ───────────────────────────────────────────────────────────
	"application/x-shockwave-flash":       "swf",
	"application/x-silverlight-app":       "xap",
	"application/vnd.apple.installer+xml": "dist",
	"application/x-cpio":                  "cpio",
	"application/x-shar":                  "shar",
	"application/vnd.tcpdump.pcap":        "pcap",
	"application/x-bittorrent":            "torrent",
	"application/x-nintendo-nes-rom":      "nes",
	"application/x-genesis-rom":           "gen",
	"message/rfc822":                      "eml",
	"application/mbox":                    "mbox",
	"application/vnd.ms-outlook":          "msg",

	// ── Fallback binary ────────────────────────────────────────────────
	"application/octet-stream": "bin",
}

// normalizeMIME strips parameters (e.g., "; codecs=opus") and lowercases the MIME type.
func normalizeMIME(contentType string) string {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	if idx := strings.Index(ct, ";"); idx != -1 {
		ct = strings.TrimSpace(ct[:idx])
	}
	return ct
}

func (u *S3Uploader) getExtensionFromContentType(contentType string, mediaType string) string {
	// 1. Try the raw content type first (handles "audio/ogg; codecs=opus" → check full string)
	if ext, ok := mimeExtensionMap[strings.ToLower(strings.TrimSpace(contentType))]; ok {
		return ext
	}

	// 2. Normalize (strip params like "; codecs=opus") and try again
	normalized := normalizeMIME(contentType)
	if ext, ok := mimeExtensionMap[normalized]; ok {
		return ext
	}

	// 3. Try Go stdlib as fallback (OS-dependent but covers exotic types)
	if exts, err := mime.ExtensionsByType(contentType); err == nil && len(exts) > 0 {
		ext := strings.TrimPrefix(exts[0], ".")
		// Avoid known bad mappings from stdlib
		if ext != "jfif" && ext != "jpe" {
			return ext
		}
		// If stdlib returned a bad mapping, try to find a better one
		for _, e := range exts {
			e = strings.TrimPrefix(e, ".")
			if e == "jpg" || e == "jpeg" || e == "png" || e == "mp4" || e == "mp3" {
				return e
			}
		}
		return ext
	}

	// 4. Extract subtype from MIME (e.g., "video/x-matroska" → "x-matroska")
	if normalized != "" {
		parts := strings.SplitN(normalized, "/", 2)
		if len(parts) == 2 {
			subtype := parts[1]
			// Strip vendor prefix for cleaner extensions
			subtype = strings.TrimPrefix(subtype, "x-")
			subtype = strings.TrimPrefix(subtype, "vnd.")
			if subtype != "" && subtype != "plain" && subtype != "octet-stream" && len(subtype) <= 10 {
				return subtype
			}
		}
	}

	// 5. Final fallback based on WhatsApp media type
	switch mediaType {
	case "image":
		return "jpg"
	case "video":
		return "mp4"
	case "audio", "voice":
		return "ogg"
	case "document":
		return "bin"
	case "sticker":
		return "webp"
	default:
		return "bin"
	}
}

func classifyS3Error(err error) string {
	errStr := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errStr, "timeout"):
		return "timeout"
	case strings.Contains(errStr, "connection"):
		return "connection"
	case strings.Contains(errStr, "access denied"):
		return "access_denied"
	case strings.Contains(errStr, "no such bucket"):
		return "bucket_not_found"
	case strings.Contains(errStr, "file too large"):
		return "file_too_large"
	case strings.Contains(errStr, "network"):
		return "network"
	default:
		return "unknown"
	}
}

func (u *S3Uploader) ObjectExists(ctx context.Context, key string) (bool, error) {
	_, err := u.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (u *S3Uploader) GetObjectSize(ctx context.Context, key string) (int64, error) {
	resp, err := u.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return 0, err
	}

	return *resp.ContentLength, nil
}
