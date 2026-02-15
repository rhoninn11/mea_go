package main

import (
	"context"
	"fmt"
	"log"
	"mea_go/src/internal/translte"
	"strings"

	"github.com/ollama/ollama/api"
)

func main() {
	textEng := "Uncle ben went fishing today, weather is warm he feels calm drinking cool beverage"
	var eng2plPrompt = translte.PromptEng2Pl(textEng)
	fmt.Println(eng2plPrompt)

	textPl := "Pan zdzisiek wybrał się na ryby, \"ale dziś będą brały\" myśli sobie... zadowolny"
	pl2EngPrompt := translte.PromptPl2Eng(textPl)
	fmt.Println(pl2EngPrompt)

	ollama, err := translte.StartApi()
	if err != nil {
		log.Println(err.Error())
	}

	_ = ollama

	tokens := make([]string, 0, 64)
	err = ollama.Chat(
		context.Background(),
		&api.ChatRequest{
			Model: "translategemma",
			Messages: []api.Message{
				{
					Role:    "user",
					Content: pl2EngPrompt,
				},
			},
		},
		func(cr api.ChatResponse) error {
			token := cr.Message.Content
			tokens = append(tokens, token)
			fmt.Print(token)
			return nil
		},
	)
	fmt.Printf("\n")
	if err != nil {
		log.Println("ollama chat failed")
	}

	fmt.Printf("translatation: %s\n", strings.Join(tokens, ""))
}
