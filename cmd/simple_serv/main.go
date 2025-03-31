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

func ApiHandler(w RespWriter, r *Req) {
	dir := mojUkladOdniesienia[keys[lastUsage]]
	msg := dir.LatentVector

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "%s", msg)
	lastUsage += 1
	if lastUsage == len(keys) {
		lastUsage = 0
	}
	fmt.Println("+++ request handled")
}

func main() {
	// for _, key := range keys {
	// 	fmt.Println(mojUkladOdniesienia[key])
	// }

	http.HandleFunc("/axis", ApiHandler)
	fmt.Println("+++ niby wystartowałem api, api route:\n http://localhost:8000/axis")
	http.ListenAndServe("localhost:8080", nil)
}
