package main

import (
	"bytes"
	"fmt"
	"log"
	"mea_go/internal"
	"os"
	"strings"
)

func parseProto(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("!!! failed to read file, %v", err)
	}

	fmt.Println("+++ listing content of ", filePath)
	lines := bytes.Split(content, []byte("\n"))

	var types [][][]byte = make([][][]byte, 0, 64)
	var start int
	var end int
	var opened bool = false
	for i, line := range lines {
		a := bytes.Contains(line, []byte("message"))
		b := bytes.Contains(line, []byte("{"))
		c := bytes.Contains(line, []byte("}"))

		if a && b && !opened {
			start = i
			opened = true
		}

		if c && opened {
			end = i + 1
			types = append(types, lines[start:end])
			opened = false
		}
	}

	for _, tpy := range types {
		whole := bytes.Join(tpy, []byte("\n"))
		fmt.Printf("---found type: \n%s\n\n", string(whole))
	}

	return nil
}

func main() {
	fmt.Println("dzien dobry")

	dir := "api/proto"
	dirs, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal("error at read dir")
	}

	prots := make([]string, 0, 16)
	for _, entry := range dirs {
		if entry.Type().IsRegular() {
			if strings.HasSuffix(entry.Name(), ".proto") {
				filePath := internal.JoinPath(dir, entry.Name())
				prots = append(prots, filePath)
			}
		}
	}

	for _, proto := range prots {
		err := parseProto(proto)
		if err != nil {
			log.Fatalf("failed to parse %s", proto)
		}
		break
	}
}
