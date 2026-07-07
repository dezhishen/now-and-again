#!/usr/bin/env bash
# =============================================================================
# cleanup-preview-releases.sh
#
# 正式版发布时，删除该版本对应的所有 preview Release。
# 用法:
#   ./scripts/cleanup-preview-releases.sh <stable_tag> <github_token>
#
# 例如: v1.0.7 发布 → 删除所有 v1.0.7-preview-* 的 Release
# =============================================================================

set -euo pipefail

STABLE_TAG="${1:-}"
GITHUB_TOKEN="${2:-}"
GITHUB_REPO="${GITHUB_REPOSITORY:-dezhishen/now-and-again}"

if [ -z "$STABLE_TAG" ] || [ -z "$GITHUB_TOKEN" ]; then
    echo "Usage: $0 <stable_tag> <github_token>"
    echo "  Example: $0 v1.0.7 \${{ secrets.GITHUB_TOKEN }}"
    exit 1
fi

# 验证是否为正式版 tag
if ! [[ "$STABLE_TAG" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "ERROR: $STABLE_TAG is not a stable release tag (expected v<major>.<minor>.<patch>)"
    exit 1
fi

echo "=== 清理 Preview Releases ==="
echo "  Stable:  ${STABLE_TAG}"
echo "  Repo:    ${GITHUB_REPO}"

# 查找所有匹配的 preview tags
PREVIEW_TAGS=$(git tag -l "${STABLE_TAG}-preview-*" 2>/dev/null || true)

if [ -z "$PREVIEW_TAGS" ]; then
    echo "  没有找到 preview tags，跳过"
    exit 0
fi

echo "  找到 preview tags:"
echo "$PREVIEW_TAGS" | while read -r tag; do echo "    - $tag"; done

# 通过 GitHub API 删除每个 preview Release
for tag in $PREVIEW_TAGS; do
    # 获取 Release ID
    release_id=$(curl -s -H "Authorization: token ${GITHUB_TOKEN}" \
        "https://api.github.com/repos/${GITHUB_REPO}/releases/tags/${tag}" \
        | jq -r '.id // empty')

    if [ -z "$release_id" ]; then
        echo "  ⚠️  找不到 Release for tag ${tag}，跳过"
        continue
    fi

    echo -n "  🗑️  删除 Release ${tag} (id=${release_id}) ... "
    status=$(curl -s -o /dev/null -w "%{http_code}" \
        -X DELETE \
        -H "Authorization: token ${GITHUB_TOKEN}" \
        "https://api.github.com/repos/${GITHUB_REPO}/releases/${release_id}")

    if [ "$status" = "204" ]; then
        echo "✅"
    else
        echo "❌ HTTP ${status}"
    fi
done

# 本地删除 preview tags
echo ""
echo "  删除本地 preview tags..."
for tag in $PREVIEW_TAGS; do
    git tag -d "$tag" 2>/dev/null || true
    echo "    - 已删除本地: $tag"
done

echo ""
echo "=== 清理完成 ==="
