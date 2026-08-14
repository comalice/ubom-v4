.PHONY: server frontend dev

server:
	$(MAKE) -C server run

frontend:
	$(MAKE) -C frontend dev

dev:
	trap 'kill $(server_pid) $(frontend_pid) 2>/dev/null || true' INT TERM EXIT; \
	$(MAKE) server & server_pid=$$!; \
	$(MAKE) frontend & frontend_pid=$$!; \
	wait $(server_pid) $(frontend_pid)
