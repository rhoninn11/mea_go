

templ:
	templ generate

serve: templgen
	go run cmd/comfy/main.go

dev:
	go run cmd/serve_site/main.go

css:
	tailwindcss -i ./static/tailwind.css -o ./static/style.css
