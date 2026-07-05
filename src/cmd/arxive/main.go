package main

import (
	"fmt"
	"mea_go/src/internal"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/a-h/templ"
)

type PdfViewer struct {
	actualPage int
	lastPage   int
	mutex      sync.RWMutex
}

type HtmxId = internal.HtmxId

var pageTurn = internal.LinkBind{
	EntryPoint: "/turn/{dir}",
	FmtStr:     "/turn/%s",
}
var pageRefresh = internal.LinkBind{
	EntryPoint: "/show",
	FmtStr:     "/show",
}

var actionSink HtmxId = internal.NamedHid("action_sink")
var reader HtmxId = internal.NamedHid("reader")
var pageSpot HtmxId = internal.NamedHid("page_spot")

func (pv *PdfViewer) PageRefresh(w http.ResponseWriter, r *http.Request) {
	pv.page().Render(r.Context(), w)
}

func (pv *PdfViewer) page() templ.Component {
	pagelink := fmt.Sprintf("/pages/strona%d.png", pv.actualPage+1)
	pdfPage := internal.PageImgS(pagelink)
	return internal.HidWrap(pageSpot, pdfPage)
}

func (pv *PdfViewer) ShowPage(w http.ResponseWriter, r *http.Request) {

	info := internal.LightTextEntry("we will be showing pdf's here")
	prev := internal.LightTextEntry("previous")
	next := internal.LightTextEntry("next")

	prevBtn := internal.ButtonAction(internal.ActionLink{
		Target:       actionSink.TargName,
		LinkToAction: pageTurn.Format("prev"),
		Trigger:      "keyup[key=='q'] from:body",
	}, prev)

	nextBtn := internal.ButtonAction(internal.ActionLink{
		Target:       actionSink.TargName,
		LinkToAction: pageTurn.Format("next"),
		Trigger:      "keyup[key=='e'] from:body",
	}, next)

	content := internal.Col([]templ.Component{
		info, pv.page(),
		internal.RowVar(prevBtn, internal.Space(), nextBtn),
	}, reader.JustName)

	var opts = internal.PageOpts{
		PageContent: content,
		Sinks:       []HtmxId{actionSink},
	}
	page := internal.PageWithSidebar(opts)
	page.Render(r.Context(), w)
}

func (pv *PdfViewer) TurnPage(w http.ResponseWriter, r *http.Request) {
	pv.mutex.Lock()
	defer pv.mutex.Unlock()

	val := r.PathValue("dir")
	switch val {
	case "next":
		if pv.actualPage < pv.lastPage {
			pv.actualPage += 1
		}
	case "prev":
		if pv.actualPage > 0 {
			pv.actualPage -= 1
		}
	}

	step := internal.ProcedeNextVisible(internal.TargetAction{
		Target:       pageSpot.TargName,
		LinkToAction: pageRefresh.EntryPoint,
	})
	internal.HidWrap(actionSink, step).Render(r.Context(), w)
}

func main() {
	msg := "trying to render pdf here"
	fmt.Printf("hello world %s\n", internal.ColoredText(msg))
	links := []string{
		"https://arxiv.org/pdf/2510.25741",
		"https://arxiv.org/pdf/2512.06818",
	}
	var pageCount int64
	var renderDst string
	for _, link := range links {
		id := strings.TrimPrefix(link, "https://arxiv.org/pdf/")
		renderDst = internal.DirGuard(path.Join("fs", id))
		pageCount = processLink(link, renderDst)
	}

	hostPdf(pageCount, renderDst)
}

func hostPdf(total int64, renderSrc string) {
	addr := "localhost:8080"
	var pv = PdfViewer{
		lastPage: int(total - 1),
	}

	static := http.FileServer(http.Dir("static/"))
	static = http.StripPrefix("/static/", static)
	static = internal.NoCacheMiddleware(static)
	http.Handle("/static/", static)

	pages := http.FileServer(http.Dir(renderSrc))
	pages = http.StripPrefix("/pages/", pages)
	http.Handle("/pages/", pages)

	http.HandleFunc("/", pv.ShowPage)
	http.HandleFunc(pageTurn.EntryPoint, pv.TurnPage)
	http.HandleFunc(pageRefresh.EntryPoint, pv.PageRefresh)

	fmt.Printf("+++ listening on %s\n", internal.ColoredText2(addr))
	http.ListenAndServe(addr, nil)
}
