package main

import (
	"fmt"
	"io"
	"mea_go/src/internal"
	"net/http"
	"os"
)

func main() {
	msg := "trying to render pdf here"
	fmt.Printf("hello world %s\n", internal.ColoredText(msg))
	link := "https://arxiv.org/pdf/2510.25741"
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

	if err := internal.RenderPdf(pdfile, renderDst, 0); err != nil {
		fmt.Printf("%s | %s\n", internal.ColoredText("render failed"), err.Error())
		os.Exit(1)
	}

	if err := internal.CountPages(pdfile); err != nil {
		fmt.Printf("%s | %s\n", internal.ColoredText("count failed"), err.Error())
		os.Exit(1)
	}
}
