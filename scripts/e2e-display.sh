#!/usr/bin/env bash
# ──────────────────────────────────────────────────────────────────
# e2e-display — Display wrapper for Playwright E2E tests.
#
# GPU + DISPLAY → use Xorg.  GPU only → xvfb + DRI.  Neither → xvfb.
# ──────────────────────────────────────────────────────────────────
set -euo pipefail

XVFB_SCREEN="${XVFB_SCREEN:-1920x1080x24}"

has_gpu() {
	[ -e /dev/dri/renderD128 ] || [ -e /dev/dri/renderD129 ] ||
	[ -c /dev/dri/card0 ] || [ -c /dev/dri/card1 ]
}

if [ -n "${DISPLAY:-}" ]; then
	echo "[e2e-display] Xorg ${DISPLAY}" >&2
	glxinfo -B 2>/dev/null | grep -E 'Vendor|Renderer|direct rendering' | sed 's/^/  /' >&2 || true
	exec "$@"
elif has_gpu; then
	echo "[e2e-display] xvfb + DRI 硬件加速" >&2
	xvfb-run -a -s "-screen 0 ${XVFB_SCREEN} +extension GLX +extension DRI3 +render" "$@"
else
	echo "[e2e-display] xvfb 软件渲染" >&2
	xvfb-run -a -s "-screen 0 ${XVFB_SCREEN}" "$@"
fi
