package main

import (
	"context"
	"fmt"
	"log"
	"mea_go/src/components"
	"mea_go/src/internal"
	"net/http"
)

func RecivePrompt(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Fatal("empty form")
	}

	form := r.Form
	fmt.Printf("+++ form len: %d\n", len(form))
	for k, v := range form {
		fmt.Println("+++", k, v)
	}

	internal.SetContentType(w, internal.ContentType_Html)
	render := components.Block(0)
	render.Render(context.Background(), w)
}

// func noCacheMiddleware(base internal.HttpFunc) internal.HttpFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		w.Header(internal.HCacheControl, no-cache)
// 		base(w, r)
// 	}
// }

const (
	DEBUG = "DEBUG"
	PROD  = "PROD"
)

func main() {
	var mode = DEBUG
	_ = mode

	const host = "0.0.0.0"
	const port = 8080
	var base = fmt.Sprintf("%s:%d", host, port) // eg localhost:8080

	static := http.FileServer(http.Dir("./static/"))
	static = http.StripPrefix("/static/", static)
	static = internal.NoCacheMiddleware(static)
	http.Handle("/static/", static)

	deeper := internal.PromptModuleAccess()
	register := internal.RegisterHandler

	sampleState := internal.GetGlobState()
	register("/axis", sampleState.AxisFn)
	register("/history", sampleState.HistoryFn)
	register("/loading", sampleState.LoadingPage)

	mapping := deeper.LoadFns()
	for k, v := range mapping {
		register(k, v)
	}

	var url = fmt.Sprintf("http://%s/%s", base, "history")
	fmt.Printf("+++ niby wystartowałem api, api route: \n%s\n", url)
	_ = http.ListenAndServe(base, nil)
}
