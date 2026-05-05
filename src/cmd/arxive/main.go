package main

import (
	"fmt"
	"io"
	"mea_go/src/internal"
	"net/http"
	"os"
	"sync"
)

type PdfViewer struct {
	actualPage int
}

type HtmxId = internal.HtmxId

func (pv *PdfViewer) ShowPage(w http.ResponseWriter, r *http.Request) {

	mainContent := internal.Entry("we will be showing pdf's here")
	imgContent := internal.PageImg("/pages/strona1.png")
	_ = mainContent

	var opts = internal.PageOpts{
		PageContent: imgContent,
		Sinks:       []HtmxId{},
	}
	page := internal.PageWithSidebar(opts)
	page.Render(r.Context(), w)
}

func main() {
	msg := "trying to render pdf here"
	fmt.Printf("hello world %s\n", internal.ColoredText(msg))
	link := "https://arxiv.org/pdf/2510.25741"
	checkoutThisOneAlso := "https://arxiv.org/pdf/2512.06818"
	_ = checkoutThisOneAlso
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
	_ = total
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
	var pv PdfViewer

	static := http.FileServer(http.Dir("static/"))
	static = http.StripPrefix("/static/", static)
	static = internal.NoCacheMiddleware(static)
	http.Handle("/static/", static)

	pages := http.FileServer(http.Dir(renderDst))
	pages = http.StripPrefix("/pages/", pages)
	http.Handle("/pages/", pages)

	http.HandleFunc("/", pv.ShowPage)
	fmt.Printf("+++ listening on %s\n", internal.ColoredText(addr))
	http.ListenAndServe(addr, nil)
}
