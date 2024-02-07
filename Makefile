.PHONY: all build_local run_local local build_docker clean_docker run_docker docker

all: local

build_local:
	@go build -o bin/slug
run_local:
	@. ./.env && ./bin/slug -document_root=$$DOCUMENT_ROOT -port=$$PORT -timeout=$$TIMEOUT
local: build_local run_local

build_docker:
	docker build -t slug . && \
	docker images slug
run_docker:
	. ./.env && \
	docker run --env DOCUMENT_ROOT=$$DOCUMENT_ROOT --env PORT=$$PORT --env TIMEOUT=$$TIMEOUT -p $$PORT:$$PORT slug 
clean_docker:
	docker container kill $$(docker container ls -aq); \
	docker rm $$(docker ps -a -q) \
	docker system prune --all --force --volumes; \
	docker image rm $$(docker image ls -a -q)
docker : build_docker run_docker