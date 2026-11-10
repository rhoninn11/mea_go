package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

const cmd = "./llama/llama-server --port 8080 --host 0.0.0.0 -lv 3 -ngl 99 -c 4096 --temp 1.0 --top-p 0.95 --top-k 20 --min-p 0.00 -fa on -m ./gguf/HY-MT2-7B-Q8_0.gguf"

func main() {
	ctx, _ := context.WithCancel(context.Background())

	argv := strings.Split(cmd, " ")
	cmdol := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmdol.Dir = "../ai"
	cmdol.Stderr = os.Stdout
	cmdol.Stdout = os.Stdout

	if err := cmdol.Start(); err != nil {
		fmt.Printf("!!! start error | %s\n", err.Error())
		return
	}
	defer func() {
		_ = cmdol.Process.Signal(os.Interrupt)
		status, err := cmdol.Process.Wait()
		if err != nil {
			fmt.Printf("!!! exit code was %w\n", err.Error())
			return

		}
		code := status.ExitCode()
		fmt.Printf("exit code was: %d\n", code)
	}()

	time.Sleep(time.Second * 5)

	if err := submain(); err != nil {
		fmt.Printf("!!! sumain failed | err was %s\n", err.Error())
	}
	time.Sleep(time.Second * 5)

}

var apiLink = "http://0.0.0.0:8080/v1"

func submain() error {

	var netClient = http.DefaultClient
	aiClient := openai.NewClient(
		option.WithBaseURL(apiLink),
		// option.WithBaseURL("https://mtls.api.openai.com/v1"),
		option.WithHTTPClient(netClient),
	)
	fmt.Printf("+++ befoere models\n")

	pag, err := aiClient.Models.List(context.Background())
	if err != nil {
		return err
	}

	for i, d := range pag.Data {
		fmt.Printf("model %d: %v\n", i, d)
	}
	fmt.Printf("+++ after models\n")

	stream := aiClient.Chat.Completions.NewStreaming(context.Background(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("be respectful and curious"),
			openai.UserMessage("jak się się mówi na kiełasę w niemczech, a jak się mówi w wietnami (fonetycznie)?"),
		},
	})

	for stream.Next() {
		chunk := stream.Current()
		token := chunk.Choices[0].Delta.Content
		fmt.Printf("+ %s\n", token)
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
