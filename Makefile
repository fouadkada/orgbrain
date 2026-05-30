.PHONY: dev stop test migrate codegen load-test

dev:
	docker-compose up --build -d

stop:
	docker-compose down

test:
	cd api && go test ./...
	cd ai-worker && uv run pytest tests/
	cd rag && uv run pytest tests/
	cd signal-job && uv run pytest tests/

migrate:
	goose -dir migrations/shared postgres "$(DIRECT_DATABASE_URL)" up

codegen:
	cd api && oapi-codegen --config oapi-codegen-ai-worker.yaml ../openapi/ai-worker.yaml
	cd api && oapi-codegen --config oapi-codegen-rag.yaml ../openapi/rag.yaml

load-test:
	@echo "load-test: not yet enabled"
