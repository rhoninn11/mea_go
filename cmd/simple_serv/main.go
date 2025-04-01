package main

import (
	"fmt"
	"net/http"
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
var keys = []string{"x", "y", "z"}
var lastUsage = 0

type RespWriter = http.ResponseWriter
type Req = http.Request

func fnHandler(w RespWriter, r *Req) {
	dir := mojUkladOdniesienia[keys[lastUsage]]
	msg := dir.LatentVector

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "%s", msg)
	lastUsage += 1
	if lastUsage == len(keys) {
		lastUsage = 0
	}
	fmt.Println(spf("+++ last idx: %d", lastUsage))
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

	const fnName = "axis"
	http.HandleFunc(spf("/%s", fnName), fnHandler)

	var url = spf("http://%s/%s", base, fnName)
	fmt.Printf("+++ niby wystartowałem api, api route: \n%s\n", url)
	http.ListenAndServe(base, nil)
}
