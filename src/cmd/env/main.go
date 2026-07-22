package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
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
		if _, exist := keyChain[kv[0]]; exist == false {
			continue
		}

		var marker = "<+> "
		fmt.Printf("+++ %s %32s :key  |   val: %s\n", marker, kv[0], kv[1])
	}
}

func CowsayPrint(here io.Writer, msg string) error {
	var makeErrGoBrr bytes.Buffer
	cmd := exec.Command("cowsay", msg)
	cmd.Stderr = &makeErrGoBrr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(
			" run failed | %w\n----\n%s---- more info ^\n",
			err, ColoredText(makeErrGoBrr.String()),
		)
	}

	return nil
}

func main() {
	msg := "Odkąd dołączyłam do \"szkoła bezczeleności\", jestem takaaaaa beszczelna"
	var runes = []rune(msg)
	for i := range runes {
		if err := CowsayPrint(os.Stdout, string(runes[:i+1])); err != nil {
			fmt.Printf("+++ jednak nie taka bezczelana | %s", err.Error())
		}

		time.Sleep(time.Millisecond * 50)
	}

	envName := "DATASET_NAME"
	err := os.Setenv(envName, "bicycle")
	if err != nil {
		fmt.Printf("+++ failed to set env | %s", err.Error())
		os.Exit(1)
	}

	showEnvKyes(envName, "PWD", "CUDA", "ROCM")
}
