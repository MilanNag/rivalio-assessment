#!/bin/sh
# Runs the Go API and the Next.js server in one container and exits if
# either of them dies, so the platform (Fly.io) can restart the machine.
set -e

/app/server &
backend=$!

PORT=3000 node /app/web/server.js &
frontend=$!

shutdown() {
  kill -TERM "$backend" "$frontend" 2>/dev/null
  wait
  exit 0
}
trap shutdown TERM INT

while kill -0 "$backend" 2>/dev/null && kill -0 "$frontend" 2>/dev/null; do
  sleep 1
done

echo "a process exited unexpectedly; shutting down" >&2
kill -TERM "$backend" "$frontend" 2>/dev/null
wait
exit 1
