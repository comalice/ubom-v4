.PHONY: run seed

DB ?= dev.db
PARTS ?= 50
MAX_REVISIONS ?= 3
MAX_BOM_DEPTH ?= 3
SEED ?= 1

run:
	go run ./cmd/server $(if $(DB),--db $(DB),)

seed:
	rm -f $(DB)
	go run ./cmd/seed --db $(DB) --parts $(PARTS) --max-revisions $(MAX_REVISIONS) --max-bom-depth $(MAX_BOM_DEPTH) --seed $(SEED)
