package main

import (
	"encoding/xml"
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
	Chars []Char `xml:"font"`
}

func (l *Line) read() string {
	letters := make([]string, len(l.Chars))
	for i := range letters {
		letters[i] = l.Chars[i].C
	}
	return strings.Join(letters, "")
}

type Char struct {
	BBox string `xml:"bbox,attr"`
	C    string `xml:"c,attr"`
}

func parse(file string) *SText {
	data, err := os.ReadFile(file)
	if err != nil {
		log.Fatalf("file open failed %s | %s", file, err.Error())
	}
	var here SText
	if err := xml.Unmarshal(data, &here); err != nil {
		log.Fatal("filed to unmarshal %s file | %s", file, err.Error())
	}
	return &here
}
