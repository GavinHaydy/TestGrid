.PHONY: help up down restart logs ps build clean dev start stop local-restart

COMPOSE_FILE := deploy/docker-compose.yml

RUNTIME := $(shell if command -v docker >/dev/null 2>&1; then echo docker; elif command -v podman >/dev/null 2>&1; then echo podman; fi)

COMPOSE := $(RUNTIME) compose -f $(COMPOSE_FILE)

help:
	@echo "make up          	Start containers"
	@echo "make down        	Stop containers"
	@echo "make reload      	Restart containers"
	@echo "make logs        	View container logs"
	@echo "make ps          	View container status"
	@echo "make build       	Build images"
	@echo "make clean       	Remove containers and volumes"
	@echo "make start       	Start local services"
	@echo "make stop        	Stop local services"
	@echo "make restart 		Restart local services"

check:
	@if [ -z "$(RUNTIME)" ]; then \
		echo "docker or podman not found"; \
		exit 1; \
	fi

up: check
	$(COMPOSE) up -d --build

down: check
	$(COMPOSE) down

reload: check
	$(COMPOSE) restart

logs: check
	$(COMPOSE) logs -f

ps: check
	$(COMPOSE) ps

build: check
	$(COMPOSE) build

clean: check
	$(COMPOSE) down -v --remove-orphans

# 本地开发服务
start:
	./dev.sh start

stop:
	./dev.sh stop

restart:
	./dev.sh restart