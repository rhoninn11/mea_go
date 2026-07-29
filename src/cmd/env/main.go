package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func ColoredText(text string) string {
	return fmt.Sprintf("\033[38;5;198m%s\033[0m", text)
}

type KeyHanger = struct{}

var keyHangerInUse = KeyHanger{}

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

const Path_Docker = "assets/docker"

func DockerTest(stdout io.Writer, stderr io.Writer) error {
	cmd := exec.Command("make", "python")
	cmd.Dir = Path_Docker
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("fun failed | %w", err)
	}

	return nil
}

func cmdAssembly() []string {
	cmdStr := fmt.Sprintf("docker exec -t %s_basic /bin/bash -lc", os.Getenv("USER"))
	argv := strings.Split(cmdStr, " ")
	argv = append(argv, "exec python simple.py")
	return argv
}

func DockerTestAlt(stdout io.Writer, stderr io.Writer) error {
	argv := cmdAssembly()
	fmt.Println(argv)
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = Path_Docker
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("alt run faile | %w", err)
	}

	fmt.Println(cmd.ProcessState.ExitCode())

	return nil
}

var alt bool = false

func parseArgs() {
	flag.BoolVar(&alt, "alt", false, "run alternative function")
	flag.Parse()
}

type CountingWriter struct {
	w      io.Writer
	lines  int
	prefix string
}

func (c *CountingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	delta := bytes.Count(p[:n], []byte{'\n'})
	c.lines += delta
	return n, err
}

func main() {
	parseArgs()

	msg := "Odkąd dołączyłam do \"szkoła bezczeleności\", jestem takaaaaa beszczelna"
	if err := CowsayPrint(os.Stdout, msg); err != nil {
		fmt.Printf("+++ jednak nie taka bezczelana | %s", err.Error())
	}

	envName := "DATASET_NAME"
	if err := os.Setenv(envName, "bicycle"); err != nil {
		fmt.Printf("+++ failed to set env | %s", err.Error())
		os.Exit(1)
	}

	showEnvKyes(envName, "PWD", "USER")

	var next = DockerTest

	if alt {
		next = DockerTestAlt
	}

	// cStdout := &CountingWriter{w: os.Stdout, prefix: "out"}
	// cStderr := &CountingWriter{w: os.Stderr, prefix: "err"}
	if err := next(os.Stdout, nil); err != nil {
		fmt.Printf("+++ docker test failed | %s", err.Error())
		os.Exit(1)
	}

}
