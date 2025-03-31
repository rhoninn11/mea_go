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
	fmt.Println("do we get something over api?")

	url := "http://localhost:8080/api"

	resp, err := http.Get(url)
	closeOnError(err, "!!! request failed")
	data, err := io.ReadAll(resp.Body)
	closeOnError(err, "!!! read error")
	fmt.Println("+++ recived data", data)
}
