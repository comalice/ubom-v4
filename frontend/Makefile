.PHONY: dev

dev:
	@command -v npm >/dev/null 2>&1 || { echo "npm is required to run the frontend; install Node.js and npm first" >&2; exit 1; }
	@if [ ! -d node_modules ]; then npm ci; fi
	npm run dev -- --host 0.0.0.0
