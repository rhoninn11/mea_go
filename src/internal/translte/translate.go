package translte

import (
	"bytes"
	"fmt"
	"mea_go/src/internal"
	"text/template"
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
