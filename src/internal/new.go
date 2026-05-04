package internal

import (
	"bytes"
	"fmt"
	"os/exec"
	"path"
	"strings"
)

func ColoredText(text string) string {
	return fmt.Sprintf("\033[38;5;198m%s\033[0m", text)
}
func ColoredText2(text string) string {
	return fmt.Sprintf("\033[38;5;198m%s\033[0m", text)
}

func RenderPdf(pdf string, here string, pageI int) error {
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

func CountPages(pdf string) error {
	cmd := fmt.Sprintf("mutool pages %s", pdf)

	argv := strings.Split(cmd, " ")
	cmdExec := exec.Command(argv[0], argv[1:]...)

	data, err := cmdExec.Output()
	if err != nil {
		return fmt.Errorf("%s - failed | %w", ColoredText(cmd), err)
	}

	var pageNum int
	lines := bytes.Split(data, []byte("\n"))
	for _, line := range lines {
		if bytes.HasPrefix(line, []byte("<page pagenum=\"")) {
			_, err := fmt.Sscanf(string(line), "<page pagenum=\"%d\">", &pageNum)
			if err != nil {
				return err
			}
		}
	}
	fmt.Printf("+++ page count was: %d\n", pageNum)
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
