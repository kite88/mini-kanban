# 更新日志

本项目遵循 [语义化版本](https://semver.org/lang/zh-CN/)，版本号形如 `v主版本.次版本.修订号`。
发版说明放在 `docs/release-notes/<tag>.md`（推送 `v*` 标签时由 GitHub Actions 自动创建 Release）。

## [v1.0.1] - 2026-10-16

品牌与标题统一：产品名改为 **Mini-Kanban**，浏览器标签页标题统一成「页面名 - 站点名」。
功能与接口不变，纯界面文案调整。

### 变更

- **站点名统一为 Mini-Kanban**：`app.title` 由「任务看板」/ `Task Board` 改为 `Mini-Kanban`。导航栏左上角品牌、
  标签页标题、登录页标题都取同一个键，中英文一致（产品名是专有名词，不参与翻译）。
- **逐页标签页标题**：新增「`页面名 - {app.title}`」口径（如 `登录 - Mini-Kanban`、`项目 - Mini-Kanban`）。
  页面本身都是 `"use client"`、不能导出 `metadata`，因此页面名在各自路由段的 `layout.tsx` 里用 `pageMetadata()`
  声明；路由 → 文案键的映射与拼装集中在新增的 `web/lib/i18n/pageTitle.ts`（服务端与客户端共用同一张表），
  客户端跳转 / 切换语言时由新增的 `web/components/DocumentTitle.tsx` 同步刷新（静态导出产物在构建期
  只能写死默认语言）。首页与 404 这类没有页面名的路由只显示 `Mini-Kanban`。
- **登录页文案**：卡片标题由「登录任务看板」改为 `Mini-Kanban - 任务看板系统`（新增站点级键 `app.tagline`，
  英文 `Task board system`；原 `login.title` 键删除）。
- **初始化向导页面标题**：`/setup` 与状态页的 `<title>` 同步为「初始化配置 - Mini-Kanban」「运行状态 - Mini-Kanban」，
  页内主标题统一成 `Mini-Kanban`（原来是小写的 `mini-kanban`）。
- **文案键**：新增 `app.tagline` 与 `page.*`（7 个页面名，中英双语），删除 `login.title`。
- **小重构**：根布局里的 `currentLang()`（服务端读语言 cookie）抽到 `web/lib/i18n/server.ts`，
  与各路由段 layout 的 `pageMetadata()` 共用，避免每个 layout 重复一遍语言判定。

### 文档

- README（中英）补上标签页标题口径与 `lib/i18n/pageTitle.ts` / `lib/i18n/server.ts` /
  `components/DocumentTitle.tsx` 三处说明，产品标题挂上 `Mini-Kanban`。
- 界面截图按新品牌重拍（登录页标题与导航栏品牌文字都变了）。

## [v1.0.0] - 2026-10-15

首个正式版本：一个可自托管的拖拽式任务看板，**下载一个包、解压即用**。

### 新增

- **开箱即用（后端直接托管页面目录）**：`server/webstatic.go` 在启动时解析一次 `web-dist/` 目录
  （默认工作目录下的 `web-dist/`，可用 `WEB_DIST_DIR` 指向别处）并挂在 `r.NoRoute` 上，
  `/api`、`/healthz`、`/setup` 仍由原路由优先匹配。同一个进程、同一个端口同时提供接口与页面，
  目标机不需要 Node、不需要额外的静态目录，也不需要 nginx 与 CORS。
  - **刻意不用 `go:embed`**：内嵌会让每个平台的可执行文件多出约 1MB，六平台交叉编译时这部分要被
    反复写进每一个产物，而且改一次前端就要把 Go 侧全量重编；改成随包分发后，改前端只需
    `next build` + `bash build.sh`，Go 侧命中编译缓存。
  - 目录缺失时自动降级为「只提供接口」：启动日志打印它找的那个绝对路径，非 `/api` 路径返回 404，
    接口行为完全不变。
  - 缓存策略：`/_next/static/*` 文件名带内容哈希，发 `max-age=31536000, immutable`；页面等产物发
    `no-cache`，避免换版本后浏览器拿旧 `index.html` 去请求已删除的 chunk。
  - `build-web-static.sh`（静态导出到 `web/out` 再拷进 `server/web-dist`）与 Windows 版 `build-web-static.bat`
    （GBK 编码，可双击；`--skip-install` 可复用现有 `node_modules`）。
  - `build.sh` 把前端静态产物当成**硬前提**：`server/web-dist/index.html` 不存在就直接报错退出，
    不会再产出「只有接口没有页面」的包。
- **只发布一种产物**：不再提供前后端分离的 standalone 前端包（`build-web.sh` 与 `tools/release-web/` 已移除），
  避免用户挑错包；Release 下每个平台一个归档 + `checksums.txt`，归档内是
  `可执行文件 + web-dist/ + 启动脚本 + 配置模板 + README（中英）+ LICENSE + version.txt`。
- **产物名带版本号**：`mini-kanban-v1.0.0-<os>-<arch>.zip`（Windows）/ `.tar.gz`（其余），顶层目录同名，
  包内附 `version.txt`，`start.sh` / `start.bat` 启动时打印版本号 —— 发布版本多了也不会拿错包。
  - `build.sh` 新增 `-v` 选项；不传时自动取最近的 git 标签。
  - CI 在发 Release 前会校验归档数量、文件名是否带标签、以及归档里确实含内嵌的 `index.html`。
- **登录与账号**：JWT + bcrypt，登出后 token 进 Redis 黑名单；管理员在页面内增删改用户与部门（有成员 / 有任务关联时阻止删除）。
- **项目与看板**：四列看板（待办 / 进行中 / 已完成 / 已验收），`@dnd-kit` 支持同列排序与跨列拖拽，服务端按 `position` 重排，任务可按「全部 / 我创建的 / 指派给我的」筛选。
- **任务**：富文本描述、负责人、截止时间、附件列表，卡片直接展示负责人、附件数与是否逾期。
- **项目资料**：挂在项目下的文档型条目（需求文档 / 会议纪要 / 设计稿 / 项目密钥 / 测试账密），支持富文本、内嵌图片、附件与按标题检索。
- **附件**：扩展名白名单 + 32 位随机文件名落盘；下载走独立的短时效「下载票据」，登录 token 不出现在任何 URL 里，且下载还要按数据归属校验可见性（即看即下）。
- **回收箱**：项目 / 任务 / 资料软删除后按「批次时间戳」集中管理，恢复时精确还原同批数据；彻底删除会在操作日志里留下数据快照。
- **操作日志**：覆盖项目 / 任务 / 资料 / 用户 / 部门五类对象的增删改与拖拽留痕，支持按类型 / 动作 / 关键词检索。
- **多语言**：界面简体中文 / English；后端接口文案按 `Accept-Language` 返回，并有测试扫描遗漏的译文。
- **主题**：浅色 / 深色 / 跟随系统，写 `localStorage`。
- **初始化向导**：首次启动无 `config.yaml` 时进入 `/setup`，含数据库与 Redis 连通性测试与写回；保存后弹窗显示
  「配置已写入 → 等待服务就绪 → 3.5 秒后自动打开状态页」的倒计时，点遮罩可关闭并取消自动跳转。
- **Docker Compose**：仓库根 `docker-compose.yml` 一次拉起 nginx / web / server / postgres / redis 五个服务，附一键部署脚本。

### 修复

- **静态托管的目录型路由 301 死循环**：原实现把交给 `http.FileServer` 的路径填成 `.../index.html`，
  而 `serveFile` 会把它 301 到 `./`，浏览器解析 `./` 又落回同一地址 —— `/`、`/projects/`、`/login/`
  等全部无限重定向、浏览器直接打不开。现在统一用「目录形式」路径，实测各页均 200 且与构建产物字节一致。
- **看板深链在静态产物下丢 id**：Next 以「预渲染路径数 > 0」判断有没有 `generateStaticParams`，
  返回空数组会直接构建失败。改为预渲染占位页 `/projects/_/`，静态服务器把真实的 `/projects/<id>`
  及其 RSC 载荷映射到它，项目 id 由前端从地址栏解析（静态产物里 `useParams()` 只有构建期占位值）。
- **启动脚本不再预生成 `config.yaml`**：模板里的数据库密码是占位值，配置一存在程序就不会进入初始化向导，
  接着因连不上数据库退出，Windows 下表现为双击 `start.bat` 窗口一闪而关。现在脚本只在 `config.yaml` 存在时
  提示其路径，缺失时直接进入 `/setup` 向导。
- **异常退出时不再静默关闭**：`start.sh` / `start.bat` 在退出码非 0 时打印退出码与常见原因
  （数据库连不上 / Redis 连不上 / 8281 端口被占用 / 配置缺失或格式错误）并等待按键。
- **`start.bat` 中文提示乱码**：脚本是 UTF-8 编码，中文 Windows 控制台默认代码页为 936（GBK），
  不切换就会显示成乱码。脚本启动时先切到 65001（UTF-8）。

### 实现要点

- 主键选型：用户 / 部门用雪花 ID（可按节点号区分实例），项目 / 任务 / 资料用 UUID v7，均由应用侧在落库前生成且整体按时间递增；对外一律按字符串序列化，避免前端大整数精度丢失。
- 分层单向依赖：`router → handlers → models / db`，写操作统一经 `oplog` 留痕；物理删除与快照写入同一事务，不会出现「数据删了、快照没留下」。
- 回收箱用 `deleted_at` 时间戳作为批次标识：项目被删时连同其任务 / 资料一起恢复，而此前单独删除的任务仍留在回收箱各自分区。
- 附件下载是「票据 + 数据权限」两道校验：票据只证明"是谁在下载"（`aud=file`，与登录 token 用派生密钥互不相通），能否下载还要按附件归属反查项目 / 任务可见性；无权一律返回 404，不泄露文件是否存在。
- 富文本用 bluemonday 白名单净化，图片地址只放行本站 `/api/files/<随机名>`，票据只在渲染时拼接、保存时剥离。
- 缓存：GET 列表走 Redis 缓存（命中返回 `X-Cache: HIT`），写操作按命名空间前缀失效；Redis 不可用时降级为无缓存继续运行。
- 映射深链载荷 `projects/<id>/index.txt` 用 `strings.CutSuffix` 精确切 id 段，避免把静态路由自己的载荷 `projects/index.txt` 误判成深链载荷。
- 静态目录改从磁盘读之后，`server/web-dist` 不再是编译期输入：不需要 `//go:embed` 撑着目录，
  `server/web-dist/.gitkeep` 随之删除，Docker 构建上下文也把它排除掉。
- GitHub Actions 的 Release 说明必须用 YAML 折叠标量（`>-`）而不是 `run` 里的多行双引号字符串：flow scalar 中不能出现空行，否则整个 workflow 解析失败（表现为 run 0 秒失败）。