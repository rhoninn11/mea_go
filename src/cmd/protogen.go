package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func spf(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}

func errTerm(err error) {
	// would be nice to identify where from error originated
	// some code inside err whould be huuuuge
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func allFiles(pref string, ext string) []string {
	entrySlice, err := os.ReadDir(pref)
	errTerm(err)
	var files []string
	suf := strings.HasSuffix
	for _, entry := range entrySlice {
		if !entry.IsDir() && suf(entry.Name(), ext) {
			files = append(files, entry.Name())
		}
	}
	return files
}

func sysCmd(args []string) []byte {
	fmt.Println(args)
	run := exec.Command(args[0], args[1:]...)
	// run := exec.Command("protoc", "-h")
	out, err := run.Output()
	if err != nil {
		fmt.Print(out)
		log.Fatal(err.Error())
	}
	return out
}

// eg.: protoc --go_out=./api  --go-grpc_out=./api  ./api/proto/comfy.proto
func genGoFor(inProto string) {
	goPathCmd := []string{
		"go",
		"env",
		"GOPATH",
	}
	goPath := string(sysCmd(goPathCmd))
	goPath = goPath[:len(goPath)-1]
	typesPlug := spf("--plugin=%s/bin/protoc-gen-go", goPath)
	servicePlug := spf("--plugin=%s/bin/protoc-gen-go-grpc", goPath)
	fmt.Println(typesPlug)
	fmt.Println(servicePlug)
	protoCompilation := []string{
		"protoc", typesPlug, servicePlug,
		"--go_out=./api",
		"--go-grpc_out=./api",
		inProto,
	}

	_ = sysCmd(protoCompilation)
	fmt.Printf("+++ %s compiled\n", inProto)
}

func join(dir string, file string) string {
	return spf("%s/%s", dir, file)
}

func main() {
	protoPrefix := "src/api/proto"
	protos := allFiles(protoPrefix, ".proto")

	for _, fname := range protos {
		path := join(protoPrefix, fname)
		genGoFor(path)
	}

	// exec.Command()
}
