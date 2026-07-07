#!/usr/bin/env bash
# =============================================================================
# validate-tag.sh
#
# 验证推送的 tag 格式是否合法。
# 允许的类型:
#   - 正式版:    v<major>.<minor>.<patch>           e.g. v1.0.7
#   - Preview:   v<major>.<minor>.<patch>-preview-<NNN>  e.g. v1.0.7-preview-001
#   - Patch:     v<major>.<minor>.<patch>-<N>       e.g. v1.0.7-1
#
# 不允许的类型会自动拒绝（可通过 CI 集成阻止不符合规范的 tag 出包）
#
# 用法:
#   ./scripts/validate-tag.sh <tag>
#   退出码 0 = 合法, 1 = 非法
# =============================================================================

set -euo pipefail

TAG="${1:-}"

if [ -z "$TAG" ]; then
    echo "ERROR: TAG is required"
    echo "Usage: $0 <tag>"
    exit 1
fi

# ── 验证函数 ──────────────────────────────────────────────────

is_stable() {
    [[ "$1" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]
}

is_preview() {
    [[ "$1" =~ ^v[0-9]+\.[0-9]+\.[0-9]+-preview-[0-9]+$ ]]
}

is_patch() {
    [[ "$1" =~ ^v[0-9]+\.[0-9]+\.[0-9]+-[0-9]+$ ]]
}

# ── 验证逻辑 ──────────────────────────────────────────────────

if is_stable "$TAG"; then
    echo "✅ 合法 - 正式版: $TAG"
    exit 0
elif is_preview "$TAG"; then
    echo "✅ 合法 - Preview: $TAG"
    exit 0
elif is_patch "$TAG"; then
    echo "✅ 合法 - Patch: $TAG"
    exit 0
else
    echo "❌ 非法 tag 格式: $TAG"
    echo ""
    echo "允许的格式:"
    echo "  正式版:  v<major>.<minor>.<patch>              e.g. v1.0.7"
    echo "  Preview: v<major>.<minor>.<patch>-preview-<NNN>   e.g. v1.0.7-preview-001"
    echo "  Patch:   v<major>.<minor>.<patch>-<N>          e.g. v1.0.7-1"
    echo ""
    echo "常见错误:"
    echo "  v1.0.7-preview   → 缺少编号，应该是 v1.0.7-preview-001"
    echo "  v1.0.7-beta      → 不允许 beta，请使用 -preview-NNN"
    echo "  release-1.0.7    → 缺少 v 前缀"
    echo "  v1.0             → 缺少 patch 版本号"
    exit 1
fi
