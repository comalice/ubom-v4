# Development Process Cleanup

## Scope

The root `make dev` target starts the Go server and Vite in separate sessions.
Stopping the target must also stop the compiled child created by `go run` and
the Vite/esbuild child processes.

## Manual Check

From the repository root, use a temporary database and stop the target after
both listeners are ready:

```sh
db=$(mktemp /tmp/ubom-dev.XXXXXX.db)
make dev DB="$db" > /tmp/ubom-dev.log 2>&1 &
dev_pid=$!
while ! ss -ltn '( sport = :8080 or sport = :5173 )' | grep -q LISTEN; do
  kill -0 "$dev_pid" 2>/dev/null || { cat /tmp/ubom-dev.log; exit 1; }
  sleep 1
done
kill -TERM "$dev_pid"
wait "$dev_pid" || test "$?" -eq 143

! ss -ltn '( sport = :8080 or sport = :5173 )' | grep -q LISTEN
rm -f "$db" /tmp/ubom-dev.log
```

Use `ps -eo pid,ppid,pgid,sid,args` to confirm that no server, Vite, or
esbuild descendants from this run remain. Exclude any process that was already
running before the check. In a foreground terminal, press `Ctrl-C` instead of
running `kill -TERM` to exercise the interrupt trap.
