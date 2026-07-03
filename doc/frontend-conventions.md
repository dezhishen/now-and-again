# 前端开发约定

> 本文档跟随项目 Git 仓库，所有开发者必须遵守。

## 按钮

**必须使用全局样式类，禁止裸 Tailwind：**

| 用途 | 类名 |
|---|---|
| 主操作（创建/保存/确认） | `btn-primary` |
| 次要操作（取消/返回） | `btn-secondary` |
| 危险操作（删除） | `btn-danger` |
| 尺寸 | 附加 `text-sm` 或 `text-xs` |

```html
<!-- ✅ 正确 -->
<button class="btn-primary text-sm">创建</button>
<button class="btn-secondary text-sm">取消</button>

<!-- ❌ 禁止 -->
<button class="px-3 py-1 rounded-md bg-green-500 hover:bg-green-600 text-white">创建</button>
```

## 输入框

**必须使用 `input` 全局类：**

```html
<!-- ✅ 正确 -->
<input class="input" />

<!-- ❌ 禁止 -->
<input class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
```

## 页面宽度

| 页面类型 | max-w | 宽度 |
|---|---|---|
| 管理面板（含表格） | `max-w-7xl` | 1280px |
| 卡片列表页（首页/家庭管理） | `max-w-5xl` | 1024px |
| 表单页（个人中心/API Key） | `max-w-3xl` | 768px |
| 登录/注册 | `max-w-md` | 448px |

所有页面容器统一 `mx-auto p-4`。

## 卡片

内容区块使用 `card` 类（含 padding、圆角、阴影、暗色模式）。

同一页面的不同标签页内容必须使用一致的包装方式（如统一用 `card` 包裹），避免切换时宽度跳跃。

## 标签页

项目使用 **3 级页签体系**，通过全局 CSS 类统一视觉层级。**必须使用以下类，禁止裸 Tailwind 拼凑页签样式。**

### L1 — 主导航（`.nav-item`）

用于侧边栏等一级导航入口，圆角药丸 + 背景高亮。

```html
<button class="nav-item" :class="{ active: current === 'xxx' }">
  📊 概览
</button>
```

| 状态 | 视觉 |
|------|------|
| 默认 | `text-gray-700 dark:text-gray-300` |
| Hover | `bg-primary text-white` |
| 激活 | `bg-primary/10 text-primary font-medium` |

### L2 — 布局页签栏（`.tab-bar` + `.tab`）

用于 FamilyView 顶部多页签栏，带灰色底色条，是页面布局的一部分。

```html
<div class="tab-bar">
  <button class="tab" :class="{ active: current === 'xxx' }">
    📊 概览
    <span class="…">✕</span>  <!-- 关闭按钮（可选） -->
  </button>
</div>
```

| 状态 | 视觉 |
|------|------|
| 默认 | `border-transparent text-gray-400` |
| Hover | `text-gray-600` |
| 激活 | `border-primary text-primary font-medium` |

### L3 — 内容页签（`.tabs` + `.tab`）

用于页面内部子视图切换（如管理面板、成员列表），最轻量的下划线样式。

```html
<div class="tabs">
  <button class="tab" :class="{ active: current === 'xxx' }">
    标签A
  </button>
  <button class="tab" :class="{ active: current === 'yyy' }">
    标签B
  </button>
</div>
```

### 紧凑变体（`.tabs-sm` + `.tab-sm`）

用于弹窗等空间受限场景，字号更小（`text-xs`）。

```html
<div class="tabs-sm">
  <button class="tab-sm" :class="{ active: current === 'xxx' }">标签</button>
</div>
```

### 激活状态

**统一使用 `:class="{ active: ... }"` 对象语法**，不再用三元表达式拼接字符串：

```html
<!-- ✅ 正确 -->
<button class="tab" :class="{ active: current === 'xxx' }">标签</button>

<!-- ❌ 禁止 -->
<button :class="current === 'xxx' ? 'border-primary text-primary' : 'border-transparent text-gray-400'">标签</button>
```

## 弹窗

使用 `<Teleport to="body">` + 固定定位遮罩，取消/确认按钮用 `btn-secondary` / `btn-primary`。

## 图标操作按钮

小型图标按钮（✎/✕/⭐）可用裸 Tailwind，仅限图标无文字场景：

```html
<button class="px-2 py-1 text-xs rounded hover:bg-gray-100 text-gray-500">✎</button>
<button class="px-2 py-1 text-xs rounded hover:bg-red-50 text-red-500">✕</button>
```
