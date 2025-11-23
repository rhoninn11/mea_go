package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func closeOnError(err error, msg string) {
	if err != nil {
		fmt.Println(msg)
		os.Exit(1)
	}
}

func main() {
	url := "http://localhost:8080/axis"

	fmt.Printf("+++ connecting over url:\n %s\n", url)

	resp, err := http.Get(url)
	closeOnError(err, "!!! request failed")

	data, err := io.ReadAll(resp.Body)
	closeOnError(err, "!!! read error")

	fmt.Printf("+++ recived data:\n %s \n", string(data[:]))
}
