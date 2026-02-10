

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

air:
	@air


build_site: templ css
	go build -o ./tmp/main cmd/serve_site/main.go

proto: 
	go run src/cmd/protogen.go
