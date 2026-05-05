package internal

import (
	"fmt"
	"net/http"
)

type HeaderType string

const (
	HContentType  HeaderType = "Content-Type"
	HCacheControl HeaderType = "Cache-Control"
)

type ContentType string

const (
	ContentType_PlainText   ContentType = "text/plain"
	ContentType_Html        ContentType = "text/html"
	ContentType_EventStream ContentType = "text/event-stream"
	ContentType_Png         ContentType = "image/png"
)

type CacheType string
type HtmxId struct {
	JustName string
	TargName string
}

func NamedHid(name string) HtmxId {
	return HtmxId{
		JustName: name,
		TargName: fmt.Sprintf("#%s", name),
	}
}

type LinkBind struct {
	EntryPoint string
	FmtStr     string
}

func (lb *LinkBind) Format(a ...any) string {
	return fmt.Sprintf(lb.FmtStr, a...)
}

const (
	CacheType_NoCache CacheType = "no-cache"
)

func SetContentType(w http.ResponseWriter, val ContentType) {
	w.Header().Set(string(HContentType), string(val))
}

func SetCacheControl(w http.ResponseWriter, val CacheType) {
	w.Header().Set(string(HCacheControl), string(val))
}
