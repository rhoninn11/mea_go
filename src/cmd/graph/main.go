package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
)

const X_DIM uint16 = 8
const Y_DIM uint16 = 8
const scale uint16 = 128

var matrix []opts.Chart3DData

func init() {
	matrix = make([]opts.Chart3DData, 0, X_DIM*Y_DIM)
	for yy := range Y_DIM {
		for x := range X_DIM {
			matrix = append(matrix, opts.Chart3DData{
				Value: []interface{}{x * scale, yy * scale, (x + yy) * scale},
			})
		}
	}
	fmt.Printf("matrix values set\n")
}

func basciChart() *charts.Bar3D {
	var chart = charts.NewBar3D()
	chart.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "basic 3D pilars"}),
		charts.WithGrid3DOpts(opts.Grid3D{BoxWidth: float32(scale), BoxHeight: float32(scale), BoxDepth: float32(scale)}),
	)
	chart.AddSeries("test", matrix)
	chart.SetSeriesOptions(
		charts.WithBar3DChartOpts(opts.Bar3DChart{Shading: "lambert"}),
		charts.WithItemStyleOpts(opts.ItemStyle{
			BorderColor: "#fff",
			BorderWidth: 1,
		}),
	)

	return chart
}

func main() {
	chart := basciChart()
	page := components.NewPage()
	page.AddCharts(chart)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		page.Render(w)
	})
	fmt.Printf("listning on port 19000")
	log.Fatal(http.ListenAndServe(":19000", nil))
}
