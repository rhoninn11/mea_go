package internal

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

func GetGlobState() *State {
	return &globState
}

func init() {
	globState = State{
		keys:      []string{"x", "y", "z"},
		lastUsage: 0,
		history:   make([]string, 0, 16),
	}
}

func SetContentType[ValT ~string](w http.ResponseWriter, val ValT) {
	w.Header().Set(HContentType, string(val))
}

func (s *State) AxisFn(w http.ResponseWriter, r *http.Request) {
	dir := mojUkladOdniesienia[s.keys[s.lastUsage]]
	text := dir.LatentVector

	SetContentType(w, ContentType_Html)
	historyNote := fmt.Sprintf("%s |może data|", text)
	entry := components.Entry(historyNote)
	glob := PageWithSidebar(entry)
	glob.Render(r.Context(), w)

	s.lastUsage += 1
	s.history = append(s.history, historyNote)
	if s.lastUsage == len(s.keys) {
		s.lastUsage = 0
	}
}

func (s *State) HistoryFn(w http.ResponseWriter, r *http.Request) {
	SetContentType(w, ContentType_Html)
	history := components.HistoryWhole(s.history)
	page := PageWithSidebar(history)
	page.Render(context.Background(), w)
}

func (s *State) LoadingPage(w http.ResponseWriter, r *http.Request) {
	SetContentType(w, ContentType_Html)
	render := components.SectionWithLoading()
	render = components.Global("Loading Page", render)
	render.Render(context.Background(), w)
}

func LoadingTest(w http.ResponseWriter, r *http.Request) {
	SetContentType(w, ContentType_Html)
	for i := range 10 {
		render := components.Block(i)
		render.Render(context.Background(), w)
		w.(http.Flusher).Flush()
		time.Sleep(10 * time.Millisecond)
	}
}
