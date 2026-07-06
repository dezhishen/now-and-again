#!/usr/bin/env bash
# ──────────────────────────────────────────────────────────────────
# e2e-entrypoint — Root-first, then user-switch.
#
# Container starts as root → gosu to NA_RUN_AS user.
# Display setup is delegated to e2e-display.sh (xvfb+DRI or software).
#
# NA_RUN_AS: "1000:1000" or "root" or "keep"
# ──────────────────────────────────────────────────────────────────
set -euo pipefail

RUN_AS="${NA_RUN_AS:-root}"

if [ -e /dev/dri/renderD128 ] || [ -e /dev/dri/renderD129 ]; then
	echo "[e2e-entry] ✓ GPU 已检测到" >&2
else
	echo "[e2e-entry] 无 GPU → 软件渲染" >&2
fi

case "$RUN_AS" in
root|keep|"")
	exec "$@"
	;;
*)
	uid="${RUN_AS%%:*}"
	gid="${RUN_AS##*:}"
	[ "$gid" = "$uid" ] && gid="$uid"
	groupadd -g "$gid" e2euser 2>/dev/null || true
	useradd -m -u "$uid" -g "$gid" -d /tmp/e2ehome -s /bin/bash e2euser 2>/dev/null || true
	exec gosu "$uid:$gid" env HOME=/tmp "$@"
	;;
esac
