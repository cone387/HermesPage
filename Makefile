.PHONY: build run dev mcp clean docker-up docker-down test-upload

BINARY=acta
GOFLAGS=-v

build:
	go build $(GOFLAGS) -o $(BINARY) .

run: build
	ACTA_JWT_SECRET=dev-secret ./$(BINARY) serve

dev:
	ACTA_JWT_SECRET=dev-secret go run . serve

mcp:
	ACTA_SERVER_URL=http://localhost:5487 go run . mcp

clean:
	rm -f $(BINARY) $(BINARY).exe

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

# usage: make test-upload ACTA_TOKEN=tok_xxx
test-upload:
	@echo '<html><head><title>Test Report</title><meta name="acta-tags" content="test,demo"></head><body><h1>Hello Acta</h1></body></html>' > /tmp/test-report.html
	curl -X POST -H "Authorization: Bearer $(ACTA_TOKEN)" -F "file=@/tmp/test-report.html" -F "category=test" http://localhost:5487/api/upload
	@echo ""
