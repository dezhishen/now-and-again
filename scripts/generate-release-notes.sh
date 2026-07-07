#!/usr/bin/env bash
# =============================================================================
# generate-release-notes.sh
# 
# 智能生成 Release Notes：
#   1. 正式版 (v1.0.7) → 对比上一个正式版 (v1.0.6)，跳过 preview 版本
#   2. preview 版 (v1.0.7-preview-004) → 对比上一个 preview (v1.0.7-preview-003)
#
# 用法:
#   ./scripts/generate-release-notes.sh <current_ref> <github_repo> <github_repo_owner>
#
# 输出: /tmp/release-notes.md
# =============================================================================

set -euo pipefail

CURRENT_REF="${1:-}"
GITHUB_REPO="${2:-}"
GITHUB_REPO_OWNER="${3:-}"

if [ -z "$CURRENT_REF" ]; then
    echo "ERROR: CURRENT_REF is required"
    exit 1
fi

REPO_FULL="${GITHUB_REPO_OWNER}/${GITHUB_REPO#*/}"
REPO_FULL="${REPO_FULL#/}"

# ── 判断当前 tag 类型 ──────────────────────────────────────────
is_stable() {
    local ref="$1"
    # 正式版: v<num>.<num>.<num> 无后缀
    [[ "$ref" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]
}

is_preview() {
    local ref="$1"
    # preview: v<num>.<num>.<num>-preview-<NNN>
    [[ "$ref" =~ ^v[0-9]+\.[0-9]+\.[0-9]+-preview-[0-9]+$ ]]
}

is_patch() {
    local ref="$1"
    # patch/hotfix: v<num>.<num>.<num>-<N>
    [[ "$ref" =~ ^v[0-9]+\.[0-9]+\.[0-9]+-[0-9]+$ ]]
}

tag_type() {
    local ref="$1"
    if is_stable "$ref"; then
        echo "stable"
    elif is_preview "$ref"; then
        echo "preview"
    elif is_patch "$ref"; then
        echo "patch"
    else
        echo "unknown"
    fi
}

# ── 提取版本号 ────────────────────────────────────────────────

# 提取主版本号 e.g. v1.0.7-preview-004 → v1.0.7
extract_base_version() {
    local ref="$1"
    echo "$ref" | sed -E 's/^v([0-9]+\.[0-9]+\.[0-9]+).*$/v\1/'
}

# 提取 preview 编号 e.g. v1.0.7-preview-004 → 004
extract_preview_num() {
    local ref="$1"
    echo "$ref" | sed -E 's/^v[0-9]+\.[0-9]+\.[0-9]+-preview-([0-9]+)$/\1/'
}

# ── 查找对比基准 tag ──────────────────────────────────────────

find_prev_tag() {
    local current="$1"
    local ttype
    ttype=$(tag_type "$current")

    case "$ttype" in
    stable)
        # 正式版: 找上一个正式版 (按版本号排序，不是时间)
        local base
        base=$(extract_base_version "$current")
        git tag -l 'v[0-9]*.[0-9]*.[0-9]' \
            | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' \
            | sort -V \
            | grep -B1 "^${current}$" \
            | head -1 \
            | grep -v "^${current}$" || echo ""
        ;;
    preview)
        # preview: 找上一个同版本的 preview
        local base
        base=$(extract_base_version "$current")
        local num
        num=$(extract_preview_num "$current")
        if [ -n "$num" ] && [ "$num" -gt 1 ] 2>/dev/null; then
            local prev_num=$((10#$num - 1))
            local prev_tag
            prev_tag=$(printf "%s-preview-%03d" "$base" "$prev_num")
            if git rev-parse "$prev_tag" >/dev/null 2>&1; then
                echo "$prev_tag"
            else
                # 同版本没有上一个 preview，找上一个正式版
                git tag -l 'v[0-9]*.[0-9]*.[0-9]' \
                    | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' \
                    | sort -V \
                    | grep -B1 "^${base}$" \
                    | head -1 \
                    | grep -v "^${base}$" || echo ""
            fi
        else
            # preview-001 对比上一个正式版
            git tag -l 'v[0-9]*.[0-9]*.[0-9]' \
                | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' \
                | sort -V \
                | grep -B1 "^${base}$" \
                | head -1 \
                | grep -v "^${base}$" || echo ""
        fi
        ;;
    *)
        # 按时间找上一个 tag
        git describe --tags --abbrev=0 HEAD~ 2>/dev/null || echo ""
        ;;
    esac
}

# ── 查找所有相关 preview tags (用于清理) ─────────────────────

find_preview_tags_for_version() {
    local base_version="$1"
    git tag -l "${base_version}-preview-*" 2>/dev/null || true
}

# ── 导出函数供 release.yml 使用 ───────────────────────────────

# 如果脚本被 source 而不是直接执行，导出函数
if [[ "${BASH_SOURCE[0]}" != "${0}" ]]; then
    return 0
fi

# ── Main: 生成 Release Notes ──────────────────────────────────

echo "=== 生成 Release Notes ==="
echo "  Current: ${CURRENT_REF}"
echo "  Type:    $(tag_type "$CURRENT_REF")"

PREV_TAG=$(find_prev_tag "$CURRENT_REF")
echo "  Prev:    ${PREV_TAG:-<none> (first release)}"

# 计算 commit range
if [ -z "$PREV_TAG" ]; then
    RANGE="HEAD"
else
    RANGE="${PREV_TAG}..HEAD"
fi

# Docker tag 去掉 v 前缀
DOCKER_TAG="${CURRENT_REF#v}"
IS_STABLE=false
if is_stable "$CURRENT_REF"; then
    IS_STABLE=true
fi

# 计算额外浮动 tag
FLOATING_TAG=""
TTYPE=$(tag_type "$CURRENT_REF")
if [ "$TTYPE" = "preview" ]; then
    # 1.0.7-preview-004 → 1.0.7-preview (浮动 latest-preview)
    FLOATING_TAG="${DOCKER_TAG%-preview-*}-preview"
elif [ "$TTYPE" = "patch" ]; then
    # 1.0.7-1 → 1.0.7 (更新主版本镜像)
    FLOATING_TAG="${DOCKER_TAG%-*}"
fi

# ── 生成 Release Notes ────────────────────────────────────────
NOTES_FILE="/tmp/release-notes.md"

echo "## 📦 变更内容" > "$NOTES_FILE"
echo "" >> "$NOTES_FILE"
if [ -n "$PREV_TAG" ]; then
    echo "> 对比: \`${PREV_TAG}\` → \`${CURRENT_REF}\`" >> "$NOTES_FILE"
else
    echo "> 首次发布" >> "$NOTES_FILE"
fi
echo "" >> "$NOTES_FILE"

# 分类提取 commits
collect_section() {
    local title="$1"
    local grep_pattern="$2"
    local commits
    commits=$(git log "${RANGE}" --pretty=format:"- %s" --grep="${grep_pattern}" 2>/dev/null | head -50)
    if [ -n "$commits" ]; then
        echo "### ${title}" >> "$NOTES_FILE"
        echo "$commits" >> "$NOTES_FILE"
        echo "" >> "$NOTES_FILE"
    fi
}

collect_section "🚀 新功能" "^feat"
collect_section "🐛 修复" "^fix"
collect_section "⚡ 性能优化" "^perf"
collect_section "♻️ 重构" "^refactor"
collect_section "🔧 杂项" "^chore\|^docs\|^style\|^ci\|^test\|^build"

# 未分类的
OTHER=$(git log "${RANGE}" --pretty=format:"- %s" --invert-grep --grep="^feat\|^fix\|^perf\|^refactor\|^chore\|^docs\|^style\|^ci\|^test\|^build\|^Merge" 2>/dev/null | head -50)
if [ -n "$OTHER" ]; then
    echo "### 📋 其他" >> "$NOTES_FILE"
    echo "$OTHER" >> "$NOTES_FILE"
    echo "" >> "$NOTES_FILE"
fi

echo "---" >> "$NOTES_FILE"
echo "## 🐳 Docker 镜像" >> "$NOTES_FILE"
echo "" >> "$NOTES_FILE"
echo "| 镜像 | 标签 |" >> "$NOTES_FILE"
echo "|------|------|" >> "$NOTES_FILE"

if [ "$IS_STABLE" = true ]; then
    echo "| \`ghcr.io/${REPO_FULL}\` | \`${DOCKER_TAG}\`, \`latest\` |" >> "$NOTES_FILE"
    echo "| \`ghcr.io/${REPO_FULL}-cli\` | \`${DOCKER_TAG}\`, \`latest\` |" >> "$NOTES_FILE"
elif [ -n "$FLOATING_TAG" ]; then
    echo "| \`ghcr.io/${REPO_FULL}\` | \`${DOCKER_TAG}\`, \`${FLOATING_TAG}\` |" >> "$NOTES_FILE"
    echo "| \`ghcr.io/${REPO_FULL}-cli\` | \`${DOCKER_TAG}\`, \`${FLOATING_TAG}\` |" >> "$NOTES_FILE"
else
    echo "| \`ghcr.io/${REPO_FULL}\` | \`${DOCKER_TAG}\` |" >> "$NOTES_FILE"
    echo "| \`ghcr.io/${REPO_FULL}-cli\` | \`${DOCKER_TAG}\` |" >> "$NOTES_FILE"
fi

echo "" >> "$NOTES_FILE"
echo '```bash' >> "$NOTES_FILE"
echo "# 服务端" >> "$NOTES_FILE"
echo "docker pull ghcr.io/${REPO_FULL}:${DOCKER_TAG}" >> "$NOTES_FILE"
echo "# CLI" >> "$NOTES_FILE"
echo "docker pull ghcr.io/${REPO_FULL}-cli:${DOCKER_TAG}" >> "$NOTES_FILE"
echo '```' >> "$NOTES_FILE"
echo "" >> "$NOTES_FILE"

echo "---" >> "$NOTES_FILE"
echo "## 📥 直接下载" >> "$NOTES_FILE"
echo "" >> "$NOTES_FILE"

BASE_URL="https://github.com/${REPO_FULL}/releases/download/${CURRENT_REF}"
echo "| 平台 | 架构 | 服务端 | CLI |" >> "$NOTES_FILE"
echo "|------|------|--------|-----|" >> "$NOTES_FILE"
for os in linux darwin windows; do
    for arch in amd64 arm64; do
        case "$os-$arch" in
            windows-arm64) continue ;;
            linux-*)    os_label="Linux" ;;
            darwin-*)   os_label="macOS" ;;
            windows-*)  os_label="Windows" ;;
        esac
        srv="now-and-again_${os}_${arch}.tar.gz"
        cli="na_${os}_${arch}.tar.gz"
        echo "| ${os_label} | ${arch} | [${srv}](${BASE_URL}/${srv}) | [$cli](${BASE_URL}/${cli}) |" >> "$NOTES_FILE"
    done
done

echo "" >> "$NOTES_FILE"
echo "<details><summary>SHA256 校验</summary>" >> "$NOTES_FILE"
echo "" >> "$NOTES_FILE"
echo '```' >> "$NOTES_FILE"
if [ -f dist/sha256sum.txt ]; then
    cat dist/sha256sum.txt >> "$NOTES_FILE"
else
    echo "(构建时生成)" >> "$NOTES_FILE"
fi
echo '```' >> "$NOTES_FILE"
echo "</details>" >> "$NOTES_FILE"
echo "" >> "$NOTES_FILE"

echo "---" >> "$NOTES_FILE"
if [ -n "$PREV_TAG" ]; then
    echo "**完整对比**: [${PREV_TAG}...${CURRENT_REF}](https://github.com/${REPO_FULL}/compare/${PREV_TAG}...${CURRENT_REF})" >> "$NOTES_FILE"
else
    echo "**首次发布**" >> "$NOTES_FILE"
fi

echo ""
echo "=== Release Notes 已生成: $NOTES_FILE ==="
cat "$NOTES_FILE"
