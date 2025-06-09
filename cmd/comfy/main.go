package main

import (
	"log"
	"mea_go/internal"
)

func main() {
	// TODO: As we connecting with commpressed comfy api
	// we also consider retransimiting data over websocket,
	// for other technoogies not supported by main protocol
	// but at first it can be simple cli?! inpaint editor??

	glob := internal.DefaultComfySpawn()

	prompts := []string{
		"warrior wakes up early morning, in his friend house after yesterdey's happy bonfire day:D",
		"mage wakes up erlier so he can go fishing with uncle bob",
	}
	err := glob.GenFew(prompts)
	if err != nil {
		log.Fatalln(err.Error())
	}

}
