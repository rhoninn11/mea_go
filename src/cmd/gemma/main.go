package main

import (
	"fmt"
	"mea_go/src/internal/translte"
)

func main() {
	var prompt string
	textEng := "Uncle ben went fishing today, weather is warm he feels calm drinking cool beverage"
	prompt = translte.PromptEng2Pl(textEng)
	fmt.Println(prompt)

	textPl := "Pan zdzisiek wybrał się na ryby, \"ale dziś będą brały\" myśli sobie... zadowolny"
	prompt = translte.PromptPl2Eng(textPl)
	fmt.Println(prompt)
}
