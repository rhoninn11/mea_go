package main

import (
	"fmt"
	"mea_go/src/internal"
	"mea_go/src/internal/translte"
	"mea_go/src/internal/txt2img"
	"net/http"
)

const (
	DEBUG = "DEBUG"
	PROD  = "PROD"
)

var registerFn = internal.RegisterHandler

func main() {
	var mode = DEBUG
	_ = mode

	client, err := translte.StartApi()
	if err != nil {
		fmt.Println(err.Error())
	}
	_ = client

	const host = "0.0.0.0"
	const port = 8080
	var base = fmt.Sprintf("%s:%d", host, port) // eg localhost:8080

	static := http.FileServer(http.Dir("./static/"))
	static = http.StripPrefix("/static/", static)
	static = internal.NoCacheMiddleware(static)
	http.Handle("/static/", static)

	promptModule := txt2img.PromptModuleAccess()

	mapping := promptModule.LoadFns()
	for k, v := range mapping {
		fmt.Println("halo", k, v)
		registerFn(k, v)
	}

	var url = fmt.Sprintf("http://%s/%s", base, "gen_page")
	fmt.Printf("+++ niby wystartowałem api, api route: \n%s\n", url)

	_ = http.ListenAndServe(base, nil)
}
