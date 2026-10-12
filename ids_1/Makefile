.PHONY: up down down-v logs rebuild ps psql test-api build run install-migrate migrate-up migrate-down migrate-version migrate-new swagger

MIGRATE_VERSION := v4.19.0

ENV_FILE ?= .env
COMPOSE   = ENV_FILE=$(ENV_FILE) docker compose --env-file $(ENV_FILE)

MIGRATE_RUN = set -a && . $(abspath $(ENV_FILE)) && . ./migrate.conf && set +a && \
	bin/migrate -path "$$MIGRATE_PATH" -database "$$MIGRATE_DATABASE_URL"

up:              ## поднять БД, накатить миграции, поднять приложение
	$(COMPOSE) up -d --wait db
	$(MAKE) migrate-up
	$(COMPOSE) up --build -d app

down:            ## остановить (с удалением данных: make down-v)
	$(COMPOSE) down

down-v:
	$(COMPOSE) down -v

logs:
	$(COMPOSE) logs -f app

rebuild:
	$(COMPOSE) up --build -d --force-recreate app

ps:
	$(COMPOSE) ps

psql:            ## зайти в базу внутри контейнера
	$(COMPOSE) exec db psql -U app -d app

bin/migrate:
	GOBIN=$(CURDIR)/bin go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION)

install-migrate: bin/migrate ## поставить golang-migrate в ./bin

migrate-up: bin/migrate      ## применить все новые миграции
	@$(MIGRATE_RUN) up

migrate-down: bin/migrate    ## откатить последнюю миграцию
	@$(MIGRATE_RUN) down 1

migrate-version: bin/migrate ## текущая версия схемы
	@$(MIGRATE_RUN) version

migrate-new: bin/migrate     ## новая пара файлов: make migrate-new name=add_column
	bin/migrate create -ext sql -dir migrations -seq -digits 6 $(name)

swagger:         ## перегенерировать docs/ из аннотаций
	go run github.com/swaggo/swag/cmd/swag@v1.16.6 fmt
	go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g main.go -o docs

build:
	go build -o bin/app .

run:             ## запуск приложения на хосте (БД в докере, см. README)
	set -a && . $(abspath $(ENV_FILE)) && set +a && go run .

test-api:        ## дымовой прогон всех ручек
	./scripts/smoke.sh
