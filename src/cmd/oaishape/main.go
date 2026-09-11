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

const cmd = "/llama-server --port 8080 --host 0.0.0.0 -lv 3 -ngl 99 -c 65536 --temp 1.0 --top-p 0.95 --top-k 20 --min-p 0.00 -fa on -m ../gguf/Qwen3.8-27B-UD-Q3_K_XL.gguf"

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
