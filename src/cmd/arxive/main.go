package main

import (
	"fmt"
	"io"
	"mea_go/src/internal"
	"net/http"
	"os"
	"sync"

	"github.com/a-h/templ"
)

type PdfViewer struct {
	actualPage int
	maxPage    int
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
	fmt.Printf("+++ page refresh hit")
	pv.page().Render(r.Context(), w)
}

func (pv *PdfViewer) page() templ.Component {
	pagelink := fmt.Sprintf("/pages/strona%d.png", pv.actualPage+1)
	pdfPage := internal.PageImg(pagelink)
	return internal.HidWrap(pageSpot, pdfPage)
}

func (pv *PdfViewer) ShowPage(w http.ResponseWriter, r *http.Request) {

	info := internal.LightTextEntry("we will be showing pdf's here")
	prev := internal.LightTextEntry("previous")
	next := internal.LightTextEntry("next")

	prevBtn := internal.ButtonAction(internal.ActionLink{
		Target:       actionSink.TargName,
		LinkToAction: pageTurn.Format("prev"),
	}, prev)

	nextBtn := internal.ButtonAction(internal.ActionLink{
		Target:       actionSink.TargName,
		LinkToAction: pageTurn.Format("next"),
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
		fmt.Printf("turn to next page\n")
		if pv.actualPage < pv.maxPage {
			pv.actualPage += 1
		}
	case "prev":
		fmt.Printf("turn to prev page\n")
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
	link := "https://arxiv.org/pdf/2510.25741"
	checkoutThisOneAlso := "https://arxiv.org/pdf/2512.06818"
	link = checkoutThisOneAlso
	data, err := http.Get(link)
	if err != nil {
		fmt.Println(internal.ColoredText("request failed"))
		os.Exit(1)
	}

	pdfile := "fs/arxiv.pdf"
	renderDst := internal.DirGuard("fs/pdfrenders")

	f, err := os.Create(pdfile)
	if err != nil {
		fmt.Printf("%s\n", internal.ColoredText("file create failed"))
		os.Exit(1)
	}

	total, err := io.Copy(f, data.Body)
	if err != nil {
		fmt.Printf("%s\n", internal.ColoredText("resp to file failed"))
		os.Exit(1)
	}
	pageCount, err := internal.CountPages(pdfile)
	if err != nil {
		fmt.Printf("%s | %s\n", internal.ColoredText("count failed"), err.Error())
		os.Exit(1)
	}
	var wg sync.WaitGroup
	for i := range pageCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := internal.RenderPdf(pdfile, renderDst, i); err != nil {
				fmt.Printf("%s | %s\n", internal.ColoredText("render failed"), err.Error())
				os.Exit(1)
			}
		}()
	}
	wg.Wait()
	fmt.Printf("+++ pages renderd\n")

	addr := "localhost:8080"
	var pv = PdfViewer{
		maxPage: int(total),
	}

	static := http.FileServer(http.Dir("static/"))
	static = http.StripPrefix("/static/", static)
	static = internal.NoCacheMiddleware(static)
	http.Handle("/static/", static)

	pages := http.FileServer(http.Dir(renderDst))
	pages = http.StripPrefix("/pages/", pages)
	http.Handle("/pages/", pages)

	http.HandleFunc("/", pv.ShowPage)
	http.HandleFunc(pageTurn.EntryPoint, pv.TurnPage)
	http.HandleFunc(pageRefresh.EntryPoint, pv.PageRefresh)

	fmt.Printf("+++ listening on %s\n", internal.ColoredText(addr))
	http.ListenAndServe(addr, nil)
}
