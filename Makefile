

templgen:
	templ generate

serve: templgen
	go run cmd/comfy/main.go