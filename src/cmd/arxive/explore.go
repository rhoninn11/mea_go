package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"strings"
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
