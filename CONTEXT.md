# flickr-cli 领域术语表

## 核心守则

**flickr-cli 是 Flickr 网页版的替代品**，额外提供自动化便利和 agent 友好。能参照 Flickr 官方做法的自行决定，只询问 UX 相关的决策。

## 核心概念

**flickr-cli** — 个人 Flickr 管理 CLI 工具，用于管理自己的照片、相册、备份和迁移。

## 用户

**用户** — 希望用命令行替代网页版 Flickr 的人。一个人可能有多个账户（个人、公司、小号），通过 profile 管理。

**Profile** — 用户的 Flickr 账户凭证集合（API key + OAuth token）。默认 profile 是用户添加的第一个账户，可切换到其他 profile 操作不同账户。

## 命令设计决策

**下载/备份合并** — 将 `backup albums`、`backup user`、`backup id-dirs` 和 `photos download` 合并为统一的 `photos download` 命令，通过 flag 区分行为：
- `--album` / `--album-id` — 从指定相册下载
- `--all` — 下载所有照片
- `--dest` — 目标目录
- `--layout` — 目录布局（flat、album、id-dirs）
- existing files are skipped by default (use `--force` to re-download)
- `--metadata` — 元数据格式（json、csv、both）

**安全门控** — 三级安全机制：
- `--read-only` — 全局开关，阻止所有远程修改（适合脚本/Agent）
- `--dry-run` — 操作级，只预览不执行（优先级高于 --confirm）
- `--confirm` — 操作级，确认高风险操作（删除操作必须）

**缓存** — 本地 SQLite 缓存，用于去重和加速查询。大多数操作自动使用缓存，但也提供显式管理命令：`cache sync`（手动同步元数据）、`cache stats`（查看缓存统计）、`cache cleanup`（清理过期条目）。

**Piwigo 迁移** — 只实现 Piwigo→Flickr 方向。通过 Piwigo REST API（ws.php）读取数据，密码通过 flag 传递（不持久化）。Flickr→Piwigo 方向已有官方 Flickr2Piwigo 插件，不需要实现。

**输出格式** — 三级输出：
- 默认 — 简洁关键信息（ID + 标题）
- `--full` — 完整信息
- `--json` — 自动化/Agent 用

**进度报告** — 默认显示简洁进度（当前/总数 + 文件名），`--quiet` 关闭，`--events` 给 Agent 用。不显示 ETA。

**确认提示** — 删除等危险操作默认弹出交互确认（`Are you sure? [y/N]`），`--confirm` 跳过提示（Agent 用）。

**错误展示** — 人类模式显示友好消息 + 建议操作，`--debug` 显示技术细节。`--json` 模式下错误包含完整结构化信息（code、message、details）。

**API 覆盖范围** — CLI 应覆盖 Flickr API 的所有方法。`api call` 作为兜底入口，用于 Flickr 新增方法但 CLI 尚未适配的场景。

**命令结构** — 按资源类型分组，每个组下放 CRUD 操作，与 Flickr API 方法名对应：
- `auth` — 认证
- `photos` — 照片管理（含 download，合并自 backup）
- `albums` — 相册管理
- `favorites` — 收藏
- `galleries` — 画廊
- `groups` — 群组
- `comments` — 评论
- `contacts` — 联系人
- `stats` — 统计
- `urls` — URL 查找
- `checksums` — 校验和
- `cache` — 缓存
- `piwigo` — 迁移
- `api` — 原始 API 调用
- `doctor` — 诊断
- `version` — 版本

## Flickr API 客户端

**flickr.Client** — 内置的 Flickr API HTTP 客户端。处理 OAuth 1.0a 签名、REST 调用、multipart 上传、分页和重试。

**REST 调用** — 签名 POST 到 `api.flickr.com/services/rest`，带 `method` 参数（如 `flickr.photos.getInfo`）。返回 JSON 包装的 `stat`/`code`/`message` 信封。

**上传** — multipart POST 到 `up.flickr.com/services/upload`。返回 XML 包装的 JSON 类信封。需要 `title` 参数。

**OAuth 1.0a** — 认证协议。三步流程：request token → 用户授权 → access token。支持 OOB（headless 环境）。

**分页** — Flickr 端点返回 `page`/`pages`/`perpage`/`total` 字段。`FetchAll` 自动遍历所有页。

**Endpoint** — Flickr API 操作的 URL（REST、Upload、RequestToken、Authorize、AccessToken）。每个 profile 可覆盖，用于测试。

## 操作

**下载** — 从 Flickr 获取照片到本地磁盘。两种模式：按指定 photo ID（photos 命令内联）或通过备份计划（委托给 backup 包）。

**备份计划** — `backup.BuildPlan()` 枚举要下载的照片（按相册、用户或 ID 目录布局）。`backup.Downloader` 执行下载。

**扫描** — 按扩展名查找本地文件用于上传。

**去重** — 上传前检查照片是否已存在于 Flickr（通过 machine tag 校验和或 Flickr 搜索）。

**相册解析** — 按标题查找现有相册或创建新相册。

**Sidecar** — 与下载照片一起写入的元数据文件（JSON 或 YAML），包含 API 返回的完整照片信息。

## 安全

**Gate** — 安全检查，根据 read-only、dry-run 和 confirm 标志评估是否允许变更操作。

**Risk** — 变更危险等级分类（read、low-write、medium-write、high-write、destructive）。

**Audit** — JSONL 格式的变更操作日志，带时间戳和参数。

## 输出

**Envelope** — 标准 JSON 响应格式：`{ok, data, error, meta}`。

**Renderer** — 向 stdout 输出人类可读文本或 JSON 信封。

**Event** — 向 stderr 输出 NDJSON 进度事件（如 `download_complete`、`upload_started`）。
