
build:
	docker build -f Dockerfile.basic -t ${USER}_basic_img .

up:
	docker compose -f compose.basic.yaml up

down:
	docker compose -f compose.basic.yaml down

shell:
	docker exec -it ${USER}_basic /bin/bash -lc

python:
	docker exec -t ${USER}_basic /bin/bash -lc "exec python simple.py"