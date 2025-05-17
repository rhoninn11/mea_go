

templgen:
	templ generate

serve: templgen
	go run cmd/simple_serv/main.go