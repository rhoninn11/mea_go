package internal

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
)

type ResponseWriter = http.ResponseWriter
type Request = http.Request
type HttpFunc = func(ResponseWriter, *Request)

type HttpFuncPack struct {
	Fn   HttpFunc
	Show bool
}

var endpotins = make([]templ.SafeURL, 0, 16)

func RegisterHandler(endpoint templ.SafeURL, fnPack HttpFuncPack) {
	middle := func(w ResponseWriter, r *Request) {
		if fnPack.Show {
			fmt.Println("+++ called ", endpoint, "+++")
		}
		fnPack.Fn(w, r)
	}

	// if fnPack.Show {
	// 	endpotins = append(endpotins, endpoint)
	// }

	http.HandleFunc(string(endpoint), middle)
}

type PageOpts struct {
	PageContent templ.Component
	Sinks       []HtmxId
}

func PageWithSidebar(data PageOpts) templ.Component {

	page := data.PageContent
	sinks := data.Sinks
	side := SideLinks(endpotins, sinks)
	twoTabs := TwoTabs(side, page)
	return Global("Tua editro", twoTabs)
}

func NoCacheMiddleware(base http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		SetCacheControl(w, CacheType_NoCache)
		base.ServeHTTP(w, r)
	})
}
