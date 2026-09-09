package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func main() {
	if err := raw(); err != nil {
		fmt.Printf("sumain failed")
	}
}

var apiLink = "http://0.0.0.0:8080/v1"

func submain() error {

	var netClient = http.DefaultClient
	aiClient := openai.NewClient(
		option.WithBaseURL(apiLink),
		// option.WithBaseURL("https://mtls.api.openai.com/v1"),
		option.WithHTTPClient(netClient),
	)

	if _, err := aiClient.Models.List(context.Background()); err != nil {
		return err
	}

	return nil
}

func raw() error {
	resp, err := http.Get(apiLink + "/models")
	if err != nil {
		return fmt.Errorf("get failed | %w", err)
	}

	stat := resp.Status
	fmt.Printf("models status %s\n", stat)

	bNum, err := io.Copy(os.Stdout, resp.Body)
	if err != nil {
		return fmt.Errorf("copy failed | %w", err)

	}
	fmt.Printf("+++ writen %d to stdout\n", bNum)
	return nil
}
