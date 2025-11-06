package internal

const (
	HContentType  = "Content-Type"
	HCacheControl = "Cache-Control"
)

type ContentType string

const (
	ContentType_PlainText   ContentType = "text/plain"
	ContentType_Html        ContentType = "text/html"
	ContentType_EventStream ContentType = "text/event-stream"
	ContentType_Png         ContentType = "image/png"
)
