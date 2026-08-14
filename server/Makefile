.PHONY: run seed

run:
	go run ./cmd/server $(if $(DB),--db $(DB),)

seed:
	rm -f $(DB)
	go run ./cmd/seed --db $(DB) --parts $(PARTS) --max-revisions $(MAX_REVISIONS) --max-bom-depth $(MAX_BOM_DEPTH) --seed $(SEED)
