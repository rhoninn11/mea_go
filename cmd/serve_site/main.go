package main

import (
	"context"
	"fmt"
	"mea_go/components"
	"net/http"
	"time"
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

type RespWriter = http.ResponseWriter
type Req = http.Request

var htmlType = "text/html"
var textStreamType = "text/event-stream"

func (s *State) AxisFn(w RespWriter, r *Req) {
	dir := mojUkladOdniesienia[s.keys[s.lastUsage]]
	text := dir.LatentVector

	w.Header().Set("Content-Type", htmlType)
	historyNote := fmt.Sprintf("%s |może data|", text)
	entry := components.Entry(historyNote)
	glob := components.Global("Axis lottery:", entry)
	glob.Render(r.Context(), w)

	s.lastUsage += 1
	s.history = append(s.history, historyNote)
	if s.lastUsage == len(s.keys) {
		s.lastUsage = 0
	}
}

func (s *State) HistoryFn(w RespWriter, r *Req) {
	w.Header().Set("Content-Type", htmlType)
	render := components.HistoryWhole(s.history)
	render = components.Global("Adam Grzelak", render)
	render.Render(context.Background(), w)
}

func (s *State) LoadingPage(w RespWriter, r *Req) {

	w.Header().Set("Content-Type", htmlType)
	render := components.SectionWithLoading()
	render = components.Global("Loading Page", render)
	render.Render(context.Background(), w)
}

func loading(w RespWriter, r *Req) {
	w.Header().Set("Content-Type", textStreamType) // but i hope html can be pushed fine😅
	for i := range 10 {
		render := components.Block(i)
		render.Render(context.Background(), w)
		w.(http.Flusher).Flush()
		time.Sleep(10 * time.Millisecond)
	}
}

func spf(format string, a ...any) string {
	var temp = fmt.Sprintf(format, a...)
	// fmt.Println("+++Debug: ", temp)
	return temp
}

func main() {
	// for _, key := range keys {
	// 	fmt.Println(mojUkladOdniesienia[key])
	// }
	const host = "localhost"
	const port = 8080
	var base = spf("%s:%d", host, port) // eg localhost:8080

	static := http.FileServer(http.Dir("./static/"))
	http.Handle("/static/", http.StripPrefix("/static/", static))

	http.HandleFunc("/axis", globState.AxisFn)
	http.HandleFunc("/history", globState.HistoryFn)
	http.HandleFunc("/loading", loading)
	http.HandleFunc("/page/loading", globState.LoadingPage)

	var url = spf("http://%s/%s", base, "history")
	fmt.Printf("+++ niby wystartowałem api, api route: \n%s\n", url)
	http.ListenAndServe(base, nil)
}
