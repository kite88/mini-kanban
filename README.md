**简体中文** | [English](./README.en.md)

# 拖拽式任务看板（Next.js + @dnd-kit / Go Gin + GORM / PostgreSQL / Redis）

[![Release](https://img.shields.io/github/v/release/kite88/mini-kanban?v=2)](https://github.com/kite88/mini-kanban/releases/latest)
[![License](https://img.shields.io/github/license/kite88/mini-kanban?v=2)](LICENSE)

包含登录、用户 CRUD、部门管理、项目 CRUD、任务 CRUD、四列拖拽看板、附件上传、回收箱（软删除）与操作日志（含物理删除快照）。

## 界面预览

**登录**（支持中英文与浅色 / 深色主题）

![登录页](docs/screenshots/login.png)

**四列拖拽看板**（待办 / 进行中 / 已完成 / 已验收，任务卡带负责人、附件数与截止时间，可按「全部 / 我创建的 / 指派给我的」筛选）

![四列拖拽看板](docs/screenshots/board.png)

**任务详情**（富文本描述、附件列表、负责人 / 所在列 / 截止时间）

![任务详情](docs/screenshots/task-modal.png)

**项目列表与项目资料**（资料支持富文本、内嵌图片、附件与按标题检索）

![项目列表](docs/screenshots/projects.png)

![项目资料](docs/screenshots/project-docs.png)

**回收箱**（软删除数据的恢复 / 彻底删除，项目条目带同批任务数）

![回收箱](docs/screenshots/trash.png)

**用户管理与操作日志**（管理员）

![用户管理](docs/screenshots/users.png)

![操作日志](docs/screenshots/logs.png)

**深色主题**

![深色主题看板](docs/screenshots/board-dark.png)

## 目录结构

```
server/                                    Go Gin 后端（模块名 mini-kanban）
  main.go                                  启动入口：加载配置 / 未配置时进入初始化向导
  config.example.yaml                      配置模板（入库）
  config.yaml                              本地配置（.gitignore 忽略）
  internal/config                          配置加载、校验、向导渲染写回
  internal/setup                           首次启动的配置向导（页面 + 探测 + 写回）
  internal/db                              GORM 连接、建表 / 补列 / 补索引、初始管理员
  internal/cache                           Redis 封装（Get/Set/DeletePrefix）
  internal/i18n                            接口文案多语言：语言协商 + 英文目录（以中文原文为键）
  internal/storage                         附件落盘、扩展名白名单、随机文件名
  internal/models                          User / Department / Project / Task /
                                           ProjectMember / TaskAttachment / OperationLog
  internal/middleware                      Auth(JWT)、Admin、CORS、Redis GET 缓存
  internal/handlers                        auth / user / project / task / project_doc /
                                           department / trash / file / log
  internal/oplog                           操作日志写入 + 物理删除的数据快照
  internal/router                          路由与中间件集中注册
  internal/utils                           JWT、邮箱与密码校验、文件名清理
  webstatic.go                             托管 web-dist 目录里的前端静态站点（NoRoute + SPA 兜底）
  web-dist/                                前端静态导出产物目录（构建生成，.gitignore 忽略）
  data/uploads/                            附件存放目录（storage.dir，.gitignore 忽略）

build.sh                                   交叉编译并打包可分发归档（带 web-dist 目录，产物名带版本号）
build-web-static.sh                        前端静态导出并拷进 server/web-dist（build.sh 的前置步骤）
build-web-static.bat                       ↑ 的 Windows 版（GBK 编码，可双击）

tools/release/                             归档里的启动脚本（start.sh / start.bat）
docs/screenshots/                          README 用的界面截图
docs/release-notes/                        各版本 Release 说明（推标签时自动取用）

web/                                       Next.js 前端（App Router + TS + Tailwind）
  app/login                                登录
  app/                                     首页（按登录态跳转 /projects 或 /login）
  app/projects                             项目列表（增删改）
  app/projects/[id]                        看板（四列拖拽 + 附件 + 富文本）
  app/departments                          部门管理（管理员）
  app/users                                用户管理（管理员）
  app/trash                                回收箱（恢复 / 彻底删除）
  app/logs                                 操作日志（管理员，含数据快照查看）
  components/Board.tsx                     @dnd-kit 跨列 / 同列排序
  components/                              AuthProvider / AuthGuard / NavBar / TaskCard /
                                           TaskModal / ProjectModal / ProjectDocsModal /
                                           MemberPicker / AttachmentList / RichTextEditor /
                                           Modal / ConfirmProvider / UserSelect /
                                           ThemeProvider / ThemeToggle /
                                           I18nProvider / LanguageToggle
  lib/i18n/                                多语言：zh-CN.ts（源语言 + 全部键）/
                                           en.ts / index.ts（协商、插值、持久化）
  lib/labels.ts                            业务枚举 → 文案（列名 / 角色 / 日志类型与动作）
  lib/api.ts                               请求封装、token(localStorage)、fileUrl
  lib/types.ts                             类型与看板列定义（COLUMNS）
  lib/html.ts · date.ts · validate.ts      富文本处理 / 日期 / 表单校验
```

## 下载与部署（最快路径）

发布包在 [Releases](https://github.com/kite88/mini-kanban/releases/latest)，**每个版本只有一种产物**：
一个可执行文件 + 一个 `web-dist/` 页面目录，由同一个进程、同一个端口提供接口与页面，
目标机不需要 Node、不需要额外的静态目录，也不需要 nginx。

| 平台 | 产物 |
| --- | --- |
| Windows x64 / ARM64 | `mini-kanban-v1.0.0-windows-amd64.zip` / `mini-kanban-v1.0.0-windows-arm64.zip` |
| Linux x64 / ARM64 | `mini-kanban-v1.0.0-linux-amd64.tar.gz` / `mini-kanban-v1.0.0-linux-arm64.tar.gz` |
| macOS Intel / Apple Silicon | `mini-kanban-v1.0.0-darwin-amd64.tar.gz` / `mini-kanban-v1.0.0-darwin-arm64.tar.gz` |

> **文件名里的版本号就是发布版本号**（例如 `v1.0.0`），顶层目录与压缩包同名，包内还有一份 `version.txt`；
> 启动脚本启动时会打印版本号，所以下载的是哪一版一目了然。`checksums.txt` 是全部归档的 SHA256，
> 可用 `sha256sum -c checksums.txt` 校验。

三步跑起来：

1. **完整解压**对应平台的归档（`web-dist/` 必须与可执行文件在同一目录）；
2. 运行 `start.sh`（Windows 双击 `start.bat`）；
3. 浏览器打开 <http://localhost:8281>。目录里没有 `config.yaml` 时会自动进入初始化向导
   （<http://localhost:8281/setup>），填好数据库 / Redis 并保存，配置写入后弹窗显示服务就绪倒计时，
   3.5 秒后自动打开状态页。

前置条件只有一个：**PostgreSQL 数据库需要提前创建**（服务只建表不建库）；Redis 可选，留空即关闭缓存。

想自己打包 / 改前端后再打包，见下面的「部署」章节；只想跑起来的话上面这三步就够了。

## 环境要求

- Go 1.21+
- Node.js 18.17+（Next.js 14 要求）
- PostgreSQL（**数据库需提前创建**，服务只建表不建库）
- Redis（**可选**：`redis.host` 留空即关闭缓存，服务照常运行）

## 快速开始

> **命令约定**：标 `bash` 的代码块在 macOS / Linux 终端可直接执行，Windows 用户在 PowerShell 里同样可用（`cd`、`go run .`、`npm run dev` 这类命令是通用的）；
> 存在平台差异的操作（复制文件、交叉编译）会分别给出两版。

### 方式一：初始化配置向导（推荐首次部署）

不提供 `config.yaml` 时，后端会先启动一个只用于填写配置的临时服务：

```bash
cd server
go run .
```

浏览器打开 `http://localhost:8281/setup`（端口可用 `SETUP_PORT` 指定；在服务器上启动时，日志会打印可远程访问的 IP）。

向导提供：数据库 / Redis 连接测试、配置校验与写回（按 `config.example.yaml` 模板渲染，**保留注释与排版**）。数据库连不上会拦下保存，可勾选忽略；Redis 连不上只提示，缓存降级运行。保存成功后向导自动关闭，业务服务按正式配置启动。

重新配置：用 `SETUP_FORCE=1` 启动，原配置会备份为 `config.yaml.bak`。已配置完成后访问 `/setup` 会看到「服务运行正常」的反馈页。

> 注意：向导表单只包含 `server` / `database` / `redis` / `jwt` / `seed` 五段；`storage` 与 `database.logLevel` 刻意不进表单，向导写回时按模板值补齐，需要修改请直接编辑 `config.yaml`。

### 方式二：手动复制模板

```bash
# macOS / Linux
cd server
cp config.example.yaml config.yaml

cd ../web
cp .env.example .env.local
```

```powershell
# Windows PowerShell
cd server
Copy-Item config.example.yaml config.yaml

cd ../web
Copy-Item .env.example .env.local
```

## 配置文件

| 文件 | 作用 |
| --- | --- |
| `server/config.yaml` | 后端配置（**已被 .gitignore 忽略**，不入库）：端口/模式、PostgreSQL、Redis 与缓存时长、JWT、初始管理员、附件存储、日志级别 |
| `server/config.example.yaml` | 配置模板，同时是初始化向导写回配置时的底稿（向导会保留其注释与排版） |
| `web/.env.local` | 前端本地配置（**已被 .gitignore 忽略**，不入库）：`NEXT_PUBLIC_API_BASE`（后端接口地址） |
| `web/.env.example` | 前端环境变量模板，首次使用复制为 `.env.local` |

`server/config.yaml` 完整字段：

```yaml
server:
  port: "8281"          # 必填。HTTP 监听端口（字符串）
  mode: "release"       # 必填。debug | release | test，其他值启动报错

database:
  # dsn 非空时优先使用（此时下面 host/port/user/dbname/sslmode 不再校验）
  dsn: ""               # 例：postgres://postgres:123456@localhost:5432/mini-kanban?sslmode=disable
  host: "localhost"     # dsn 留空时必填
  port: 5432            # dsn 留空时必填
  user: "postgres"      # dsn 留空时必填
  password: "123456"    # dsn 留空时必填；含特殊字符无需手工转义（拼装时自动 URL 编码）
  dbname: "mini-kanban" # dsn 留空时必填；库需提前创建，服务只建表不建库
  sslmode: "disable"    # dsn 留空时必填
  preferSimpleProtocol: false  # 出现 pgx 协议兼容问题时设为 true（关闭扩展查询协议）
  logLevel: "warn"             # silent | error | warn | info，其他值启动报错；向导不展示

redis:
  # 可选组件：host 留空即关闭缓存（其余字段不再校验，服务照常运行）
  host: "localhost"
  port: 6379            # host 非空时必填
  password: ""          # 无密码留空
  db: 0
  cacheTTL: "30s"       # host 非空时必填且需 > 0：项目 / 任务列表缓存时长
  userCacheTTL: "1m"    # host 非空时必填且需 > 0：用户 / 部门列表缓存时长

jwt:
  secret: "换成足够随机的长字符串"   # 必填；修改后已签发的 token 全部失效
  ttlHours: 24                     # 必填且需 > 0；token 有效期（小时）

seed:
  # 仅在「用户表为空」时写入初始管理员；写入成功后自动清空本段（避免明文密码留在磁盘）
  enabled: false        # 设为 true 时下面三项必填
  adminEmail: ""        # 初始管理员登录邮箱
  adminUsername: ""     # 初始管理员显示名
  adminPassword: ""     # 初始管理员密码，建议 8 位以上

storage:
  # 附件与富文本内嵌图片目录；向导不展示，只能在本文件里维护
  dir: "data/uploads"   # 必填；相对路径按后端工作目录解析，不存在会自动创建（需可写）
  maxSizeMB: 10         # 必填且需 > 0；单次最多 10 个文件，请求体上限 = maxSizeMB × 10
```

加载规则：

1. 查找顺序：`CONFIG_FILE` 环境变量 → `config.yaml` → `config.yml` → `configs/config.yaml` → `server/config.yaml` → `../config.yaml` → `../server/config.yaml`（因此从仓库根目录或 `server/` 目录启动都能找到）
2. 环境变量优先级更高：`PORT`、`DB_DSN`、`REDIS_ADDR`（`host:port`）、`REDIS_PASSWORD`、`JWT_SECRET`
3. **没有代码内兜底值**：必需项缺失或格式非法会直接报错退出，不会静默使用默认值；`redis.host` 留空是「关闭缓存」的合法写法
4. 其他环境变量：`SETUP_FORCE`（强制重新进入初始化向导，原配置备份为 `.bak`）、`SETUP_PORT`（向导自身端口，默认 8281）、`SNOWFLAKE_NODE_ID`（雪花 ID 节点号，0~1023；**多实例部署时必须各不相同**，否则各实例生成的 ID 会冲突，单机保持默认 0 即可）

更详细的逐字段说明（含取值范围与注意事项）见 `server/config.example.yaml`。

前端 `web/.env.local`：

```bash
NEXT_PUBLIC_API_BASE=http://localhost:8281/api   # 端口需与 server.port 一致
```

> `NEXT_PUBLIC_*` 是**构建期内联**的：改了地址必须重新 `npm run build`，运行时再设置环境变量无效（详见「部署」章节）。

## 启动

后端：

```bash
cd server
go run .              # 或先编译：go build -o mini-kanban .
```

启动日志会打印端口、数据库、Redis 与缓存状态；`GET /healthz` 返回服务与 Redis 状态。

前端：

```bash
cd web
npm install
npm run dev           # http://localhost:3201
```

如后端不在本机 8281，修改 `web/.env.local`：

```bash
NEXT_PUBLIC_API_BASE=http://localhost:8281/api
```

## 初始管理员账号

默认不写入任何账号（`seed.enabled: false`）。需要初始化时先开启并填好三项：

```yaml
seed:
  enabled: true
  adminEmail: "admin@mail.example.com"
  adminUsername: "Admin"
  adminPassword: "至少 8 位"
```

规则：

- 仅当**用户表为空**时创建；已有用户则跳过，启动日志会打印创建的邮箱
- 创建成功后会自动把 `seed` 段清空（`enabled=false`，邮箱 / 用户名 / 密码置空），**避免初始密码长期以明文留在磁盘上**；需要再次 seed 时重新填写
- 其余成员账号由管理员在「用户管理」页面创建

## 附件存储

任务附件、项目资料附件与富文本内嵌图片统一落在 `storage.dir`：

- **单次最多 10 个文件**，整个请求体上限 = `storage.maxSizeMB × 10`
- 扩展名白名单：图片（png / jpg / jpeg / gif / webp / bmp / ico）、文档（pdf / txt / md / csv / doc / docx / xls / xlsx / ppt / pptx）、压缩包（zip / 7z / rar / gz）、音视频（mp4 / mp3 / wav）。**`.svg` 不允许**（可携带脚本）
- 落盘文件名由服务端生成：32 位随机 hex + 白名单扩展名，用户提供的文件名不进入路径（防目录穿越）
- 对外地址为 `/api/files/<32位随机名>`，前端用 `fileUrl()` 按 API 基址补全
- 下载路由 **需要登录态**，但 URL 里带的**不是登录 token**，而是一张专用的「下载票据」：登录后前端调 `POST /api/files/ticket` 换取（`aud=file`，默认 2 小时），再拼成 `/api/files/<随机名>?ticket=...`
- 这样隔离的原因是富文本 `<img src>` 与附件链接由浏览器直接发起、带不了 `Authorization` 头，凭据只能进 URL：票据只能读附件、不能调用其他接口，所以下载地址被转发 / 记录也不等于登录态泄露。登录 token 不会出现在任何 URL 中
- 票据 JTI 与登录会话的 jti 相同：登出后登录 token 进 Redis 黑名单，票据随之立即失效
- **下载还要过数据权限**（票据只说明「谁在下载」，不说明「能不能下这个文件」）：
  服务端按文件的引用关系反查归属，再套用对应入口的可见性，原则是**「即看即下」**——
  - 资料附件、资料正文里的内嵌图片：按**资料所属项目的可见性**（能看到项目就能看资料）；
  - 任务附件、任务描述里的内嵌图片：按**任务可见性**（管理员 / 任务创建人 / 被指派人 / 项目创建人），
    即看板里看得见这个任务的人才能下它的附件。

  引用它的数据已被删除（资料 / 任务软删）时一律 404；还没被任何数据引用的临时文件
  （编辑器里刚粘贴的图片、上传后尚未保存的附件）只允许**上传者本人**下载 —— 上传时会在 `uploaded_files`
  登记上传者。鉴权通过但没有数据权限时同样返回 404，不泄露文件是否存在
- 票据用**独立密钥**签名：密钥由 `jwt.secret` 经 HMAC-SHA256 单向派生（标签 `mini-kanban/file-ticket/v1`），不新增配置项。因此登录 token 过不了下载校验、票据也过不了登录校验，即使某一侧漏检 `aud` 也无法互换使用；`aud=file` 与 `AuthRequired` 的显式拒绝作为纵深防御
- 两类凭据统一限定 `HS256` 且必须带 `exp`（防算法混淆与永不过期票据）
- 前端 `AuthProvider` 负责换取票据并在临近过期时静默续期（`fileUrl()` / `resolveFileUrls()` 自动拼上当前票据，`TaskModal` 在票据续期后刷新描述里的图片地址）；富文本保存前 `toStoredHtml()` 会把票据剥掉，入库的仍是纯 `/api/files/<随机名>`
- 响应带 `X-Content-Type-Options: nosniff`；鉴权失败为 `Cache-Control: no-store`，正常响应为 `private, max-age=3600`（避免被代理 / CDN 缓存后绕过鉴权）。图片 / 视频内联展示，其他类型强制下载并尽量还原原始文件名

## 接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | /api/auth/login | 登录，返回 JWT |
| GET | /api/auth/me | 当前用户 |
| POST | /api/auth/logout | 登出（token 加入 Redis 黑名单） |
| PUT | /api/auth/profile | 修改本人资料（用户名 / 邮箱 / 密码，不含角色）；前端暂无入口 |
| GET | /api/files/:name | 附件下载（需登录 + 数据权限：`?ticket=`，见「附件存储」） |
| POST | /api/files/ticket | 换取附件下载票据（登录态，默认 2 小时） |
| GET | /api/departments | 部门列表 |
| GET | /api/departments/:id | 部门详情 |
| POST | /api/departments | 新建部门（管理员） |
| PUT | /api/departments/:id | 编辑部门（管理员） |
| DELETE | /api/departments/:id | 删除部门（管理员，有成员时拒绝） |
| GET | /api/users | 用户列表 |
| GET | /api/users/:id | 用户详情 |
| POST | /api/users | 新建用户（管理员） |
| PUT | /api/users/:id | 编辑用户（管理员） |
| DELETE | /api/users/:id | 删除用户（管理员，含数据快照） |
| GET | /api/projects | 项目列表 |
| POST | /api/projects | 新建项目 |
| GET | /api/projects/:id | 项目详情 |
| PUT | /api/projects/:id | 编辑项目 |
| DELETE | /api/projects/:id | 软删除项目 |
| GET | /api/projects/:id/tasks | 任务列表 |
| POST | /api/projects/:id/tasks | 新建任务 |
| GET | /api/projects/:id/docs | 项目资料列表（需求文档 / 会议纪要 / 设计稿 / 项目密钥 / 测试账密等） |
| POST | /api/projects/:id/docs | 新建项目资料（项目成员） |
| PUT | /api/docs/:id | 编辑项目资料（项目成员中的资料创建人 / 项目创建人 / 管理员） |
| DELETE | /api/docs/:id | 删除项目资料（软删除，权限同上，可在回收箱恢复） |
| POST | /api/uploads | 附件上传（multipart/form-data，字段名 `files`，最多 10 个） |
| GET | /api/tasks/:id | 任务详情 |
| PUT / PATCH | /api/tasks/:id | 编辑任务 |
| DELETE | /api/tasks/:id | 软删除任务 |
| PATCH | /api/tasks/:id/move | 看板拖拽落位（`{status, index}`），服务端重排 position |
| GET | /api/trash | 回收箱：已删项目 + 已删任务 + 已删项目资料（附同批数量 / 所属项目标题） |
| POST | /api/trash/projects/:id/restore | 恢复项目（连同同批删除的任务、资料与附件） |
| DELETE | /api/trash/projects/:id | 彻底删除项目（物理删除项目及其任务、资料、附件、成员关联） |
| POST | /api/trash/tasks/:id/restore | 恢复任务（连同同批删除的附件） |
| DELETE | /api/trash/tasks/:id | 彻底删除任务（物理删除任务及附件） |
| POST | /api/trash/docs/:id/restore | 恢复项目资料（连同同批删除的附件） |
| DELETE | /api/trash/docs/:id | 彻底删除项目资料（物理删除资料及附件） |
| GET | /api/logs | 操作日志分页（管理员）：`targetType` / `action` / `operatorId` / `keyword` / `page` / `pageSize`（默认 20，上限 100） |
| GET | /api/logs/:id | 操作日志详情（管理员），物理删除的条目含 `snapshot` 数据快照 |
| GET | /healthz | 健康检查（含 Redis 状态） |
| GET | /setup | 配置完成后的反馈页 |

## 回收箱（软删除）

删除项目 / 任务 / 项目资料只写 `deleted_at`，数据留在库中，默认查询自动过滤（GORM 软删除）。
同一批被删的记录共用同一个 `deleted_at` 时间戳，恢复时按该时间戳精确还原：

- 恢复项目会带回随它一起删掉的任务、项目资料与附件；在项目删除**之前**单独删掉的任务 / 资料
  不会被误恢复，它们仍留在回收箱的任务区 / 资料区，可以单独恢复。
- 项目删除时任务与资料跟着删，它们只在回收箱的**项目条目**里体现（`taskCount` / `docCount`
  就是恢复时会回来的任务数与资料数），不重复出现在任务区 / 资料区。
- 资料与任务一样支持单独恢复 / 彻底删除；随项目一起被删的资料与项目同批次，恢复项目时一并还原。
- 「彻底删除」是物理删除（`Unscoped`），项目会连带其下全部任务、项目资料、附件与成员关联一起清掉，不可恢复，
  因此会在操作日志里留下整棵数据的快照（见下节）。
- 可见范围与处置权限对称：管理员看全部，其他用户只看自己有权处置的数据（项目按创建人；
  任务按创建人 / 被指派人 / 项目创建人；资料按资料创建人 / 项目创建人，且必须仍是项目成员）。

## 操作日志

覆盖**项目 / 任务 / 项目资料 / 用户 / 部门**五类对象的操作留痕，写库表为 `operation_logs`，
在「操作日志」页面（仅管理员可见）按类型 / 动作 / 关键词检索。

- 记录内容：操作人（id + 当时的用户名快照）、对象类型与名称、动作、变更摘要、时间。
  摘要由服务端按字段级差异生成，例如 `状态：待办 → 进行中；负责人：张三 → 李四`；
  富文本正文只在变更时提示「描述已更新」，不把正文写进日志。
- 动作取值：`create` 新建、`update` 修改、`move` 看板拖拽改状态（仅任务）、`delete` 删除、
  `restore` 回收箱恢复、`purge` 回收箱彻底删除。
- **物理删除留数据快照**：`delete` 用户、`delete` 部门、`purge` 项目 / 任务 / 项目资料这几类操作会把被删数据
  整份序列化进 `snapshot`（JSON 文本列），因为删掉之后无从追溯：
  - 用户快照：账号信息（不含密码）+ 所属部门名 + 被指派 / 创建任务数、拥有项目数、项目成员关联数；
  - 部门快照：部门信息 + 删除时剩余人数（有成员时不允许删除，正常为 0）；
  - 项目快照：项目 + 成员 + 其下全部任务与项目资料及其附件（彻底删除项目会连带清掉这些数据）；
  - 任务快照：任务 + 全部附件 + 所属项目标题；
  - 资料快照：项目资料 + 全部附件 + 所属项目标题。
  软删除（`delete` 项目 / 任务 / 项目资料）的数据仍在库中，只记动作、不留快照。
- 写入时机：普通操作在业务成功后异步补记（失败只打服务端日志，不影响接口结果）；
  物理删除与它的快照写在**同一个事务**里，避免出现「数据删了、快照却没留下」。
- 日志只存操作人 id 与用户名快照、不建外键关联，因此删除用户不会连带删掉它的历史日志。
- 列表接口不下发 `snapshot`（单条可能很大，彻底删除项目会带上全部任务），
  点「查看」调详情接口取单条。日志接口不挂缓存中间件，保证审计实时性。

## 数据库迁移

服务启动时自动执行，无需手动操作：

- 建表与补列：按 `Department → User → Project → Task → ProjectMember → TaskAttachment → ProjectDoc → ProjectDocAttachment → UploadedFile → OperationLog`
  顺序建表；已存在的表按模型字段补缺失列（规避 PostgreSQL 18 + pgx 组合下 AutoMigrate 的报错）
- 补索引：项目 / 任务 / 附件 / 项目资料及其附件的 `deleted_at` 软删除索引按需创建
- 数据补齐：老任务缺失的 `creator_id` 回退为所属项目创建人（幂等，只补空值）
- 首次出现 `project_members` 表时，把已有项目初始化为「全员可见」，避免历史项目突然对所有人不可见

## 前端页面与权限

| 路由 | 页面 | 可见范围 |
| --- | --- | --- |
| /login | 登录 | 公开 |
| / | 首页（按登录态跳转） | — |
| /projects | 项目列表 | 登录用户 |
| /projects/[id] | 四列看板（待办 / 进行中 / 已完成 / 已验收）+ 顶部「项目资料」入口（需求文档 / 会议纪要 / 设计稿 / 项目密钥 / 测试账密，支持富文本、附件与按标题检索），任务可按「全部 / 我创建的 / 指派给我的」筛选 | 登录用户 |
| /trash | 回收箱（项目 / 任务 / 项目资料，可恢复或彻底删除） | 登录用户（按权限过滤） |
| /users | 用户管理 | 管理员 |
| /departments | 部门管理 | 管理员 |
| /logs | 操作日志 | 管理员 |

前端支持浅色 / 深色主题切换，token 存 `localStorage`（键名 `mini-kanban_token`），401 时自动清除并跳回登录页。

## 多语言

界面支持**简体中文**与 **English** 两种语言，默认中文；语言切换按钮在导航栏右上角（登录页也有），
URL 不随语言变化，刷新 / 分享出去的链接对方看到的是自己 cookie 里的语言。

### 语言怎么存、怎么生效

- **cookie `mini-kanban-lang`**（值 `zh-CN` / `en`，一年，`SameSite=Lax`，HTTPS 下加 `Secure`）是唯一口径：
  根布局在服务端读它，用来定 `<html lang>` 与页面标题，正文文案也跟着首屏就渲染成对应语言，
  不会出现「先闪一下中文再切成英文」。`localStorage` 里存一份同名镜像，只在 cookie 被清掉时兜底。
- 切换语言时 cookie / localStorage / `<html lang>` / 请求头一起更新，无需刷新页面。
- 代价是路由从静态预渲染变为**按请求渲染**（用了 `cookies()`）。前端本来就依赖登录态取数，影响可忽略。
- **静态导出模式下这一条不成立**（`output: "export"` 没有请求上下文）：首屏只能按默认语言渲染，
  由 `<head>` 里的内联脚本先纠正 `<html lang>`，挂载后 `I18nProvider` 再按 cookie / localStorage
  切换文案 —— 因此非默认语言会闪一下。发布包用的就是静态导出（页面要随包分发）；
  如果更在意首屏不闪，就用仓库根的 Docker Compose 那套部署，服务端渲染版本不受这条限制。

### 前端文案

| 文件 | 作用 |
| --- | --- |
| `web/lib/i18n/zh-CN.ts` | **源语言**，同时是全部文案键的定义（`域.用途` 扁平点号，`{name}` 是插值占位符） |
| `web/lib/i18n/en.ts` | 英文，类型是 `Record<MessageKey, string>` —— 中文新增 / 改名的键这里没跟上，`npm run build` 直接报缺键 |
| `web/lib/i18n/index.ts` | 语言常量、字典查找与插值、cookie / localStorage 读写、`tNow()` |
| `web/components/I18nProvider.tsx` | 全局语言 Provider：`useI18n()` 取 `{ lang, locale, setLang, t }` |
| `web/lib/labels.ts` | 业务枚举 → 文案（看板列名、角色、操作日志的对象类型 / 动作），未知值原样回显 |

用法约定：

- **渲染时**用 `const { t } = useI18n()`，带参数时 `t("key", { name })`；
- **非渲染路径**（`lib/api.ts` 的兜底文案、异步回调里的失败提示）用 `tNow(...)` ——
  它是模块级的当前语言，不参与渲染，因此不会因为切语言触发无谓的重新取数；
- 日期时间统一走 `lib/date.ts` 的 `formatDateTime(iso, locale)`，`locale` 取 `useI18n().locale`，
  中文输出 `2026/09/20 14:30`、英文输出 `09/20/2026, 14:30`；
- 涉及数量的英文文案刻意写成 `Tasks: {count}` 这类，避开单复数变化，因此没有引入 ICU 复数规则。

新增一种语言：加一份字典 → 在 `lib/i18n/index.ts` 的 `LANGS` / `DICTS` / `LANG_LABELS` /
`LANG_LOCALES` 里登记 → 切换控件自动多出一个选项。

### 后端接口文案

后端的 `error` / `message` 也按语言返回，协商顺序（`internal/i18n`）：

1. **`Accept-Language` 头** —— 前端 `lib/api.ts` 的每个请求都会带上当前界面语言；
2. **`mini-kanban-lang` cookie** —— 附件下载（`/api/files/xxx?ticket=`）由浏览器直接发起、带不了自定义头，靠它兜底；
3. 都没有 → 中文。

英文目录在 `server/internal/i18n/messages.go`，**以中文原文为键**，因此 `fail` / `badRequest` /
`okMessage` 这些出口内部统一查表即可，调用方零改动；带动态内容的消息（`保存失败: <详情>`、
`一次最多上传 10 个文件`、`附件「a.png」不存在`）用「前缀 + 尾缀 + `「」`」三类规则拼装。
未收录的原文（例如数据库返回的英文错误）原样透传。

护栏：`go test ./internal/i18n/` 会扫描 `internal/handlers` 里写死的中文文案，
逐条确认能在英文目录里找到译文 —— 新增 / 修改文案忘记翻译会在测试里直接失败。

**已知边界**（刻意不做）：

- 操作日志的 `summary` / 变更项与数据快照是**写入时落库的文本**，保持中文：
  否则同一条日志会随查看者语言变化，审计语义就不稳了。日志页其余文案（表头、筛选、分页）都是英文的。
- 初始化配置向导 `/setup` 是一次性运维页面，保持中文。
- `Provider` 里 `useXxx 必须在 XxxProvider 内使用` 这类断言只会在控制台出现，不参与翻译。


## Redis 的用途

1. **缓存中间件**：`middleware.Cache(ttl, namespace)` 缓存 GET 响应（按用户 + 路径 + 查询串生成 key），
   命中时返回 `X-Cache: HIT`；所有写操作通过 `cache.DeletePrefix("httpcache:<namespace>:")` 主动失效。
2. **登出黑名单**：`auth:blacklist:<jti>` 保存至 token 自然过期，鉴权中间件会拒绝已登出 token。
3. Redis 不可用时自动降级为无缓存模式，服务仍可运行（`GET /healthz` 返回 `redis` 状态）。

## 自检与测试

改动后端或前端后，建议先跑这几项再部署：

```bash
# 后端：静态检查 + 单元测试
cd server
go vet ./...
go test ./...
```

其中 `internal/i18n` 的测试是一道**文案护栏**：它扫描 `internal/handlers` 里写死的中文文案，
逐条确认英文目录（`internal/i18n/messages.go`）里有对应译文 —— 新增接口文案忘了翻译会在这里失败。

```bash
# 前端：类型检查（不产出文件）
cd web
npx tsc --noEmit
```

前端的多语言完整性也由类型系统兜底：`lib/i18n/zh-CN.ts` 导出 `MessageKey` 联合类型，
`en.ts` 声明为 `Record<MessageKey, string>`，所以**英文缺键、组件里写错文案键**都会在 `tsc --noEmit`（以及 `npm run build`）阶段直接报错。
同理，`lib/api.ts` 的请求参数与返回类型写错也会被这一步拦下。

服务起来后的快速冒烟：

```bash
curl http://localhost:8281/healthz    # {"redis":true,"status":"ok"}
```

## 部署

发布的**唯一产物是「开箱即用」包**：后端可执行文件 + `web-dist/` 页面目录，
由同一个进程、同一个端口提供接口与页面。目标机不需要 Node、不需要额外的静态目录，也不需要 nginx 与 CORS。

> 为什么不把前端 `go:embed` 进二进制：那样每个平台的可执行文件会凭空多出约 1MB，
> 六平台交叉编译时这部分要被反复写进每一个产物，而且改一次前端就要把 Go 侧全量重编一遍。
> 改成随包分发后，可执行文件里只剩业务代码，页面文件还能被网关 / CDN 直接读盘缓存。
> 代价是 `web-dist/` 必须与可执行文件同目录（可以用 `WEB_DIST_DIR` 指向别处）；
> 缺了它服务仍能启动，但只提供接口、非 `/api` 路径返回 404。

> 源码层面仍支持把前端单独跑起来（`npm run dev` 联调、或仓库根的 Docker Compose），
> 但**发布只发这一个包**，避免用户挑错。Docker Compose 那套是给容器化部署用的另一条路，与发布包无关。

### 自己打一个发布包

```bash
bash build-web-static.sh          # 1) 前端静态导出到 web/out，再拷进 server/web-dist
bash build.sh -v v1.0.0           # 2) 交叉编译 + 打包（把 web-dist/ 拷进每个归档）
```

Windows 上等价于：

```bat
build-web-static.bat                 :: 静态导出并拷进 server\web-dist
bash build.sh -v v1.0.0              :: Git Bash 里跑；也可以 cd server && go build -o mini-kanban.exe .
```

产物落在 `dist/`，每个平台一个归档，顶层目录名与归档名一致：

```
dist/
├── mini-kanban-v1.0.0-windows-amd64.zip
├── mini-kanban-v1.0.0-windows-arm64.zip
├── mini-kanban-v1.0.0-linux-amd64.tar.gz
├── mini-kanban-v1.0.0-linux-arm64.tar.gz
├── mini-kanban-v1.0.0-darwin-amd64.tar.gz
├── mini-kanban-v1.0.0-darwin-arm64.tar.gz
└── checksums.txt                  # 全部归档的 SHA256（sha256sum -c 可校验）
```

每个归档解开是：

```
mini-kanban-v1.0.0-linux-amd64/
├── mini-kanban(.exe)     可执行文件（只有后端代码）
├── web-dist/              前端静态产物，必须与可执行文件同目录
├── start.sh / start.bat  启动脚本（打印版本号，异常退出时列出常见原因）
├── version.txt           本包版本号
├── config.example.yaml   配置模板（复制为 config.yaml 后按环境修改）
├── README.md / README.en.md
└── LICENSE
```

- **`-v` 省略时自动取最近的 git 标签**，所以在打好标签的工作区里直接 `bash build.sh` 也不会漏版本号；
  版本号里的非法字符会被拒绝（避免拼出奇怪的文件名）。
- **前端静态产物是硬前提**：`server/web-dist/index.html` 不存在时 `build.sh` 直接报错退出，
  不会再产出「只有接口没有页面」的包。
- **改前端不用重编 Go**：`web-dist` 不参与编译，`next build` 完直接 `bash build.sh` 即可，
  Go 侧命中编译缓存；这比 `go:embed` 方案快得多（embed 的内容一变，整个包就要重编）。
- Git Bash 默认不带 `zip` 命令，`build.sh` 会自动退回用 Python 的 `zipfile` 打 Windows 归档，产物结构一致。
- 推送 `v*` 标签即触发 GitHub Actions：跑静态检查与单元测试 → 构建前端 → 交叉编译 → 校验产物
  （校验和、数量、文件名是否带标签、包内是否带 `web-dist/index.html` 与 `README.en.md`）
  → 解包真跑一遍冒烟（页面、深链、缓存头、接口 404）→ 创建 / 更新 Release。

### 跑起来是什么样

```
浏览器 → mini-kanban :8281
          ├── /api/*    → 接口
          ├── /healthz  → 健康检查
          ├── /setup    → 初始化向导
          └── 其他      → web-dist 里的静态页面（未命中的前端路由回退 index.html）
```

- 实现见 `server/webstatic.go`：启动时 `ResolveWebDist()` 解析一次目录（默认工作目录下的 `web-dist/`，
  可用 `WEB_DIST_DIR` 覆盖），结果挂在 `r.NoRoute` 上，因此 `/api`、`/healthz`、`/setup` 仍由原路由优先匹配；
  `/api/` 前缀未命中时返回 JSON 404，不会把页面当接口响应返回。
- 目录缺失时自动降级：启动日志写明 `前端: 未找到（仅提供接口）` 并打印它找的那个绝对路径，
  非 `/api` 路径返回 404，接口行为完全不变。**发布包不会出现这种状态**，`build.sh` 已经卡住了这一步。
- 缓存策略：`/_next/static/*` 文件名带内容哈希，发 `max-age=31536000, immutable`；
  页面与其他产物发 `no-cache`，避免换版本后浏览器拿着旧 `index.html` 去请求已删除的 chunk。
- 看板深链 `/projects/<id>` 是唯一带动态段的路由：静态导出要求动态段至少预渲染一条路径才肯构建
  （Next 用「预渲染路径数 > 0」判断页面有没有 `generateStaticParams`），所以 `web/app/projects/[id]/page.tsx`
  预渲染了一份占位页 `/projects/_/`；`server/webstatic.go` 把真实的 `/projects/<id>` 映射到这份占位页，
  项目 id 由前端从地址栏解析（静态产物里 `useParams()` 拿到的只是构建期占位值，
  见 `ProjectBoard.tsx` 的 `resolveProjectId`）。
- 因为页面与接口同源，**不再需要 nginx 与 CORS**；要加 TLS / gzip 再在前面挂一层网关即可（见「前面挂网关」）。

### 方式一：Docker Compose 一键部署（容器化场景）

想用容器跑（自带 PostgreSQL / Redis，或想把前端单独扩多实例），走仓库根目录的 `docker-compose.yml`：
一次拉起 5 个服务 —— `nginx`（统一入口）、`web`（Next.js）、`server`（Go）、`postgres`、`redis`。

```bash
# macOS / Linux
cd <仓库根目录>
cp .env.example.docker-compose .env       # 改掉里面的数据库 / Redis 密码与 JWT_SECRET
docker compose up -d --build
# 浏览器打开 http://<服务器IP>:8080（宿主机端口由 .env 的 HTTP_PORT 决定）
```

```powershell
# Windows PowerShell
cd <仓库根目录>
Copy-Item .env.example.docker-compose .env  # 改掉里面的数据库 / Redis 密码与 JWT_SECRET
docker compose up -d --build
# 浏览器打开 http://<服务器IP>:8080（宿主机端口由 .env 的 HTTP_PORT 决定）
```

Linux 服务器上也可以直接跑一键脚本，它会把上面几步连起来做（生成 `.env`、随机化数据库 / Redis 密码与 JWT 密钥、**生成带随机密码的初始管理员**、挑一个未被占用的宿主机端口、构建并启动），并且**重复执行安全**：

```bash
cd <仓库根目录>

# 下面两条二选一，首次部署前执行其中一条即可（管理员密码均由脚本随机生成）
bash deploy/deploy.sh                                                # 默认管理员：admin@mail.example.com / Admin
bash deploy/deploy.sh --admin-email me@corp.com --admin-user kevin   # 自定义管理员的邮箱与显示名
```

管理员账号相关说明：

- 上面两条命令是**二选一**，不是先后执行：脚本只负责把邮箱 / 显示名写进 `.env`，而账号本身是后端首次启动（用户表为空）时才创建的。
- 密码由脚本随机生成（16 位字母数字，排除了 `0 O 1 l I` 这类易混淆字符），同时也写进 `.env` 的 `SEED_ADMIN_*`，忘了随时可查。**只有全新数据库（数据卷不存在）时**才会把密码打印到终端；重复部署时因为账号已经建好，只提示一句"账号已存在"，不再输出密码。
- 该账号**只在用户表为空时创建一次**：重复执行脚本、后续重启或重建容器都不会重建它，也不会改掉已生效的密码；后端创建成功后还会把容器内 `config.yaml` 的 seed 段自动清空。
- 想换邮箱或显示名，要在**首次启动之前**用 `--admin-email` / `--admin-user` 指定；账号已经创建后再改，只能在「用户管理」页面里改。
- 不用脚本、走手工部署时，等价做法仍然是把 `.env` 的 `SEED_ENABLED` 改为 `true` 并填好 `SEED_ADMIN_*` 三项。
- **忘了管理员密码**：不必清库，执行 `bash deploy/deploy.sh --reset-admin` 即可。它会把 `SEED_ADMIN_EMAIL` 指定的账号重置为管理员并换一个随机密码（该邮箱已不存在时重新创建），其他数据不受影响，新密码会打印在部署结果里。对应的后端配置是 `seed.forceReset`（环境变量 `SEED_FORCE_RESET`），只在本次启动生效——脚本启动后会自动把它改回 `false`，后端处理完也会清空容器内 `config.yaml` 的 seed 段。

要点：

- **宿主机端口冲突**：`HTTP_PORT`（默认 8080）是整份 compose 里**唯一**会映射到宿主机、因而可能与既有服务冲突的端口；`SERVER_PORT` / `DB_PORT` / `REDIS_PORT` / 前端 3201 都是容器内端口，与宿主机上跑着什么无关。冲突时改 `.env` 的 `HTTP_PORT`（只重建 nginx：`docker compose up -d nginx`），或用 `bash deploy/pick-port.sh` 自动挑一个空闲端口写回 `.env`（`deploy/deploy.sh` 已内置这一步）。
- **后端配置由环境变量生成**：容器入口脚本 `server/docker-entrypoint.sh` 把 `SERVER_*` / `DB_*` / `REDIS_*` / `JWT_*` / `SEED_*` / `STORAGE_*` 渲染成 `/app/config.yaml`。这么做是因为后端只认配置文件（`config.seed` 段没有环境变量入口）。想完全自定义就挂载自己的 `config.yaml` 到 `/app/config.yaml`，脚本检测到文件已存在会跳过生成。
- **前端接口地址是构建期内联的**：`NEXT_PUBLIC_API_BASE` 经 compose 的 build args 传入，默认 `/api`（同源反代，天然规避 CORS）。改了地址必须重建前端：`docker compose up -d --build web`。
- **`web/.dockerignore` 排除了 `.env*`**：本地 `.env.local`（`http://localhost:8281/api`）的优先级高于构建参数，一旦随构建上下文进镜像就会把 localhost 写进产物。
- 数据都在命名卷里：`pgdata`（数据库）、`redisdata`（AOF）、`uploads`（附件）。`docker compose down` 不删卷，需要清空时加 `-v`。
- **数据库密码只在数据卷首次初始化时生效**：`POSTGRES_PASSWORD` 是 PostgreSQL 在数据卷为空时写入的，之后改 `.env` **不会**同步到库里，表现是服务反复重启并报 `password authentication failed for user "mini-kanban" (SQLSTATE 28P01)`。库里没有需要保留的数据时，最省事的修法是：
  ```bash
  bash deploy/deploy.sh --reset      # 先 docker compose down -v，再重新部署（数据库 / Redis / 附件都会删除）
  ```
  `deploy/deploy.sh` 检测到「数据卷已存在 + 密码仍是模板占位值」时会拦下并提示这一点；若是你自己改过密码导致对不上，则先 `docker compose down -v` 清卷，或把 `.env` 改回初始化时用的值。要在**不删数据**的前提下改密码：`docker compose exec postgres psql -U postgres -c "ALTER USER \"mini-kanban\" WITH PASSWORD '新密码';"`（容器内本地连接是 trust，不需要密码）。
- 常用命令：`docker compose ps` 看健康状态、`docker compose logs -f server` 看后端日志、`docker compose up -d --build <服务名>` 只重建单个服务。
- 附件上限有两处要同步：`.env` 的 `STORAGE_MAX_SIZE_MB`（请求体上限 = 该值 × 10）与 `deploy/nginx.conf` 的 `client_max_body_size`。
- **镜像拉取慢或失败**：国内直连 Docker Hub 经常超时，报 `TLS handshake timeout` 或 `EOF`，这与本项目的配置无关，重试或配个加速器即可。Docker Desktop 在 Settings → Docker Engine 里加 `"registry-mirrors": ["https://<加速地址>"]` 后 Apply & Restart；Linux 服务器改 `/etc/docker/daemon.json` 后 `systemctl restart docker`。也可以先单独把基础镜像拉好再启动：`postgres:16-alpine`、`redis:7-alpine`、`nginx:1.27-alpine`、`golang:1.22-alpine`、`node:20-alpine`、`alpine:3.20`。

### 方式二：手工部署前端（仅在需要把前端单独跑时）

发布包已经把前端页面目录（`web-dist/`）一起带着，正常部署用不到这一节。只有下面这些情况才需要单独跑前端：
本地 `npm run dev` 联调、或用 Docker Compose 那种多服务部署。

```bash
# macOS / Linux
cd web
# NEXT_PUBLIC_* 是构建期内联的：必须先写好 .env.production 再 build，构建后改环境变量无效
echo "NEXT_PUBLIC_API_BASE=/api" > .env.production
npm ci
npm run build
pm2 start npm --name mini-kanban-web -- run start    # 或 systemd
```

```powershell
# Windows PowerShell
cd web
# 注意别用 > 重定向：PowerShell 会写成 UTF-16，Next 读出来是乱码。
# 这里内容本身是纯 ASCII，用 ascii 编码不会带 BOM。
Set-Content -Path .env.production -Value "NEXT_PUBLIC_API_BASE=/api" -Encoding ascii
npm ci
npm run build
# pm2 在 Windows 上启动 npm 脚本不稳定，建议改用 NSSM：
#   nssm install mini-kanban-web "C:\Program Files\nodejs\node.exe" "node_modules\next\dist\bin\next start -p 3201"
```

- `NEXT_PUBLIC_API_BASE=/api` 配合同域反代，可顺带规避 CORS
- 生产运行需要 `.next/` + `node_modules/`（生产依赖）+ `package.json`
- 建议在目标机器上执行 `npm ci && npm run build`：`.next` 含构建期产物，跨 Node 版本 / 架构搬运可能不兼容
- 想改成静态产物（交给任意静态服务器，或由后端从目录读取）：`bash build-web-static.sh`

### 方式三：手工部署后端

```bash
# 页面同源提供（发布包形态）：先出前端静态产物到 server/web-dist，再编译；
# 运行时把它与二进制放在一起（或用 WEB_DIST_DIR 指定）
bash build-web-static.sh
cd server && go build -o mini-kanban .

# 只提供接口（页面由前端服务负责，或本地联调）
cd server && go build -o mini-kanban .
```

```powershell
cd server
# 若在 Windows 上交叉编译出 Linux 二进制
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o mini-kanban .
```

- 一并部署 `config.yaml` 与附件目录（`storage.dir`，默认 `data/uploads`）
- `storage.dir` 的相对路径按**后端工作目录**解析，注意启动目录
- 生产环境不要误开 `SETUP_FORCE`

### 前面挂网关（可选）

一体化部署本身就是完整入口，不需要 nginx。要加 TLS、gzip 或多域名，再在前面挂一层网关即可 —— 
此时只有一个上游端口：

```nginx
server {
    listen 80;
    server_name mini-kanban.example.com;

    # 页面与接口同一个二进制、同一个端口，全部转发过去即可
    location / {
        proxy_pass http://127.0.0.1:8281;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        client_max_body_size 100m;             # 需不小于 storage.maxSizeMB × 10（默认 10 → 100MB）
    }
}
```

### 上线前检查

- `jwt.secret` 换成随机长字符串，`server.mode` 设为 `release`
- `config.yaml` 含数据库明文密码，文件权限收紧；它已被 `.gitignore` 忽略，不要入库
- 后端 CORS 中间件会回显请求的 `Origin`（便于本地联调），同域反代部署时不受影响；如前后端分域部署且需收紧来源，需自行调整 `internal/middleware/cors.go`
- `/setup` 向导不应暴露到公网（配置就绪后访问只会显示反馈页，但仍建议在网关层限制）
- 附件下载是「票据 + 随机文件名」两道防线：URL 必须带上登录后换取的下载票据（`?ticket=`，见「附件存储」），
  落盘名是 32 位随机 hex；不要把文件名改成可预测命名，也不要去掉 `middleware.FileTicketRequired()`

---

[MIT](LICENSE) © 2026 kite88
