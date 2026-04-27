default: testacc

# Run acceptance tests (mock-based, requires env vars)
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

# Run unit tests (no TF_ACC, skips acceptance tests)
.PHONY: test
test:
	go test ./... -v $(TESTARGS) -timeout 30m

# Start local Superset via Docker Compose
.PHONY: docker-up
docker-up:
	docker compose up -d
	./scripts/wait-for-superset.sh

# Stop local Superset
.PHONY: docker-down
docker-down:
	docker compose down -v

# Run acceptance tests against local Docker Superset
.PHONY: testacc-docker
testacc-docker: docker-up
	SUPERSET_HOST=http://localhost:8088 \
	SUPERSET_USERNAME=admin \
	SUPERSET_PASSWORD=admin \
	TF_ACC=1 \
	go test ./... -v $(TESTARGS) -timeout 120m
