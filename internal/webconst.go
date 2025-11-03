package internal

const (
	HContentType  = "Content-Type"
	HCacheControl = "Cache-Control"
)

type ContentType string

const (
	ContentTypePlainText   ContentType = "text/plain"
	ContentTypeHtml        ContentType = "text/html"
	ContentTypeEventStream ContentType = "text/event-stream"
	ContentTypePng         ContentType = "image/png"
)
