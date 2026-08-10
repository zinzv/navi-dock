# NaviDock 前端 UI 设计方案

> 产品品牌：NaviDot（视觉标识） / NaviDock（产品名）  
> 版本：v0.1  
> 状态：草案  
> 更新日期：2026-08-08  
> 关联文档：[技术方案.md](./技术方案.md)

---

## 1. 产品定位

### 1.1 一句话

**NaviDock —— Personal NAS Dashboard**

一个轻量、美观、可自托管的个人数字服务入口。

### 1.2 设计参照

| 参照 | 借鉴点 |
|------|--------|
| macOS Launchpad + Dock | 图标网格、一键启动感 |
| Notion Dashboard | 分组区块、可折叠、低干扰信息架构 |
| Raycast | `Ctrl + K` 极速搜索 |
| NAS 服务中心 | 家庭服务聚合、长期常开 |

### 1.3 适配场景

- NAS 大屏 / 桌面浏览器
- 平板横竖屏
- 手机（简化双列）

### 1.4 设计关键词

**极简 · 卡片化 · 空间感 · 低干扰 · 高自定义**

### 1.5 体验原则

NaviDock **不应该**像传统 NAS 管理面板：

- ❌ 列表堆砌
- ❌ 密表格
- ❌ 后台管理系统观感

而应该像：

```text
          NaviDot

     ●       ●       ●

  Plex   Notes   Books

     我的数字世界入口
```

**核心体验：** 打开浏览器 → 看到自己的数字空间 → 一键进入所有服务。

---

## 2. 整体视觉风格

### 2.1 Design Language

**Modern Glass Dashboard（现代玻璃仪表盘）**

参考气质：macOS · Linear · Arc Browser · Notion

气质描述：深色底上的轻量浮层、细边框、柔和层次，强调「空间」而非「控件密度」。

### 2.2 色彩系统（深色默认）

适合 NAS 长期开启、低眩光：

| Token | 色值 | 用途 |
|-------|------|------|
| `--bg` | `#0F1115` | 页面背景 |
| `--card` | `#181B22` | 卡片 / 浮层底 |
| `--border` | `rgba(255,255,255,0.08)` | 细边框 |
| `--text` | `#FFFFFF` | 主文字 |
| `--text-secondary` | `#8B93A7` | 次级文字 / 描述 |
| `--accent` | 待定（建议冷青或中性高亮，避免默认紫） | 焦点 / CTA |
| `--status-online` | 绿色系 | 在线 |
| `--status-unknown` | 灰色系 | 未知 |
| `--status-offline` | 红色系 | 离线 |

浅色 / Auto 主题 token 在第 14 节定义映射关系。

### 2.3 视觉示意

```text
┌─────────────────────┐
│                     │
│      NaviDot        │
│                     │
│  ┌────┐ ┌────┐      │
│  │Plex│ │Note│      │
│  └────┘ └────┘      │
│                     │
└─────────────────────┘
```

---

## 3. 首页布局设计

### 3.1 Desktop 信息架构

自上而下单一叙事流（非多栏仪表盘堆砌）：

1. Header（品牌 + 搜索入口 + 设置）
2. 分组服务区（主内容）
3. （可选 / 未来）底部 System 状态条

```text
┌────────────────────────────────────┐

             NaviDot

      Search      + Add Service

─────────────────────────────────────

 Media

┌──────┐ ┌──────┐ ┌──────┐
│ Plex │ │Photo │ │Music │
└──────┘ └──────┘ └──────┘

 Knowledge

┌──────┐ ┌──────┐
│Note  │ │Book  │
└──────┘ └──────┘

 System

┌──────┐ ┌──────┐
│ NAS  │ │Docker│
└──────┘ └──────┘

─────────────────────────────────────

              Status
         CPU 25%   RAM 42%

└────────────────────────────────────┘
```

### 3.2 布局约束

| 规则 | 说明 |
|------|------|
| 首屏焦点 | 品牌 NaviDot + 搜索/添加入口，避免首屏堆满统计 |
| 一区一事 | 每个 Group 一块，标题 + 卡片网格 |
| 留白 | 分组间距大于卡片间距，营造空间感 |
| 最大内容宽 | 建议 `max-width: 1200–1440px` 居中，大屏不拉成报纸栏 |

---

## 4. 顶部 Header

### 4.1 规格

| 项 | 值 |
|----|-----|
| 高度 | `72px` |
| 位置 | Sticky 顶部（滚动时保持可点） |
| 背景 | 接近 `--bg`，可微透明 + blur（玻璃感，克制使用） |

### 4.2 布局

```text
┌─────────────────────────────┐
 ● NaviDot          🔍 Search          ⚙
└─────────────────────────────┘
```

| 区域 | 内容 |
|------|------|
| 左 | Logo Dot + 产品名 |
| 中 / 右偏中 | 搜索触发（或仅快捷键提示） |
| 右 | 设置 / 进入编辑模式 |

### 4.3 Logo：NaviDot

设计语义：

- **Dot（●）**：服务节点、入口点
- **Navi**：导航

启动动效建议：

```text
●  →  ○   （轻微呼吸 / 脉冲一次后稳态）
```

要求：克制、可关闭（`prefers-reduced-motion` 时禁用）。

---

## 5. 搜索设计

### 5.1 交互

| 项 | 说明 |
|----|------|
| 快捷键 | `Ctrl + K`（macOS：`⌘ + K`） |
| 形态 | 居中 Modal / Command Palette（类 Raycast） |
| 关闭 | `Esc`、点击遮罩 |

### 5.2 弹窗结构

```text
┌───────────────────────┐
 🔍 Search service

 Plex
 NoteVerse
 Docker
└───────────────────────┘
```

### 5.3 检索范围

- 服务名称
- 描述
- 标签（若有）

示例：输入 `movie` → 命中 Plex / Jellyfin / Emby。

### 5.4 结果行为

- `Enter`：按当前 Network Mode 打开对应 URL
- 方向键：移动高亮
- 可选：结果旁显示分组名、状态点

---

## 6. 服务卡片设计

### 6.1 尺寸

| 断点 | 卡片尺寸 |
|------|----------|
| Desktop | `160 × 120 px`（可微调，保持统一网格） |
| Tablet | 略缩或保持，列数减少 |
| Mobile | 双列自适应宽度 |

### 6.2 结构

```text
┌──────────────┐
       ◉          ← 图标 64px
      Plex        ← 名称
   Media Server   ← 次级描述（可截断）
       🟢         ← 状态（或右上角角标）
└──────────────┘
```

推荐：状态点放在**右上角**，避免与图标抢焦点。

### 6.3 状态 Badge

| 颜色 | 状态 |
|------|------|
| 🟢 绿 | 在线 |
| 灰色 | 未知 / 未检测 |
| 🔴 红 | 离线 |

V1 可仅预留 UI；真实探测见技术方案 V2。

### 6.4 Hover

| 状态 | 表现 |
|------|------|
| 默认 | 静卡：图标 + 名称 |
| Hover | `translateY(-4px)`，阴影增强；可露出 `Open →` 提示 |

```text
正常          hover
┌──────┐     ┌────────┐
│ Plex │  →  │  Plex  │
└──────┘     │ Open → │
             └────────┘
```

动画时长建议 `150–200ms`，缓动 `ease-out`。

### 6.5 点击

- 普通模式：打开服务（`open_type`：`_blank` / `_self`）
- 编辑模式：选中 / 进入编辑，不直接跳转（或需二次确认）

---

## 7. 图标设计

### 7.1 Iconify

示例：`mdi:plex` · `mdi:docker` · `mdi:book-open`

| 项 | 值 |
|----|-----|
| 卡片内显示尺寸 | `64px` |
| 选择器 | `IconPicker` 组件，支持搜索 Iconify 集合 |

### 7.2 自定义上传

| 格式 | png / svg / webp |
|------|------------------|
| 处理 | 圆角、居中裁剪、透明背景保留 |
| 存储 | 后端 `/data/assets/icons`，前端走静态 URL |

---

## 8. 分组设计

### 8.1 Notion 式区块

```text
📁 Media
[ Plex ] [ Jellyfin ] [ Navidrome ]

📁 Knowledge
[ NoteVerse ] [ BookVerse ]
```

### 8.2 能力

| 能力 | 说明 |
|------|------|
| 折叠 | `▼ Media` 展开 / `▶ Development` 收起 |
| 标题 | 可配 Iconify / emoji（克制，默认偏图标） |
| 空状态 | 「本组暂无服务」+ 快捷 Add |

折叠状态可写入本地偏好（Pinia + localStorage），不必强依赖服务端。

---

## 9. 编辑模式

### 9.1 入口

Header 右侧：`⚙` → `Edit`（或设置内「编辑主页」）。

### 9.2 模式差异

| 普通模式 | 编辑模式 |
|----------|----------|
| 点击打开服务 | 显示 ✏ 编辑、× 删除 |
| 无拖拽把手 | 可拖拽排序 |
| 搜索打开服务 | 搜索可定位到卡片 |

```text
普通        编辑
[Plex]      [Plex]
              ✏  ×
```

退出编辑：`Done` / `Esc`（若无未保存冲突）。

---

## 10. 拖拽排序

| 维度 | 说明 |
|------|------|
| 分组排序 | 拖动 Group 标题栏 |
| 卡片排序 | 组内拖动卡片；跨组拖动（可选，V1 建议支持） |
| 技术 | Vue Draggable（如 `vuedraggable` / SortableJS） |
| 反馈 | 拖起时轻微缩放 + 落点占位幽灵块 |

持久化：松手后调用排序 API（debounce 或显式 Save，推荐松手即存）。

---

## 11. 添加 / 编辑服务弹窗

```text
┌───────────────────────┐
 Add Service

 Name
 [ Plex           ]

 Icon
 [ ◉             ]

 Internal URL
 [ http://192... ]

 External URL
 [ https://...   ]

 Group
 [ Media ▼ ]

              Save
└───────────────────────┘
```

字段与技术方案 `nav_item` 对齐：名称、图标、内外网 URL、分组、描述、打开方式等。校验：至少填写一个 URL。

---

## 12. 内外网切换 UI

设置项：**Network Mode**

| 模式 | 行为 |
|------|------|
| ○ Auto | 按访问来源 / CIDR 自动选择 |
| ○ Internal | 强制内网 URL |
| ○ External | 强制外网 URL |

Auto 逻辑示意：

```text
访问 192.168.x.x  → Internal
访问 example.com  → External
```

UI 位置：Settings → General / Network；可选在 Header 放弱化指示（当前模式），避免首屏控件过多。

---

## 13. 设置页面

路由建议：`/settings`

左侧菜单：

```text
Settings
├── General
├── Appearance
├── Groups
├── Services
├── Database
└── Backup
```

| 菜单 | 内容 |
|------|------|
| General | 站点名、Network Mode、语言 |
| Appearance | 主题、卡片密度、是否显示描述/状态 |
| Groups | 分组 CRUD、排序 |
| Services | 列表管理（表格可接受，仅在设置内） |
| Database | 只读状态 / 类型提示（SQLite / PG） |
| Backup | JSON 导入导出 |

原则：**管理密度留在 Settings；首页保持 Launchpad 感。**

---

## 14. 主题系统

| 主题 | 标识 | 说明 |
|------|------|------|
| Dark | 🌙 | 默认，NAS 常开友好 |
| Light | ☀ | 浅色 token 映射 |
| System | — | 跟随 `prefers-color-scheme` |

实现：`html` / root 上挂 `data-theme`，CSS 变量切换；Pinia 持久化用户选择。

---

## 15. 移动端设计

### 15.1 布局

```text
┌───────────┐
 NaviDot
 Search
 Media
┌─────┐
│Plex │
└─────┘
┌─────┐
│Note │
└─────┘
└───────────┘
```

| 项 | 规则 |
|----|------|
| 卡片列数 | **2 列** |
| Header | 压缩高度；搜索可改为全宽入口 |
| 拖拽 | 长按进入排序，或仅在编辑模式开启 |
| 设置 | 左侧菜单改为顶部 Tabs / 抽屉 |

### 15.2 触控

- 卡片点击热区 ≥ 卡片全幅
- Hover 动效在触控设备降级为 `active` 轻微缩放

---

## 16. NAS 状态组件（未来）

首页底部（非 V1 必做）：

```text
System
CPU     ██████░░ 35%
Memory  █████░░░ 42%
Disk    ███████░ 70%
```

约束：作为次级区域，不进首屏英雄位；可折叠或仅大屏显示。

---

## 17. 技术实现

### 17.1 前端技术栈

| 层 | 选型 |
|----|------|
| 框架 | Vue 3 + TypeScript |
| 构建 | Vite |
| 状态 | Pinia |
| 路由 | Vue Router |
| UI 库 | **Naive UI** |
| 样式 | TailwindCSS |
| 动画 | Motion Vue（或 `@vueuse/motion`） |
| 拖拽 | Vue Draggable |
| 图标 | Iconify |

> 说明：与早期技术方案中的 Element Plus 相比，本 UI 方案以 **Naive UI** 为准，更契合深色玻璃仪表盘与轻量对话框气质。技术方案文档应同步更新 UI 库选型。

### 17.2 动效预算

首屏建议至少 2–3 个有意图动效（不过度）：

1. Logo Dot 启动呼吸
2. 卡片 Hover 上浮
3. Search 弹层 / 分组折叠过渡

尊重 `prefers-reduced-motion`。

---

## 18. 前端目录设计

```text
frontend/
└── src/
    ├── layouts/
    │   └── Dashboard.vue
    ├── pages/
    │   ├── Home.vue
    │   └── Settings.vue
    ├── components/
    │   ├── ServiceCard.vue
    │   ├── GroupBlock.vue
    │   ├── SearchDialog.vue
    │   ├── IconPicker.vue
    │   ├── StatusBadge.vue
    │   ├── EditToolbar.vue
    │   └── ThemeSwitcher.vue
    ├── stores/
    │   └── navigation.ts
    ├── api/
    └── assets/
```

---

## 19. UI 组件规划

| 组件 | 功能 |
|------|------|
| `ServiceCard` | 服务卡片、Hover、编辑态操作 |
| `GroupBlock` | 分组标题、折叠、网格容器 |
| `SearchDialog` | `Ctrl+K` 命令面板 |
| `IconPicker` | Iconify + 上传 |
| `StatusBadge` | 在线 / 未知 / 离线 |
| `EditToolbar` | 进入/退出编辑、批量操作 |
| `ThemeSwitcher` | Dark / Light / System |
| `AddServiceModal` | 新增/编辑服务表单 |
| `NetworkModeSelect` | Auto / Internal / External |

---

## 20. 最终视觉定位（验收标准）

打开首页时，应满足：

1. **品牌优先**：去掉 Header 导航后，仍能一眼认出是 NaviDot 空间，而不是通用后台。
2. **低干扰**：首屏没有表格、统计墙、多块营销信息。
3. **一键到达**：任意服务 ≤ 两次点击（搜索则一次键盘 + Enter）。
4. **可自定义**：分组、图标、内外网、主题均可改，且不破坏整体气质。

一句话验收：

> **这是我的数字世界入口，不是又一个 NAS 管理后台。**

---

## 附录 A. 与技术方案的对齐点

| UI 概念 | 后端 / 数据 |
|---------|-------------|
| Service Card | `nav_item` |
| Group Block | `nav_group` |
| Internal / External URL | `internal_url` / `external_url` |
| Network Mode Auto | CIDR / RemoteAddr 解析 |
| Iconify / Upload | `icon_type` + `icon_value` |
| Backup | JSON Import / Export API |
| Status Badge | V2 健康检测 |

## 附录 B. 命名说明

| 名称 | 用途 |
|------|------|
| **NaviDock** | 产品名（仓库 / Docker / 对外） |
| **NaviDot** | 品牌标识与 Logo 文案（界面主标题） |

原草案名「NavVerse / 导航元」可视为早期代号；UI 与产品对外统一使用 NaviDock / NaviDot。

---

*本文档为前端视觉与交互规范草案，实现以组件代码与设计 Token 为准。*
