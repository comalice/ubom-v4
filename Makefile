.PHONY: server frontend seed dev

DB ?= dev.db
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
	trap 'kill $(server_pid) $(frontend_pid) 2>/dev/null || true' INT TERM EXIT; \
	$(MAKE) -C server run DB=$(DB) & server_pid=$$!; \
	$(MAKE) frontend & frontend_pid=$$!; \
	wait $(server_pid) $(frontend_pid)
