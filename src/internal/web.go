package internal

import (
	"fmt"
	"mea_go/src/components"
	"net/http"

	"github.com/a-h/templ"
)

type ResponseWriter = http.ResponseWriter
type Request = http.Request
type HttpFunc = func(ResponseWriter, *Request)

var endpotins = make([]templ.SafeURL, 0, 16)

func RegisterHandler(endpoint templ.SafeURL, fnPack HttpFuncPack) {
	middle := func(w ResponseWriter, r *Request) {
		if fnPack.Show {
			fmt.Println("+++ called ", endpoint, "+++")
		}
		fnPack.Fn(w, r)
	}

	if fnPack.Show {
		endpotins = append(endpotins, endpoint)
	}

	http.HandleFunc(string(endpoint), middle)
}

func PageWithSidebar(main templ.Component) templ.Component {
	side := components.SideLinks(endpotins)
	twoTabs := components.TwoTabs(side, main)
	return components.Global("Tua editro", twoTabs)
}

func NoCacheMiddleware(base http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		SetCacheControl(w, CacheType_NoCache)
		base.ServeHTTP(w, r)
	})
}
