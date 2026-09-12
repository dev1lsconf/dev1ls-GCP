.PHONY: all test build docker-build docker-run docker-stop clean tf-init tf-plan tf-apply

APP_NAME := devops-api
TAG := v1
PORT := 8080

all: test build

test:
	go test -v -race ./...

build:
	go build -o bin/server ./cmd/server

run:
	PORT=$(PORT) go run ./cmd/server

docker-build:
	docker build -t $(APP_NAME):$(TAG) .

docker-run:
	docker run --rm -d -p $(PORT):$(PORT) --name $(APP_NAME)-local $(APP_NAME):$(TAG)
	@echo "Container running at http://localhost:$(PORT)"
	@echo "Try: curl http://localhost:$(PORT)/healthz"

docker-stop:
	docker stop $(APP_NAME)-local || true

tf-init:
	cd terraform && terraform init

tf-plan:
	cd terraform && terraform plan

clean:
	rm -rf bin/
