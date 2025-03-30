package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ollama/ollama/api"
	"rsc.io/quote/v4"
)

func close(message string) {
	fmt.Println(message)
	os.Exit(1)
}

func assertClientExist() {
	if glob_client == nil {
		close("!!! client not initialized")
	}
}

func HasModel(name string) bool {
	resp, err := glob_client.List(ctx)
	if err != nil {
		close("!!! ollama server not running")
	}

	for _, modelResp := range resp.Models {
		if modelResp.Name == name {
			return true
		}
	}
	return false
}

func PullModel(name string) {
	pull := &api.PullRequest{
		Model: name,
	}
	_ = glob_client.Pull(ctx, pull, OnProgress)

}

var glob_client *api.Client = nil
var ctx context.Context = nil

func SpawnClient() {
	api_client, err := api.ClientFromEnvironment()
	if err != nil {
		close("!!! spawn failed")
	}
	glob_client = api_client
	ctx = context.Background()
}

func OnProgress(progress api.ProgressResponse) error {
	fmt.Println(progress.Digest)
	return nil
}

func OnResonse() error {
	return nil
}

func main() {
	fmt.Println("Hello, World!")
	fmt.Println(quote.Go())

	// ctd := context.Background()
	modelName := "gemma3:12b"

	SpawnClient()
	assertClientExist()

	itHas := HasModel(modelName)
	if !itHas {
		PullModel(modelName)
	}

}
