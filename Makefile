.PHONY: server frontend seed dev dev-postgres dev-run

DATA_DIR ?= ./.data
DB ?= $(DATA_DIR)/dev.db
POSTGRES_DATA_DIR ?= $(DATA_DIR)/postgres
POSTGRES_DSN ?= postgresql://ubom:ubom-dev-only@localhost:5432/ubom?sslmode=disable
PARTS ?= 50
MAX_REVISIONS ?= 3
MAX_BOM_DEPTH ?= 3
SEED ?= 1

server:
	$(MAKE) -C server run

frontend:
	$(MAKE) -C frontend dev

seed:
	$(MAKE) -C server seed DB=$(DB) PARTS=$(PARTS) MAX_REVISIONS=$(MAX_REVISIONS) MAX_BOM_DEPTH=$(MAX_BOM_DEPTH) SEED=$(SEED)

dev:
	$(MAKE) seed
	$(MAKE) dev-run

dev-postgres:
	POSTGRES_DATA_DIR="$(POSTGRES_DATA_DIR)" docker compose up -d --wait postgres
	$(MAKE) dev-run SERVER_ENV="UBOM_DATABASE_URL=$(POSTGRES_DSN)"

dev-run:
	stop_group() { pgid=$$1; test -n "$$pgid" || return; kill -TERM -- "-$$pgid" 2>/dev/null || true; for attempt in 1 2 3 4 5 6 7 8 9 10; do kill -0 -- "-$$pgid" 2>/dev/null || return; sleep 0.1; done; kill -KILL -- "-$$pgid" 2>/dev/null || true; }; \
	cleanup() { status=$$1; trap - INT TERM EXIT; stop_group "$$server_pgid"; stop_group "$$frontend_pgid"; wait "$$server_pid" "$$frontend_pid" 2>/dev/null || true; exit "$$status"; }; \
	trap 'cleanup $$?' INT TERM EXIT; \
	setsid --wait env $(SERVER_ENV) $(MAKE) -C server run DB=$(DB) & server_pid=$$!; server_pgid=$$(ps -o pgid= -p "$$server_pid" | tr -d ' '); \
	setsid --wait $(MAKE) frontend & frontend_pid=$$!; frontend_pgid=$$(ps -o pgid= -p "$$frontend_pid" | tr -d ' '); \
	wait "$$server_pid" "$$frontend_pid"
