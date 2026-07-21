

css:
	tailwindcss -i ./static/tailwind.css -o ./static/style.css

templ:
	go tool templ generate

serve: templ
	go run cmd/comfy/main.go

dev: templ css client
	@echo empty

client:
	go run src/cmd/mea_client/main.go	

vartiants = pages comfy

devel_up:
	@go tool air -c .air.comfy.toml

devel_up_pages:
	@go tool air -c .air.pages.toml

build_site: templ css
	go build -o ./tmp/main cmd/serve_site/main.go

proto: 
	go run src/cmd/protogen.go

ollama_cpu:
	CUDA_VISIBLE_DEVICES="" ollama serve

build:
	mkdir -p _build && go build -o ./_build/mea_client ./src/cmd/mea_client/*.go

build_arxive:
	mkdir -p _build && go build -o ./_build/arxive ./src/cmd/arxive/*.go

go_grpc:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

env:
	go run src/cmd/env/*.go