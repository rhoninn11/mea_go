package internal

import (
	"fmt"
	"mea_go/components"
	"net/http"

	"github.com/a-h/templ"
)

type ResponseWriter = http.ResponseWriter
type Request = http.Request
type HttpFunc = func(ResponseWriter, *Request)

var endpotins = make([]templ.SafeURL, 0, 16)

func RegisterHandler(endpoint templ.SafeURL, fn HttpFunc) {
	middle := func(w ResponseWriter, r *Request) {
		fmt.Println("+++ called ", endpoint, "+++")
		fn(w, r)
	}
	endpotins = append(endpotins, endpoint)
	http.HandleFunc(string(endpoint), middle)
}

func PageWithSidebar(main templ.Component) templ.Component {
	side := components.SideLinks(endpotins)
	twoTabs := components.TwoTabs(side, main)
	return components.Global("Tua editro", twoTabs)
}
