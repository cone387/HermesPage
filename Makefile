.PHONY: build run dev clean docker-up docker-down

BINARY=hermespage
GOFLAGS=-v

build:
	go build $(GOFLAGS) -o $(BINARY) .

run: build
	HERMES_API_KEY=dev-key ./$(BINARY) serve

dev:
	HERMES_API_KEY=dev-key go run . serve

mcp:
	HERMES_API_KEY=dev-key HERMES_SERVER_URL=http://localhost:8080 go run . mcp

clean:
	rm -f $(BINARY) $(BINARY).exe

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

test-upload:
	@echo '<html><head><title>Test Report</title><meta name="hermes-tags" content="test,demo"></head><body><h1>Hello HermesPage</h1></body></html>' > /tmp/test-report.html
	curl -X POST -H "Authorization: Bearer dev-key" -F "file=@/tmp/test-report.html" -F "category=test" http://localhost:8080/api/upload
	@echo ""
