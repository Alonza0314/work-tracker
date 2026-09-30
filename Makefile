.PHONY: backend frontend openapi run test tidy lint clean docker dockertest
BACKEND_SRC := $(shell find web/backend -name "*.go")
FRONTEND_SRC := $(shell find web/frontend -type f ! -path "web/frontend/dist/*" ! -path "web/frontend/node_modules/*")
FRONTEND_STAMP := build/frontend/.stamp

# keep in sync with backend.db.path in config.yaml
DB_PATH := /tmp/wt.db

all: backend frontend

build/wt: $(BACKEND_SRC)
	@echo "[+] Building backend..."
	mkdir -p build
	cd web/backend && go build -o ../../build/wt .
	@echo "[✔] Backend build finished"

build/frontend: $(FRONTEND_SRC)
	@echo "[+] Installing frontend deps..."
	cd web/frontend && yarn install

	@echo "[+] Building frontend..."
	cd web/frontend && yarn build

	@echo "[✔] Frontend build finished"
	@mkdir -p build/frontend
	@cp -r web/frontend/dist/. build/frontend/
	@touch $(FRONTEND_STAMP)

backend:
	@if [ -f build/wt ]; then \
		if [ -z "$$(find web/backend -name '*.go' -newer build/wt)" ]; then \
			echo "[✔] backend is up-to-date, no build needed"; \
			exit 0; \
		fi; \
	fi; \
	$(MAKE) build/wt

frontend:
	@if [ -f $(FRONTEND_STAMP) ]; then \
		if [ -z "$$(find web/frontend -type f -newer $(FRONTEND_STAMP))" ]; then \
			echo "[✔] frontend is up-to-date, no build needed"; \
			exit 0; \
		fi; \
	fi; \
	$(MAKE) build/frontend

openapi:
	@echo "[+] Generating OpenAPI client..."
	cd web && ./openapi-generator-docker.sh
	@echo "[✔] OpenAPI client generated"

run:
	./build/wt -c config.yaml

test:
	cd web/backend && go test ./...

tidy:
	cd web/backend && go mod tidy

lint:
	cd web/backend && golangci-lint run

clean:
	rm -rf build
	rm -f $(DB_PATH)

docker:
	./docker/build_image.sh

dockertest:
	./docker/build_image.sh test
