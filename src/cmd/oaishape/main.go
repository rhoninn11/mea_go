package main

import (
	"context"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func main() {
	client := openai.NewClient(
		option.WithBaseURL("https://mtls.api.openai.com/v1"),
		option.WithHTTPClient(httpClient),
	)

	if _, err := client.Models.List(context.Background()); err != nil {
		return err
	}
}
