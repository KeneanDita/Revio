.PHONY: dev-backend dev-frontend dev migrate-db docker-up docker-down clean

# ─── Local development ────────────────────────────────────────────────────────

dev-backend:
	cd backend && go run cmd/server/main.go

dev-frontend:
	cd frontend && npm run dev

# ─── Docker ───────────────────────────────────────────────────────────────────

docker-up:
	docker compose up --build

docker-down:
	docker compose down -v

docker-db-only:
	docker compose up postgres -d

# ─── Database ─────────────────────────────────────────────────────────────────

migrate-db:
	docker compose exec postgres psql -U $${DB_USER:-revio} -d $${DB_NAME:-revio} \
		-f /docker-entrypoint-initdb.d/001_initial.sql

migrate-db-local:
	psql -h localhost -U $${DB_USER:-revio} -d $${DB_NAME:-revio} \
		-f backend/migrations/001_initial.sql

# ─── Go ───────────────────────────────────────────────────────────────────────

backend-build:
	cd backend && go build -o bin/server ./cmd/server

backend-test:
	cd backend && go test ./...

backend-vet:
	cd backend && go vet ./...

backend-tidy:
	cd backend && go mod tidy

# ─── Frontend ─────────────────────────────────────────────────────────────────

frontend-install:
	cd frontend && npm install

frontend-build:
	cd frontend && npm run build

frontend-type-check:
	cd frontend && npm run type-check

frontend-lint:
	cd frontend && npm run lint

# ─── Setup ────────────────────────────────────────────────────────────────────

setup:
	cp .env.example .env
	cp frontend/.env.local.example frontend/.env.local
	cd frontend && npm install
	cd backend && go mod download
	@echo ""
	@echo "Setup complete. Edit .env with your secrets before starting."

# ─── Clean ────────────────────────────────────────────────────────────────────

clean:
	rm -rf backend/bin
	rm -rf frontend/.next
	rm -rf frontend/out
