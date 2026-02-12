package translte

import (
	"bytes"
	"context"
	"fmt"
	"mea_go/src/internal"
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

func StartApi() {
	client, err := ollama.ClientFromEnvironment()
	if err != nil {
		fmt.Printf("failed to connect to ollam\n")
		return
	}

	ver, err := client.Version(context.Background())
	if err != nil {
		fmt.Printf("failed to get version\n")
		return
	}
	fmt.Printf("Connected to ollama (%s)\n", ver)
}
