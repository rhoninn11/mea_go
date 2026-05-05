

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

devel_up:
	@go tool air

devel_up_pages:
	@go tool air -c .air.pages.toml

build_site: templ css
	go build -o ./tmp/main cmd/serve_site/main.go

proto: 
	go run src/cmd/protogen.go

ollama:
	CUDA_VISIBLE_DEVICES="" ollama serve

build:
	mkdir -p _build && go build -o ./_build/air_exe ./src/cmd/mea_client/main.go

build_pages:
	mkdir -p _build && go build -o ./_build/air_exe ./src/cmd/arxive/main.go