#!/bin/sh
# Runs the Go API and the Next.js server in one container and exits if
# either of them dies, so the hosting platform can restart the instance.
#
# The platform-provided PORT (Render sets this) is used by the public
# Next.js server; the Go API always listens internally on 127.0.0.1:8090.
set -e

WEB_PORT="${PORT:-3000}"

PORT=8090 /app/server &
backend=$!

PORT="$WEB_PORT" node /app/web/server.js &
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
