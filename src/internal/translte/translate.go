package translte

import (
	"bytes"
	"context"
	"fmt"
	"mea_go/src/internal"
	"text/template"

	ollama "github.com/ollama/ollama/api"
)

const TranslateTemplateFile = "assets/gemma.txt"

type TransDirection int

const (
	PL2EN TransDirection = iota
	EN2PL
)

func langSwitch(text string, Dir TransDirection) map[string]string {
	var data map[string]string
	switch Dir {
	case PL2EN:
		data = map[string]string{
			"SOURCE_LANG": "Polish",
			"TARGET_LANG": "English",
			"SOURCE_CODE": "pl",
			"TARGET_CODE": "en",
			"TEXT":        text,
		}
	case EN2PL:
		data = map[string]string{
			"SOURCE_LANG": "English",
			"TARGET_LANG": "Polish",
			"SOURCE_CODE": "en",
			"TARGET_CODE": "pl",
			"TEXT":        text,
		}
	}
	return data
}

func Translate(text string, dir TransDirection) string {
	var buf bytes.Buffer

	templateIn := langSwitch(text, dir)
	transPrompt := template.Must(template.ParseFiles(TranslateTemplateFile))
	err := transPrompt.Execute(&buf, templateIn)
	internal.CloseOnError(err)
	return buf.String()
}

func prevMain() {
	var prompt string
	textEng := "Uncle ben went fishing today, weather is warm he feels calm drinking cool beverage"
	prompt = Translate(textEng, EN2PL)
	fmt.Println(prompt)

	textPl := "Pan zdzisiek wybrał się na ryby, \"ale dziś będą brały\" myśli sobie... zadowolny"
	prompt = Translate(textPl, PL2EN)
	fmt.Println(prompt)
}

func StartApi() (*ollama.Client, error) {
	client, err := ollama.ClientFromEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ollam")
	}

	ver, err := client.Version(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get version\n")
	}
	fmt.Printf("Connected to ollama (%s)\n", ver)
	return client, nil
}

func SetLocal(external *ollama.Client) {
	local = external
}

var local *ollama.Client = nil

type TranlateJob struct {
	ToTranslate     string
	TranslateResutl string
}

func (ts *TranlateJob) StreamedTranslateion(ctx context.Context, pipeTokensHere chan string) error {
	// TODO: start request to ollama
	// https://claude.ai/chat/06d0f9d0-9a38-4923-bf99-d6b4c7988c99

	if local == nil {
		client, err := StartApi()
		if err != nil {
			return fmt.Errorf("!!! failed to init ollama")
		}
		SetLocal(client)
		fmt.Println("local client set")
	} else {
		fmt.Println("client already set")
	}

	chat := func(message string) *ollama.ChatRequest {
		return &ollama.ChatRequest{
			// Model: "translategemma:12b",
			Model: "translategemma",
			Messages: []ollama.Message{
				{
					Role:    "user",
					Content: message,
				},
			},
		}
	}

	input := Translate(ts.ToTranslate, PL2EN)
	err := local.Chat(ctx, chat(input),
		func(cr ollama.ChatResponse) error {
			pipeTokensHere <- cr.Message.Content
			return nil
		},
	)
	close(pipeTokensHere)
	return err
}
