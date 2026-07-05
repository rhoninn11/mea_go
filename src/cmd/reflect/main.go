package main

import (
	"fmt"
	"log"
	"mea_go/src/internal"
	"os"
	"regexp"
	"strings"
)

type RpcDesc struct {
	Fn     string
	InTpy  string
	OutTpy string
}
type ProtoTypes = []string
type RpcDescs = []RpcDesc

func main() {
	fmt.Println("dzien dobry")
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatal("failed to obtaine work dir")
	}
	fmt.Printf("+++ work dir is: %s\n", workDir)

	dir := "src/api/proto"
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
		protoTypes, err := rpcReflections(proto)
		if err != nil {
			log.Fatalf("failed to parse %s", proto)
		}

		fmt.Printf(" %s has %d types:\n", proto, len(protoTypes))
		fmt.Println(protoTypes)
		fmt.Printf("\n\n")
	}
}

func rpcReflections(filePath string) (ProtoTypes, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("!!! failed to read file, %v", err)
	}

	fmt.Println("+++ listing content of ", filePath)
	protos, err := protoTypes(content)
	if err != nil {
		return nil, fmt.Errorf("!!! type extraction failed, %v", err)
	}
	rpcs, err := protoRpcs(content)
	if err != nil {
		return nil, fmt.Errorf("!!! rpcs extraction failed, %v", err)
	}
	_ = rpcs
	// fmt.Println(rpcs)

	return protos, nil
}

func protoTypes(content []byte) (ProtoTypes, error) {
	reRule := `message\s(\w+)\s{`
	protoRe, err := regexp.Compile(reRule)
	if err != nil {
		return nil, fmt.Errorf("!!! fauled to compile re, %v", err)
	}

	reResult := protoRe.FindAllSubmatch(content, -1)
	typeSlice := make(ProtoTypes, len(reResult))
	for i, proto := range reResult {
		typeSlice[i] = string(proto[1])
	}

	return typeSlice, nil
}

func protoRpcs(content []byte) (RpcDescs, error) {
	rpcRule := `rpc\s(\w+)\((\w+)\)\sreturns\s\((\w+)\);`
	rpcRe, err := regexp.Compile(rpcRule)
	if err != nil {
		return nil, fmt.Errorf("!!! rpc rule compile failed, %v", err.Error())
	}

	rpcSearch := rpcRe.FindAllSubmatch(content, -1)
	rpcMethods := make([]RpcDesc, len(rpcSearch))
	for i, single := range rpcSearch {
		if len(single) < 4 {
			continue
		}
		rpcMethods[i] = RpcDesc{
			Fn:     string(single[1]),
			InTpy:  string(single[2]),
			OutTpy: string(single[3]),
		}
	}
	return rpcMethods, nil
}
