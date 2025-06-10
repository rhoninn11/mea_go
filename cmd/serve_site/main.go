package main

import (
	"context"
	"fmt"
	"log"
	"mea_go/components"
	"mea_go/internal"
	"net/http"
	"time"

	"github.com/a-h/templ"
)

type Direction struct {
	LatentVector string
	Power        float32
}

var mojUkladOdniesienia = map[string]Direction{
	"x": Direction{LatentVector: "Jeden z moich latent aspektów", Power: 0.3},
	"y": Direction{LatentVector: "Askpekt dominujący tego czego szukam", Power: 0.5},
	"z": Direction{LatentVector: "To takie moje oczko w głowie", Power: 0.2},
}

type State struct {
	keys      []string
	lastUsage int
	history   []string
}

var globState State

func init() {
	globState = State{
		keys:      []string{"x", "y", "z"},
		lastUsage: 0,
		history:   make([]string, 0, 16),
	}
}

var htmlType = "text/html"
var textStreamType = "text/event-stream"

func (s *State) AxisFn(w http.ResponseWriter, r *http.Request) {
	dir := mojUkladOdniesienia[s.keys[s.lastUsage]]
	text := dir.LatentVector

	w.Header().Set(internal.HContentType, htmlType)
	historyNote := fmt.Sprintf("%s |może data|", text)
	entry := components.Entry(historyNote)
	glob := internal.PageWithSidebar(entry)
	glob.Render(r.Context(), w)

	s.lastUsage += 1
	s.history = append(s.history, historyNote)
	if s.lastUsage == len(s.keys) {
		s.lastUsage = 0
	}
}

func (s *State) HistoryFn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", htmlType)
	history := components.HistoryWhole(s.history)
	page := internal.PageWithSidebar(history)
	page.Render(context.Background(), w)
}

func (s *State) LoadingPage(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", htmlType)
	render := components.SectionWithLoading()
	render = components.Global("Loading Page", render)
	render.Render(context.Background(), w)
}

// /page/gen
func (s *State) GeneratePage(w http.ResponseWriter, r *http.Request) {
	var content templ.Component
	defer func() {
		w.Header().Set("Content-Type", htmlType)
		fullPage := internal.PageWithSidebar(content)
		fullPage.Render(context.Background(), w)
	}()

	content = internal.PromptEditor("unique-id")
}

func (s *State) RecivePrompt(w http.ResponseWriter, r *http.Request) {
	fmt.Println("+++wywołano well")

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

	w.Header().Set("Content-Type", htmlType)
	render := components.Block(0)
	render.Render(context.Background(), w)
}

func loading(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", textStreamType) // but i hope html can be pushed fine😅
	for i := range 10 {
		render := components.Block(i)
		render.Render(context.Background(), w)
		w.(http.Flusher).Flush()
		time.Sleep(10 * time.Millisecond)
	}
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
	// for _, key := range keys {
	// 	fmt.Println(mojUkladOdniesienia[key])
	// }
	// const host = "localhost"
	const host = "0.0.0.0"
	const port = 8080
	var base = fmt.Sprintf("%s:%d", host, port) // eg localhost:8080

	static := http.FileServer(http.Dir("./static/"))
	// if mode == DEBUG {
	// 	static = noCacheMiddleware(static)
	// }

	// TODO: for static files rebuild in development would be nice to set
	// cache-control header to no cache somehow in dev mode
	http.Handle("/static/", http.StripPrefix("/static/", static))

	deeper := internal.PromptModuleAccess()
	register := internal.RegisterHandler

	register("/axis", globState.AxisFn)
	register("/history", globState.HistoryFn)
	register("/loading", loading)
	register("/page/loading", globState.LoadingPage)
	register("/page/gen", globState.GeneratePage)
	register("/well", globState.RecivePrompt)
	mapping := deeper.LoadFns()
	for k, v := range mapping {
		register(k, v)
	}

	var url = fmt.Sprintf("http://%s/%s", base, "history")
	fmt.Printf("+++ niby wystartowałem api, api route: \n%s\n", url)
	_ = http.ListenAndServe(base, nil)
}
