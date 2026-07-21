package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func ColoredText(text string) string {
	return fmt.Sprintf("\033[38;5;198m%s\033[0m", text)
}

type KeyHanger = struct{}

var keyHangerInUse = struct{}{}

func showEnvKyes(keys ...string) {
	keyChain := make(map[string]KeyHanger, 8)
	for _, keyVal := range keys {
		keyChain[keyVal] = keyHangerInUse
	}

	for _, envKeyValue := range os.Environ() {
		kv := strings.Split(envKeyValue, "=")
		if _, ok := keyChain[kv[0]]; ok == false {
			continue
		}

		var marker = "<+> "
		fmt.Printf("+++ %s %32s :key  |   val: %s\n", marker, kv[0], kv[1])
	}
}

var toShow bool = false

func main() {
	envName := "DATASET_NAME"
	err := os.Setenv(envName, "bicycle")
	if err != nil {
		fmt.Printf("+++ failed to set env | %s", err.Error())
		os.Exit(1)
	}

	showEnvKyes(envName, "PWD")

	cmd := exec.Command("make", "ollama_cpu")
	var makeErrBfr bytes.Buffer
	cmd.Stderr = &makeErrBfr
	if err := cmd.Run(); err != nil {
		fmt.Printf("!!! cmd failed to run | %s\n", err.Error())
		fmt.Printf("---- \n%s---- more info ^\n", ColoredText(makeErrBfr.String()))
		os.Exit(1)
	}

	if err := cmd.Wait(); err != nil {
		fmt.Printf("+++ cmd run failed | %s", err.Error())
		os.Exit(1)
	}

	fmt.Printf("+++ sucess\n")
}
