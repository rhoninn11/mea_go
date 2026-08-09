
build:
	docker build -f Dockerfile.basic -t ${USER}_basic_img .

up:
	docker compose -f compose.basic.yaml up

shell:
	docker exec -it ${USER}_basic /bin/bash -lc

down:
	docker compose -f compose.basic.yaml down

compose_run:
	docker compose -f compose.basic.yaml run --remove-orphans pod /bin/bash -lc "exec python simple.py"

ollama:
	ollama serve