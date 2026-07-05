package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"mea_go/src/internal"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
)

type SText struct {
	Pages []Page `xml:"page"`
}

type Page struct {
	Width  float64 `xml:"width,attr"`
	Height float64 `xml:"height,attr"`
	Blocks []Block `xml:"block"`
}

type Block struct {
	BBox  string `xml:"bbox,attr"`
	Lines []Line `xml:"line"`
}

type Line struct {
	BBox  string `xml:"bbox,attr"`
	Fonts []Font `xml:"font"`
}

func (l *Line) read() string {
	perFont := make([]string, len(l.Fonts))
	for jj, font := range l.Fonts {
		letters := make([]string, len(font.Chars))
		for i := range letters {
			letters[i] = font.Chars[i].C
		}
		perFont[jj] = strings.Join(letters, "")
	}
	return strings.Join(perFont, "|")
}

type Font struct {
	Size  string `xml:"size,attr"`
	Chars []Char `xml:"char"`
}

type Char struct {
	BBox string `xml:"quad,attr"`
	C    string `xml:"c,attr"`
}

func loadDocument(file string) *SText {
	data, err := os.ReadFile(file)
	if err != nil {
		log.Fatalf("file open failed %s | %s", file, err.Error())
	}
	var root SText
	if err := xml.Unmarshal(data, &root); err != nil {
		log.Fatal("filed to unmarshal %s file | %s", file, err.Error())
	}
	return &root
}

func readDocument(doc *SText) {
	pageCount := len(doc.Pages)
	fmt.Printf("+++ page_sections num %d", pageCount)
	for i, page := range doc.Pages {
		fmt.Printf("+++ page_section(%d) has %d blocks\n", i, len(page.Blocks))
		for jj, block := range page.Blocks {
			fmt.Printf("+++ 	block(%d) has %d lines\n", jj, len(block.Lines))
			for kkk, line := range block.Lines {
				fmt.Printf("+++       line(%d): %s\n", kkk, line.read())
			}
		}
	}
}

func fetchLink(link string, dst string) (string, error) {
	pdfile := path.Join(dst, "paper.pdf")
	if _, err := os.Stat(pdfile); err == nil {
		fmt.Printf("file (%s) present\n", internal.ColoredText2(pdfile))
		return pdfile, nil
	}

	data, err := http.Get(link)
	if err != nil {
		return "", fmt.Errorf("%s | %w", internal.ColoredText("request failed"), err)
	}

	f, err := os.Create(pdfile)
	if err != nil {
		return "", fmt.Errorf("%s | %w", internal.ColoredText("file create failed"), err)
	}

	totalBytesFetched, err := io.Copy(f, data.Body)
	if err != nil {
		return "", fmt.Errorf("%s | %w", internal.ColoredText("resp to file failed"), err)
	}
	_ = totalBytesFetched
	return pdfile, nil
}

func processLink(link string, renderDst string) int64 {
	pdfile, err := fetchLink(link, "tmp")
	if err != nil {
		fmt.Printf("errorourrus %s\n", err.Error())
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
			if err := internal.XmlizePdf(pdfile, renderDst, i); err != nil {
				fmt.Printf("%s | %s\n", internal.ColoredText("xmlization failed"), err.Error())
				os.Exit(1)
			}
		}()
	}
	wg.Wait()
	fmt.Printf("+++ pages renderd\n")
	return int64(pageCount)
}
