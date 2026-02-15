package translte

import (
	"bytes"
	"context"
	"fmt"
	"mea_go/src/internal"
	"net/http"
	"text/template"

	ollama "github.com/ollama/ollama/api"
)

func PromptEng2Pl(text string) string {
	file := "assets/gemma.txt"
	prompt := template.Must(template.ParseFiles(file))

	data := map[string]string{
		"SOURCE_LANG": "English",
		"TARGET_LANG": "Polish",
		"SOURCE_CODE": "en",
		"TARGET_CODE": "pl",
		"TEXT":        text,
	}

	var buf bytes.Buffer
	err := prompt.Execute(&buf, data)
	internal.CloseOnError(err)
	return buf.String()
}

func PromptPl2Eng(text string) string {
	file := "assets/gemma.txt"
	prompt := template.Must(template.ParseFiles(file))

	data := map[string]string{
		"SOURCE_LANG": "Polish",
		"TARGET_LANG": "English",
		"SOURCE_CODE": "pl",
		"TARGET_CODE": "en",
		"TEXT":        text,
	}

	var buf bytes.Buffer
	err := prompt.Execute(&buf, data)
	internal.CloseOnError(err)
	return buf.String()
}

func prevMain() {
	var prompt string
	textEng := "Uncle ben went fishing today, weather is warm he feels calm drinking cool beverage"
	prompt = PromptEng2Pl(textEng)
	fmt.Println(prompt)

	textPl := "Pan zdzisiek wybrał się na ryby, \"ale dziś będą brały\" myśli sobie... zadowolny"
	prompt = PromptPl2Eng(textPl)
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

type TranslateState struct {
	name string
}

func (ts *TranslateState) StreamedTranslateion(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	internal.SetContentType(w, internal.ContentType_EventStream)
	internal.SetCacheControl(w, internal.CacheType_NoCache)
	w.Header().Set("Connection", "keep-alive")

	_ = flusher
	// TODO: start request to ollama
	// https://claude.ai/chat/06d0f9d0-9a38-4923-bf99-d6b4c7988c99
}
