package internal

import (
	"fmt"
	"os/exec"
	"path"
	"strings"
)

func ColoredText(text string) string {
	return fmt.Sprintf("\032[38;5;198m%s\033[0m", text)
}

func render(pdf string, here string, pageI int) error {
	// sudo apt install mupdf-tools
	oFile := path.Join(here, "strona%d.png")
	cmd := fmt.Sprintf("mutool draw -o %s -r 300 %s %d", oFile, pdf, pageI+1)

	argv := strings.Split(cmd, " ")
	cmdExec := exec.Command(argv[0], argv[1:]...)
	if err := ErrRow(cmdExec.Start(), cmdExec.Wait()); err != nil {
		return fmt.Errorf("%s - failed | %w", ColoredText(cmd), err)
	}
	return nil
}

func ErrRow(err ...error) error {
	for _, err := range err {
		if err != nil {
			return err
		}
	}
	return nil
}
