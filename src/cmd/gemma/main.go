package main

import (
	"context"
	"fmt"
	"log"
	"mea_go/src/internal/translte"

	"github.com/ollama/ollama/api"
)

func main() {
	textEng := "Uncle ben went fishing today, weather is warm he feels calm drinking cool beverage"
	var eng2plPrompt = translte.Translate(textEng, translte.EN2PL)
	// fmt.Println(eng2plPrompt)
	_ = eng2plPrompt

	textPl := "Pan zdzisiek wybrał się na ryby, \"ale dziś będą brały\" myśli sobie... zadowolny"
	pl2EngPrompt := translte.Translate(textPl, translte.PL2EN)
	// fmt.Println(pl2EngPrompt)

	ollama, err := translte.StartApi()
	if err != nil {
		log.Println(err.Error())
	}

	_ = ollama

	chat := func(message string) *api.ChatRequest {
		return &api.ChatRequest{
			Model: "translategemma",
			Messages: []api.Message{
				{
					Role:    "user",
					Content: message,
				},
			},
		}
	}

	tokens := make([]string, 0, 64)
	err = ollama.Chat(
		context.Background(),
		chat(pl2EngPrompt),
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

}
