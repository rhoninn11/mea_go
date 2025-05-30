

templ:
	templ generate

serve: templ
	go run cmd/comfy/main.go

dev: templ css
	go run cmd/serve_site/main.go

css:
	tailwindcss -i ./static/tailwind.css -o ./static/style.css

build_site: templ css
	go build -o ./tmp/main cmd/serve_site/main.go

proto: 
	go run cmd/protogen.go
