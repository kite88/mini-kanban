[简体中文](README.md) | **繁體中文** | [English](README.en.md)

# Mini-Kanban · 拖拽式任務看板（Next.js + @dnd-kit / Go Gin + GORM / PostgreSQL / Redis）

[![Release](https://img.shields.io/github/v/release/kite88/mini-kanban?v=2)](https://github.com/kite88/mini-kanban/releases/latest)
[![License](https://img.shields.io/github/license/kite88/mini-kanban?v=2)](LICENSE)

包含登入、使用者 CRUD、部門管理、專案 CRUD、任務 CRUD、四列拖拽看板、附件上傳、回收箱（軟刪除）與操作日誌（含物理刪除快照）。

## 介面預覽

**登入**（支援中英文與淺色 / 深色主題）

![登入頁](docs/screenshots/login.png)

**四列拖拽看板**（待辦 / 進行中 / 已完成 / 已驗收，任務卡帶負責人、附件數與截止時間，可按「全部 / 我建立的 / 指派給我的」篩選）

![四列拖拽看板](docs/screenshots/board.png)

**任務詳情**（富文字描述、附件列表、負責人 / 所在列 / 截止時間）

![任務詳情](docs/screenshots/task-modal.png)

**專案列表與專案資料**（資料支援富文字、內嵌圖片、附件與按標題檢索）

![專案列表](docs/screenshots/projects.png)

![專案資料](docs/screenshots/project-docs.png)

**回收箱**（軟刪除資料的恢復 / 徹底刪除，專案條目帶同批任務數）

![回收箱](docs/screenshots/trash.png)

**使用者管理與操作日誌**（管理員）

![使用者管理](docs/screenshots/users.png)

![操作日誌](docs/screenshots/logs.png)

**深色主題**

![深色主題看板](docs/screenshots/board-dark.png)

## 目錄結構

```
server/                                    Go Gin 後端（模組名 mini-kanban）
  main.go                                  啟動入口：載入配置 / 未配置時進入初始化嚮導
  config.example.yaml                      配置模板（入庫）
  config.yaml                              本地配置（.gitignore 忽略）
  internal/config                          配置載入、校驗、嚮導渲染寫回
  internal/setup                           首次啟動的配置嚮導（頁面 + 探測 + 寫回）
  internal/db                              GORM 連線、建表 / 補列 / 補索引、初始管理員
  internal/cache                           Redis 封裝（Get/Set/DeletePrefix）
  internal/i18n                            介面文案多語言：語言協商 + 英文目錄（以中文原文為鍵）
  internal/storage                         附件落盤、副檔名白名單、隨機檔名
  internal/models                          User / Department / Project / Task /
                                           ProjectMember / TaskAttachment / OperationLog
  internal/middleware                      Auth(JWT)、Admin、CORS、Redis GET 快取
  internal/handlers                        auth / user / project / task / project_doc /
                                           department / trash / file / log
  internal/oplog                           操作日誌寫入 + 物理刪除的資料快照
  internal/router                          路由與中介軟體集中註冊
  internal/utils                           JWT、郵箱與密碼校驗、檔名清理
  webstatic.go                             託管 web-dist 目錄裡的前端靜態站點（NoRoute + SPA 兜底）
  web-dist/                                前端靜態匯出產物目錄（構建生成，.gitignore 忽略）
  data/uploads/                            附件存放目錄（storage.dir，.gitignore 忽略）

build.sh                                   交叉編譯並打包可分發歸檔（帶 web-dist 目錄，產物名帶版本號）
build-web-static.sh                        前端靜態匯出並拷進 server/web-dist（build.sh 的前置步驟）
build-web-static.bat                       ↑ 的 Windows 版（GBK 編碼，可雙擊）

tools/release/                             歸檔裡的啟動指令碼（start.sh / start.bat）
docs/screenshots/                          README 用的介面截圖
docs/release-notes/                        各版本 Release 說明（推標籤時自動取用）

web/                                       Next.js 前端（App Router + TS + Tailwind）
  app/login                                登入
  app/                                     首頁（按登入態跳轉 /projects 或 /login）
  app/projects                             專案列表（增刪改）
  app/projects/[id]                        看板（四列拖拽 + 附件 + 富文字）
  app/departments                          部門管理（管理員）
  app/users                                使用者管理（管理員）
  app/trash                                回收箱（恢復 / 徹底刪除）
  app/logs                                 操作日誌（管理員，含資料快照檢視）
  components/Board.tsx                     @dnd-kit 跨列 / 同列排序
  components/                              AuthProvider / AuthGuard / NavBar / TaskCard /
                                           TaskModal / ProjectModal / ProjectDocsModal /
                                           MemberPicker / AttachmentList / RichTextEditor /
                                           Modal / ConfirmProvider / UserSelect /
                                           ThemeProvider / ThemeToggle /
                                           I18nProvider / LanguageToggle /
                                           DocumentTitle（標籤頁標題）
  lib/i18n/                                多語言：zh-CN.ts（源語言 + 全部鍵）/
                                           en.ts / index.ts（協商、插值、持久化）/
                                           pageTitle.ts（路由 → 頁面名）/ server.ts（服務端語言與頁面 metadata）
  lib/labels.ts                            業務列舉 → 文案（列名 / 角色 / 日誌型別與動作）
  lib/api.ts                               請求封裝、token(localStorage)、fileUrl
  lib/types.ts                             型別與看板列定義（COLUMNS）
  lib/html.ts · date.ts · validate.ts      富文字處理 / 日期 / 表單校驗
```

## 下載與部署（最快路徑）

釋出包在 [Releases](https://github.com/kite88/mini-kanban/releases/latest)，**每個版本只有一種產物**：
一個可執行檔案 + 一個 `web-dist/` 頁面目錄，由同一個程序、同一個埠提供介面與頁面，
目標機不需要 Node、不需要額外的靜態目錄，也不需要 nginx。

| 平臺 | 產物 |
| --- | --- |
| Windows x64 / ARM64 | `mini-kanban-v1.0.0-windows-amd64.zip` / `mini-kanban-v1.0.0-windows-arm64.zip` |
| Linux x64 / ARM64 | `mini-kanban-v1.0.0-linux-amd64.tar.gz` / `mini-kanban-v1.0.0-linux-arm64.tar.gz` |
| macOS Intel / Apple Silicon | `mini-kanban-v1.0.0-darwin-amd64.tar.gz` / `mini-kanban-v1.0.0-darwin-arm64.tar.gz` |

> **檔名裡的版本號就是釋出版本號**（例如 `v1.0.0`），頂層目錄與壓縮包同名，包內還有一份 `version.txt`；
> 啟動指令碼啟動時會列印版本號，所以下載的是哪一版一目瞭然。`checksums.txt` 是全部歸檔的 SHA256，
> 可用 `sha256sum -c checksums.txt` 校驗。

三步跑起來：

1. **完整解壓**對應平臺的歸檔（`web-dist/` 必須與可執行檔案在同一目錄）；
2. 執行 `start.sh`（Windows 雙擊 `start.bat`）；
3. 瀏覽器開啟 <http://localhost:8281>。目錄裡沒有 `config.yaml` 時會自動進入初始化嚮導
   （<http://localhost:8281/setup>），填好資料庫 / Redis 並儲存，配置寫入後彈窗顯示服務就緒倒計時，
   3.5 秒後自動開啟狀態頁。

前置條件只有一個：**PostgreSQL 資料庫需要提前建立**（服務只建表不建庫）；Redis 可選，留空即關閉快取。

想自己打包 / 改前端後再打包，見下面的「部署」章節；只想跑起來的話上面這三步就夠了。

## 環境要求

- Go 1.21+
- Node.js 18.17+（Next.js 14 要求）
- PostgreSQL（**資料庫需提前建立**，服務只建表不建庫）
- Redis（**可選**：`redis.host` 留空即關閉快取，服務照常執行）

## 快速開始

> **命令約定**：標 `bash` 的程式碼塊在 macOS / Linux 終端可直接執行，Windows 使用者在 PowerShell 裡同樣可用（`cd`、`go run .`、`npm run dev` 這類命令是通用的）；
> 存在平臺差異的操作（複製檔案、交叉編譯）會分別給出兩版。

### 方式一：初始化配置嚮導（推薦首次部署）

不提供 `config.yaml` 時，後端會先啟動一個只用於填寫配置的臨時服務：

```bash
cd server
go run .
```

瀏覽器開啟 `http://localhost:8281/setup`（埠可用 `SETUP_PORT` 指定；在伺服器上啟動時，日誌會列印可遠端訪問的 IP）。

嚮導提供：資料庫 / Redis 連線測試、配置校驗與寫回（按 `config.example.yaml` 模板渲染，**保留註釋與排版**）。資料庫連不上會攔下儲存，可勾選忽略；Redis 連不上只提示，快取降級執行。儲存成功後向導自動關閉，業務服務按正式配置啟動。

重新配置：用 `SETUP_FORCE=1` 啟動，原配置會備份為 `config.yaml.bak`。已配置完成後訪問 `/setup` 會看到「服務執行正常」的反饋頁。

> 注意：嚮導表單只包含 `server` / `database` / `redis` / `jwt` / `seed` 五段；`storage` 與 `database.logLevel` 刻意不進表單，嚮導寫回時按模板值補齊，需要修改請直接編輯 `config.yaml`。

### 方式二：手動複製模板

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

## 配置檔案

| 檔案 | 作用 |
| --- | --- |
| `server/config.yaml` | 後端配置（**已被 .gitignore 忽略**，不入庫）：埠/模式、PostgreSQL、Redis 與快取時長、JWT、初始管理員、附件儲存、日誌級別 |
| `server/config.example.yaml` | 配置模板，同時是初始化嚮導寫回配置時的底稿（嚮導會保留其註釋與排版） |
| `web/.env.local` | 前端本地配置（**已被 .gitignore 忽略**，不入庫）：`NEXT_PUBLIC_API_BASE`（後端介面地址） |
| `web/.env.example` | 前端環境變數模板，首次使用複製為 `.env.local` |

`server/config.yaml` 完整欄位：

```yaml
server:
  port: "8281"          # 必填。HTTP 監聽埠（字串）
  mode: "release"       # 必填。debug | release | test，其他值啟動報錯

database:
  # dsn 非空時優先使用（此時下面 host/port/user/dbname/sslmode 不再校驗）
  dsn: ""               # 例：postgres://postgres:123456@localhost:5432/mini-kanban?sslmode=disable
  host: "localhost"     # dsn 留空時必填
  port: 5432            # dsn 留空時必填
  user: "postgres"      # dsn 留空時必填
  password: "123456"    # dsn 留空時必填；含特殊字元無需手工轉義（拼裝時自動 URL 編碼）
  dbname: "mini-kanban" # dsn 留空時必填；庫需提前建立，服務只建表不建庫
  sslmode: "disable"    # dsn 留空時必填
  preferSimpleProtocol: false  # 出現 pgx 協議相容問題時設為 true（關閉擴充套件查詢協議）
  logLevel: "warn"             # silent | error | warn | info，其他值啟動報錯；嚮導不展示

redis:
  # 可選元件：host 留空即關閉快取（其餘欄位不再校驗，服務照常執行）
  host: "localhost"
  port: 6379            # host 非空時必填
  password: ""          # 無密碼留空
  db: 0
  cacheTTL: "30s"       # host 非空時必填且需 > 0：專案 / 任務列表快取時長
  userCacheTTL: "1m"    # host 非空時必填且需 > 0：使用者 / 部門列表快取時長

jwt:
  secret: "換成足夠隨機的長字串"   # 必填；修改後已簽發的 token 全部失效
  ttlHours: 24                     # 必填且需 > 0；token 有效期（小時）

seed:
  # 僅在「使用者表為空」時寫入初始管理員；寫入成功後自動清空本段（避免明文密碼留在磁碟）
  enabled: false        # 設為 true 時下面三項必填
  adminEmail: ""        # 初始管理員登入郵箱
  adminUsername: ""     # 初始管理員顯示名
  adminPassword: ""     # 初始管理員密碼，建議 8 位以上

storage:
  # 附件與富文字內嵌圖片目錄；嚮導不展示，只能在本檔案裡維護
  dir: "data/uploads"   # 必填；相對路徑按後端工作目錄解析，不存在會自動建立（需可寫）
  maxSizeMB: 10         # 必填且需 > 0；單次最多 10 個檔案，請求體上限 = maxSizeMB × 10
```

載入規則：

1. 查詢順序：`CONFIG_FILE` 環境變數 → `config.yaml` → `config.yml` → `configs/config.yaml` → `server/config.yaml` → `../config.yaml` → `../server/config.yaml`（因此從倉庫根目錄或 `server/` 目錄啟動都能找到）
2. 環境變數優先順序更高：`PORT`、`DB_DSN`、`REDIS_ADDR`（`host:port`）、`REDIS_PASSWORD`、`JWT_SECRET`
3. **沒有程式碼內兜底值**：必需項缺失或格式非法會直接報錯退出，不會靜默使用預設值；`redis.host` 留空是「關閉快取」的合法寫法
4. 其他環境變數：`SETUP_FORCE`（強制重新進入初始化嚮導，原配置備份為 `.bak`）、`SETUP_PORT`（嚮導自身埠，預設 8281）、`SNOWFLAKE_NODE_ID`（雪花 ID 節點號，0~1023；**多例項部署時必須各不相同**，否則各例項生成的 ID 會衝突，單機保持預設 0 即可）

更詳細的逐欄位說明（含取值範圍與注意事項）見 `server/config.example.yaml`。

前端 `web/.env.local`：

```bash
NEXT_PUBLIC_API_BASE=http://localhost:8281/api   # 埠需與 server.port 一致
```

> `NEXT_PUBLIC_*` 是**構建期內聯**的：改了地址必須重新 `npm run build`，執行時再設定環境變數無效（詳見「部署」章節）。

## 啟動

後端：

```bash
cd server
go run .              # 或先編譯：go build -o mini-kanban .
```

啟動日誌會列印埠、資料庫、Redis 與快取狀態；`GET /healthz` 返回服務與 Redis 狀態。

前端：

```bash
cd web
npm install
npm run dev           # http://localhost:3201
```

如後端不在本機 8281，修改 `web/.env.local`：

```bash
NEXT_PUBLIC_API_BASE=http://localhost:8281/api
```

## 初始管理員賬號

預設不寫入任何賬號（`seed.enabled: false`）。需要初始化時先開啟並填好三項：

```yaml
seed:
  enabled: true
  adminEmail: "admin@mail.example.com"
  adminUsername: "Admin"
  adminPassword: "至少 8 位"
```

規則：

- 僅當**使用者表為空**時建立；已有使用者則跳過，啟動日誌會列印建立的郵箱
- 建立成功後會自動把 `seed` 段清空（`enabled=false`，郵箱 / 使用者名稱 / 密碼置空），**避免初始密碼長期以明文留在磁碟上**；需要再次 seed 時重新填寫
- 其餘成員賬號由管理員在「使用者管理」頁面建立

## 附件儲存

任務附件、專案資料附件與富文字內嵌圖片統一落在 `storage.dir`：

- **單次最多 10 個檔案**，整個請求體上限 = `storage.maxSizeMB × 10`
- 副檔名白名單：圖片（png / jpg / jpeg / gif / webp / bmp / ico）、文件（pdf / txt / md / csv / doc / docx / xls / xlsx / ppt / pptx）、壓縮包（zip / 7z / rar / gz）、音影片（mp4 / mp3 / wav）。**`.svg` 不允許**（可攜帶指令碼）
- 落盤檔名由服務端生成：32 位隨機 hex + 白名單副檔名，使用者提供的檔名不進入路徑（防目錄穿越）
- 對外地址為 `/api/files/<32位隨機名>`，前端用 `fileUrl()` 按 API 基址補全
- 下載路由 **需要登入態**，但 URL 裡帶的**不是登入 token**，而是一張專用的「下載票據」：登入後前端調 `POST /api/files/ticket` 換取（`aud=file`，預設 2 小時），再拼成 `/api/files/<隨機名>?ticket=...`
- 這樣隔離的原因是富文字 `<img src>` 與附件連結由瀏覽器直接發起、帶不了 `Authorization` 頭，憑據只能進 URL：票據只能讀附件、不能呼叫其他介面，所以下載地址被轉發 / 記錄也不等於登入態洩露。登入 token 不會出現在任何 URL 中
- 票據 JTI 與登入會話的 jti 相同：登出後登入 token 進 Redis 黑名單，票據隨之立即失效
- **下載還要過資料許可權**（票據只說明「誰在下載」，不說明「能不能下這個檔案」）：
  服務端按檔案的引用關係反查歸屬，再套用對應入口的可見性，原則是**「即看即下」**——
  - 資料附件、資料正文裡的內嵌圖片：按**資料所屬專案的可見性**（能看到專案就能看資料）；
  - 任務附件、任務描述裡的內嵌圖片：按**任務可見性**（管理員 / 任務建立人 / 被指派人 / 專案建立人），
    即看板裡看得見這個任務的人才能下它的附件。

  引用它的資料已被刪除（資料 / 任務軟刪）時一律 404；還沒被任何資料引用的臨時檔案
  （編輯器裡剛貼上的圖片、上傳後尚未儲存的附件）只允許**上傳者本人**下載 —— 上傳時會在 `uploaded_files`
  登記上傳者。鑑權透過但沒有資料許可權時同樣返回 404，不洩露檔案是否存在
- 票據用**獨立金鑰**簽名：金鑰由 `jwt.secret` 經 HMAC-SHA256 單向派生（標籤 `mini-kanban/file-ticket/v1`），不新增配置項。因此登入 token 過不了下載校驗、票據也過不了登入校驗，即使某一側漏檢 `aud` 也無法互換使用；`aud=file` 與 `AuthRequired` 的顯式拒絕作為縱深防禦
- 兩類憑據統一限定 `HS256` 且必須帶 `exp`（防演算法混淆與永不過期票據）
- 前端 `AuthProvider` 負責換取票據並在臨近過期時靜默續期（`fileUrl()` / `resolveFileUrls()` 自動拼上當前票據，`TaskModal` 在票據續期後重新整理描述裡的圖片地址）；富文字儲存前 `toStoredHtml()` 會把票據剝掉，入庫的仍是純 `/api/files/<隨機名>`
- 響應帶 `X-Content-Type-Options: nosniff`；鑑權失敗為 `Cache-Control: no-store`，正常響應為 `private, max-age=3600`（避免被代理 / CDN 快取後繞過鑑權）。圖片 / 影片內聯展示，其他型別強制下載並儘量還原原始檔名

## 介面

| 方法 | 路徑 | 說明 |
| --- | --- | --- |
| POST | /api/auth/login | 登入，返回 JWT |
| GET | /api/auth/me | 當前使用者 |
| POST | /api/auth/logout | 登出（token 加入 Redis 黑名單） |
| PUT | /api/auth/profile | 修改本人資料（使用者名稱 / 郵箱 / 密碼，不含角色）；前端暫無入口 |
| GET | /api/files/:name | 附件下載（需登入 + 資料許可權：`?ticket=`，見「附件儲存」） |
| POST | /api/files/ticket | 換取附件下載票據（登入態，預設 2 小時） |
| GET | /api/departments | 部門列表 |
| GET | /api/departments/:id | 部門詳情 |
| POST | /api/departments | 新建部門（管理員） |
| PUT | /api/departments/:id | 編輯部門（管理員） |
| DELETE | /api/departments/:id | 刪除部門（管理員，有成員時拒絕） |
| GET | /api/users | 使用者列表 |
| GET | /api/users/:id | 使用者詳情 |
| POST | /api/users | 新建使用者（管理員） |
| PUT | /api/users/:id | 編輯使用者（管理員） |
| DELETE | /api/users/:id | 刪除使用者（管理員，含資料快照） |
| GET | /api/projects | 專案列表 |
| POST | /api/projects | 新建專案 |
| GET | /api/projects/:id | 專案詳情 |
| PUT | /api/projects/:id | 編輯專案 |
| DELETE | /api/projects/:id | 軟刪除專案 |
| GET | /api/projects/:id/tasks | 任務列表 |
| POST | /api/projects/:id/tasks | 新建任務 |
| GET | /api/projects/:id/docs | 專案資料列表（需求文件 / 會議紀要 / 設計稿 / 專案金鑰 / 測試賬密等） |
| POST | /api/projects/:id/docs | 新建專案資料（專案成員） |
| PUT | /api/docs/:id | 編輯專案資料（專案成員中的資料建立人 / 專案建立人 / 管理員） |
| DELETE | /api/docs/:id | 刪除專案資料（軟刪除，許可權同上，可在回收箱恢復） |
| POST | /api/uploads | 附件上傳（multipart/form-data，欄位名 `files`，最多 10 個） |
| GET | /api/tasks/:id | 任務詳情 |
| PUT / PATCH | /api/tasks/:id | 編輯任務 |
| DELETE | /api/tasks/:id | 軟刪除任務 |
| PATCH | /api/tasks/:id/move | 看板拖拽落位（`{status, index}`），服務端重排 position |
| GET | /api/trash | 回收箱：已刪專案 + 已刪任務 + 已刪專案資料（附同批數量 / 所屬專案標題） |
| POST | /api/trash/projects/:id/restore | 恢復專案（連同同批刪除的任務、資料與附件） |
| DELETE | /api/trash/projects/:id | 徹底刪除專案（物理刪除專案及其任務、資料、附件、成員關聯） |
| POST | /api/trash/tasks/:id/restore | 恢復任務（連同同批刪除的附件） |
| DELETE | /api/trash/tasks/:id | 徹底刪除任務（物理刪除任務及附件） |
| POST | /api/trash/docs/:id/restore | 恢復專案資料（連同同批刪除的附件） |
| DELETE | /api/trash/docs/:id | 徹底刪除專案資料（物理刪除資料及附件） |
| GET | /api/logs | 操作日誌分頁（管理員）：`targetType` / `action` / `operatorId` / `keyword` / `page` / `pageSize`（預設 20，上限 100） |
| GET | /api/logs/:id | 操作日誌詳情（管理員），物理刪除的條目含 `snapshot` 資料快照 |
| GET | /healthz | 健康檢查（含 Redis 狀態） |
| GET | /setup | 配置完成後的反饋頁 |

## 回收箱（軟刪除）

刪除專案 / 任務 / 專案資料只寫 `deleted_at`，資料留在庫中，預設查詢自動過濾（GORM 軟刪除）。
同一批被刪的記錄共用同一個 `deleted_at` 時間戳，恢復時按該時間戳精確還原：

- 恢復專案會帶回隨它一起刪掉的任務、專案資料與附件；在專案刪除**之前**單獨刪掉的任務 / 資料
  不會被誤恢復，它們仍留在回收箱的任務區 / 資料區，可以單獨恢復。
- 專案刪除時任務與資料跟著刪，它們只在回收箱的**專案條目**裡體現（`taskCount` / `docCount`
  就是恢復時會回來的任務數與資料數），不重複出現在任務區 / 資料區。
- 資料與任務一樣支援單獨恢復 / 徹底刪除；隨專案一起被刪的資料與專案同批次，恢復專案時一併還原。
- 「徹底刪除」是物理刪除（`Unscoped`），專案會連帶其下全部任務、專案資料、附件與成員關聯一起清掉，不可恢復，
  因此會在操作日誌裡留下整棵資料的快照（見下節）。
- 可見範圍與處置許可權對稱：管理員看全部，其他使用者只看自己有權處置的資料（專案按建立人；
  任務按建立人 / 被指派人 / 專案建立人；資料按資料建立人 / 專案建立人，且必須仍是專案成員）。

## 操作日誌

覆蓋**專案 / 任務 / 專案資料 / 使用者 / 部門**五類物件的操作留痕，寫庫表為 `operation_logs`，
在「操作日誌」頁面（僅管理員可見）按型別 / 動作 / 關鍵詞檢索。

- 記錄內容：操作人（id + 當時的使用者名稱快照）、物件型別與名稱、動作、變更摘要、時間。
  摘要由服務端按欄位級差異生成，例如 `狀態：待辦 → 進行中；負責人：張三 → 李四`；
  富文字正文只在變更時提示「描述已更新」，不把正文寫進日誌。
- 動作取值：`create` 新建、`update` 修改、`move` 看板拖拽改狀態（僅任務）、`delete` 刪除、
  `restore` 回收箱恢復、`purge` 回收箱徹底刪除。
- **物理刪除留資料快照**：`delete` 使用者、`delete` 部門、`purge` 專案 / 任務 / 專案資料這幾類操作會把被刪資料
  整份序列化進 `snapshot`（JSON 文字列），因為刪掉之後無從追溯：
  - 使用者快照：賬號資訊（不含密碼）+ 所屬部門名 + 被指派 / 建立任務數、擁有專案數、專案成員關聯數；
  - 部門快照：部門資訊 + 刪除時剩餘人數（有成員時不允許刪除，正常為 0）；
  - 專案快照：專案 + 成員 + 其下全部任務與專案資料及其附件（徹底刪除專案會連帶清掉這些資料）；
  - 任務快照：任務 + 全部附件 + 所屬專案標題；
  - 資料快照：專案資料 + 全部附件 + 所屬專案標題。
  軟刪除（`delete` 專案 / 任務 / 專案資料）的資料仍在庫中，只記動作、不留快照。
- 寫入時機：普通操作在業務成功後非同步補記（失敗只打服務端日誌，不影響介面結果）；
  物理刪除與它的快照寫在**同一個事務**裡，避免出現「資料刪了、快照卻沒留下」。
- 日誌只存操作人 id 與使用者名稱快照、不建外來鍵關聯，因此刪除使用者不會連帶刪掉它的歷史日誌。
- 列表介面不下發 `snapshot`（單條可能很大，徹底刪除專案會帶上全部任務），
  點「檢視」調詳情介面取單條。日誌介面不掛快取中介軟體，保證審計實時性。

## 資料庫遷移

服務啟動時自動執行，無需手動操作：

- 建表與補列：按 `Department → User → Project → Task → ProjectMember → TaskAttachment → ProjectDoc → ProjectDocAttachment → UploadedFile → OperationLog`
  順序建表；已存在的表按模型欄位補缺失列（規避 PostgreSQL 18 + pgx 組合下 AutoMigrate 的報錯）
- 補索引：專案 / 任務 / 附件 / 專案資料及其附件的 `deleted_at` 軟刪除索引按需建立
- 資料補齊：老任務缺失的 `creator_id` 回退為所屬專案建立人（冪等，只補空值）
- 首次出現 `project_members` 表時，把已有專案初始化為「全員可見」，避免歷史專案突然對所有人不可見

## 前端頁面與許可權

| 路由 | 頁面 | 可見範圍 |
| --- | --- | --- |
| /login | 登入 | 公開 |
| / | 首頁（按登入態跳轉） | — |
| /projects | 專案列表 | 登入使用者 |
| /projects/[id] | 四列看板（待辦 / 進行中 / 已完成 / 已驗收）+ 頂部「專案資料」入口（需求文件 / 會議紀要 / 設計稿 / 專案金鑰 / 測試賬密，支援富文字、附件與按標題檢索），任務可按「全部 / 我建立的 / 指派給我的」篩選 | 登入使用者 |
| /trash | 回收箱（專案 / 任務 / 專案資料，可恢復或徹底刪除） | 登入使用者（按許可權過濾） |
| /users | 使用者管理 | 管理員 |
| /departments | 部門管理 | 管理員 |
| /logs | 操作日誌 | 管理員 |

前端支援淺色 / 深色主題切換，token 存 `localStorage`（鍵名 `mini-kanban_token`），401 時自動清除並跳回登入頁。

## 多語言

介面支援**簡體中文**與 **English** 兩種語言，預設中文；語言切換按鈕在導航欄右上角（登入頁也有），
URL 不隨語言變化，重新整理 / 分享出去的連結對方看到的是自己 cookie 裡的語言。

### 語言怎麼存、怎麼生效

- **cookie `mini-kanban-lang`**（值 `zh-CN` / `en`，一年，`SameSite=Lax`，HTTPS 下加 `Secure`）是唯一口徑：
  根佈局在服務端讀它，用來定 `<html lang>` 與頁面標題，正文文案也跟著首屏就渲染成對應語言，
  不會出現「先閃一下中文再切成英文」。`localStorage` 裡存一份同名映象，只在 cookie 被清掉時兜底。
- 切換語言時 cookie / localStorage / `<html lang>` / 請求頭一起更新，無需重新整理頁面。
- 標籤頁標題固定是 **`頁面名 - {app.title}`**（如 `登入 - Mini-Kanban`）：頁面名在各自路由段的 `layout.tsx`
  裡用 `pageMetadata()` 宣告成鍵（頁面本身是 `"use client"`，不能匯出 `metadata`），客戶端跳轉 / 切換語言時
  由 `DocumentTitle` 按同一張路由對映重新整理，兩處共用 `lib/i18n/pageTitle.ts`。
- 代價是路由從靜態預渲染變為**按請求渲染**（用了 `cookies()`）。前端本來就依賴登入態取數，影響可忽略。
- **靜態匯出模式下這一條不成立**（`output: "export"` 沒有請求上下文）：首屏只能按預設語言渲染，
  由 `<head>` 裡的內聯指令碼先糾正 `<html lang>`，掛載後 `I18nProvider` 再按 cookie / localStorage
  切換文案 —— 因此非預設語言會閃一下。釋出包用的就是靜態匯出（頁面要隨包分發）；
  如果更在意首屏不閃，就用倉庫根的 Docker Compose 那套部署，服務端渲染版本不受這條限制。

### 前端文案

| 檔案 | 作用 |
| --- | --- |
| `web/lib/i18n/zh-CN.ts` | **源語言**，同時是全部文案鍵的定義（`域.用途` 扁平點號，`{name}` 是插值佔位符） |
| `web/lib/i18n/en.ts` | 英文，型別是 `Record<MessageKey, string>` —— 中文新增 / 改名的鍵這裡沒跟上，`npm run build` 直接報缺鍵 |
| `web/lib/i18n/index.ts` | 語言常量、字典查詢與插值、cookie / localStorage 讀寫、`tNow()` |
| `web/lib/i18n/pageTitle.ts` | 路由 → 頁面名文案鍵的對映，以及 `頁面名 - {app.title}` 的拼裝（服務端與客戶端共用這一張表） |
| `web/lib/i18n/server.ts` | 服務端讀語言 cookie（`currentLang()`）與路由段 layout 用的 `pageMetadata()` |
| `web/components/DocumentTitle.tsx` | 客戶端按當前語言重新整理標籤頁標題（靜態匯出產物構建期只能寫死預設語言） |
| `web/components/I18nProvider.tsx` | 全域性語言 Provider：`useI18n()` 取 `{ lang, locale, setLang, t }` |
| `web/lib/labels.ts` | 業務列舉 → 文案（看板列名、角色、操作日誌的物件型別 / 動作），未知值原樣回顯 |

用法約定：

- **渲染時**用 `const { t } = useI18n()`，帶引數時 `t("key", { name })`；
- **非渲染路徑**（`lib/api.ts` 的兜底文案、非同步回撥裡的失敗提示）用 `tNow(...)` ——
  它是模組級的當前語言，不參與渲染，因此不會因為切語言觸發無謂的重新取數；
- 日期時間統一走 `lib/date.ts` 的 `formatDateTime(iso, locale)`，`locale` 取 `useI18n().locale`，
  中文輸出 `2026/09/20 14:30`、英文輸出 `09/20/2026, 14:30`；
- 涉及數量的英文文案刻意寫成 `Tasks: {count}` 這類，避開單複數變化，因此沒有引入 ICU 複數規則。

新增一種語言：加一份字典 → 在 `lib/i18n/index.ts` 的 `LANGS` / `DICTS` / `LANG_LABELS` /
`LANG_LOCALES` 裡登記 → 切換控制元件自動多出一個選項。

### 後端介面文案

後端的 `error` / `message` 也按語言返回，協商順序（`internal/i18n`）：

1. **`Accept-Language` 頭** —— 前端 `lib/api.ts` 的每個請求都會帶上當前介面語言；
2. **`mini-kanban-lang` cookie** —— 附件下載（`/api/files/xxx?ticket=`）由瀏覽器直接發起、帶不了自定義頭，靠它兜底；
3. 都沒有 → 中文。

英文目錄在 `server/internal/i18n/messages.go`，**以中文原文為鍵**，因此 `fail` / `badRequest` /
`okMessage` 這些出口內部統一查表即可，呼叫方零改動；帶動態內容的訊息（`儲存失敗: <詳情>`、
`一次最多上傳 10 個檔案`、`附件「a.png」不存在`）用「字首 + 尾綴 + `「」`」三類規則拼裝。
未收錄的原文（例如資料庫返回的英文錯誤）原樣透傳。

護欄：`go test ./internal/i18n/` 會掃描 `internal/handlers` 裡寫死的中文文案，
逐條確認能在英文目錄裡找到譯文 —— 新增 / 修改文案忘記翻譯會在測試裡直接失敗。

**已知邊界**（刻意不做）：

- 操作日誌的 `summary` / 變更項與資料快照是**寫入時落庫的文字**，保持中文：
  否則同一條日誌會隨檢視者語言變化，審計語義就不穩了。日誌頁其餘文案（表頭、篩選、分頁）都是英文的。
- 初始化配置嚮導 `/setup` 是一次性運維頁面，保持中文。
- `Provider` 裡 `useXxx 必須在 XxxProvider 內使用` 這類斷言只會在控制檯出現，不參與翻譯。


## Redis 的用途

1. **快取中介軟體**：`middleware.Cache(ttl, namespace)` 快取 GET 響應（按使用者 + 路徑 + 查詢串生成 key），
   命中時返回 `X-Cache: HIT`；所有寫操作透過 `cache.DeletePrefix("httpcache:<namespace>:")` 主動失效。
2. **登出黑名單**：`auth:blacklist:<jti>` 儲存至 token 自然過期，鑑權中介軟體會拒絕已登出 token。
3. Redis 不可用時自動降級為無快取模式，服務仍可執行（`GET /healthz` 返回 `redis` 狀態）。

## 自檢與測試

改動後端或前端後，建議先跑這幾項再部署：

```bash
# 後端：靜態檢查 + 單元測試
cd server
go vet ./...
go test ./...
```

其中 `internal/i18n` 的測試是一道**文案護欄**：它掃描 `internal/handlers` 裡寫死的中文文案，
逐條確認英文目錄（`internal/i18n/messages.go`）裡有對應譯文 —— 新增介面文案忘了翻譯會在這裡失敗。

```bash
# 前端：型別檢查（不產出檔案）
cd web
npx tsc --noEmit
```

前端的多語言完整性也由型別系統兜底：`lib/i18n/zh-CN.ts` 匯出 `MessageKey` 聯合型別，
`en.ts` 宣告為 `Record<MessageKey, string>`，所以**英文缺鍵、元件裡寫錯文案鍵**都會在 `tsc --noEmit`（以及 `npm run build`）階段直接報錯。
同理，`lib/api.ts` 的請求引數與返回型別寫錯也會被這一步攔下。

服務起來後的快速冒煙：

```bash
curl http://localhost:8281/healthz    # {"redis":true,"status":"ok"}
```

## 部署

釋出的**唯一產物是「開箱即用」包**：後端可執行檔案 + `web-dist/` 頁面目錄，
由同一個程序、同一個埠提供介面與頁面。目標機不需要 Node、不需要額外的靜態目錄，也不需要 nginx 與 CORS。

> 為什麼不把前端 `go:embed` 進二進位制：那樣每個平臺的可執行檔案會憑空多出約 1MB，
> 六平臺交叉編譯時這部分要被反覆寫進每一個產物，而且改一次前端就要把 Go 側全量重編一遍。
> 改成隨包分發後，可執行檔案裡只剩業務程式碼，頁面檔案還能被閘道器 / CDN 直接讀盤快取。
> 代價是 `web-dist/` 必須與可執行檔案同目錄（可以用 `WEB_DIST_DIR` 指向別處）；
> 缺了它服務仍能啟動，但只提供介面、非 `/api` 路徑返回 404。

> 原始碼層面仍支援把前端單獨跑起來（`npm run dev` 聯調、或倉庫根的 Docker Compose），
> 但**釋出只發這一個包**，避免使用者挑錯。Docker Compose 那套是給容器化部署用的另一條路，與釋出包無關。

### 自己打一個釋出包

```bash
bash build-web-static.sh          # 1) 前端靜態匯出到 web/out，再拷進 server/web-dist
bash build.sh -v v1.0.0           # 2) 交叉編譯 + 打包（把 web-dist/ 拷進每個歸檔）
```

Windows 上等價於：

```bat
build-web-static.bat                 :: 靜態匯出並拷進 server\web-dist
bash build.sh -v v1.0.0              :: Git Bash 裡跑；也可以 cd server && go build -o mini-kanban.exe .
```

產物落在 `dist/`，每個平臺一個歸檔，頂層目錄名與歸檔名一致：

```
dist/
├── mini-kanban-v1.0.0-windows-amd64.zip
├── mini-kanban-v1.0.0-windows-arm64.zip
├── mini-kanban-v1.0.0-linux-amd64.tar.gz
├── mini-kanban-v1.0.0-linux-arm64.tar.gz
├── mini-kanban-v1.0.0-darwin-amd64.tar.gz
├── mini-kanban-v1.0.0-darwin-arm64.tar.gz
└── checksums.txt                  # 全部歸檔的 SHA256（sha256sum -c 可校驗）
```

每個歸檔解開是：

```
mini-kanban-v1.0.0-linux-amd64/
├── mini-kanban(.exe)     可執行檔案（只有後端程式碼）
├── web-dist/              前端靜態產物，必須與可執行檔案同目錄
├── start.sh / start.bat  啟動指令碼（列印版本號，異常退出時列出常見原因）
├── version.txt           本包版本號
├── config.example.yaml   配置模板（複製為 config.yaml 後按環境修改）
├── README.md / README.en.md
└── LICENSE
```

- **`-v` 省略時自動取最近的 git 標籤**，所以在打好標籤的工作區裡直接 `bash build.sh` 也不會漏版本號；
  版本號裡的非法字元會被拒絕（避免拼出奇怪的檔名）。
- **前端靜態產物是硬前提**：`server/web-dist/index.html` 不存在時 `build.sh` 直接報錯退出，
  不會再產出「只有介面沒有頁面」的包。
- **改前端不用重編 Go**：`web-dist` 不參與編譯，`next build` 完直接 `bash build.sh` 即可，
  Go 側命中編譯快取；這比 `go:embed` 方案快得多（embed 的內容一變，整個包就要重編）。
- Git Bash 預設不帶 `zip` 命令，`build.sh` 會自動退回用 Python 的 `zipfile` 打 Windows 歸檔，產物結構一致。
- 推送 `v*` 標籤即觸發 GitHub Actions：跑靜態檢查與單元測試 → 構建前端 → 交叉編譯 → 校驗產物
  （校驗和、數量、檔名是否帶標籤、包內是否帶 `web-dist/index.html` 與 `README.en.md`）
  → 解包真跑一遍冒煙（頁面、深鏈、快取頭、介面 404）→ 建立 / 更新 Release。

### 跑起來是什麼樣

```
瀏覽器 → mini-kanban :8281
          ├── /api/*    → 介面
          ├── /healthz  → 健康檢查
          ├── /setup    → 初始化嚮導
          └── 其他      → web-dist 裡的靜態頁面（未命中的前端路由回退 index.html）
```

- 實現見 `server/webstatic.go`：啟動時 `ResolveWebDist()` 解析一次目錄（預設工作目錄下的 `web-dist/`，
  可用 `WEB_DIST_DIR` 覆蓋），結果掛在 `r.NoRoute` 上，因此 `/api`、`/healthz`、`/setup` 仍由原路由優先匹配；
  `/api/` 字首未命中時返回 JSON 404，不會把頁面當介面響應返回。
- 目錄缺失時自動降級：啟動日誌寫明 `前端: 未找到（僅提供介面）` 並列印它找的那個絕對路徑，
  非 `/api` 路徑返回 404，介面行為完全不變。**釋出包不會出現這種狀態**，`build.sh` 已經卡住了這一步。
- 快取策略：`/_next/static/*` 檔名帶內容雜湊，發 `max-age=31536000, immutable`；
  頁面與其他產物發 `no-cache`，避免換版本後瀏覽器拿著舊 `index.html` 去請求已刪除的 chunk。
- 看板深鏈 `/projects/<id>` 是唯一帶動態段的路由：靜態匯出要求動態段至少預渲染一條路徑才肯構建
  （Next 用「預渲染路徑數 > 0」判斷頁面有沒有 `generateStaticParams`），所以 `web/app/projects/[id]/page.tsx`
  預渲染了一份佔位頁 `/projects/_/`；`server/webstatic.go` 把真實的 `/projects/<id>` 對映到這份佔位頁，
  專案 id 由前端從位址列解析（靜態產物裡 `useParams()` 拿到的只是構建期佔位值，
  見 `ProjectBoard.tsx` 的 `resolveProjectId`）。
- 因為頁面與介面同源，**不再需要 nginx 與 CORS**；要加 TLS / gzip 再在前面掛一層閘道器即可（見「前面掛閘道器」）。

### 方式一：Docker Compose 一鍵部署（容器化場景）

想用容器跑（自帶 PostgreSQL / Redis，或想把前端單獨擴多例項），走倉庫根目錄的 `docker-compose.yml`：
一次拉起 5 個服務 —— `nginx`（統一入口）、`web`（Next.js）、`server`（Go）、`postgres`、`redis`。

```bash
# macOS / Linux
cd <倉庫根目錄>
cp .env.example.docker-compose .env       # 改掉裡面的資料庫 / Redis 密碼與 JWT_SECRET
docker compose up -d --build
# 瀏覽器開啟 http://<伺服器IP>:8080（宿主機埠由 .env 的 HTTP_PORT 決定）
```

```powershell
# Windows PowerShell
cd <倉庫根目錄>
Copy-Item .env.example.docker-compose .env  # 改掉裡面的資料庫 / Redis 密碼與 JWT_SECRET
docker compose up -d --build
# 瀏覽器開啟 http://<伺服器IP>:8080（宿主機埠由 .env 的 HTTP_PORT 決定）
```

Linux 伺服器上也可以直接跑一鍵指令碼，它會把上面幾步連起來做（生成 `.env`、隨機化資料庫 / Redis 密碼與 JWT 金鑰、**生成帶隨機密碼的初始管理員**、挑一個未被佔用的宿主機埠、構建並啟動），並且**重複執行安全**：

```bash
cd <倉庫根目錄>

# 下面兩條二選一，首次部署前執行其中一條即可（管理員密碼均由指令碼隨機生成）
bash deploy/deploy.sh                                                # 預設管理員：admin@mail.example.com / Admin
bash deploy/deploy.sh --admin-email me@corp.com --admin-user kevin   # 自定義管理員的郵箱與顯示名
```

管理員賬號相關說明：

- 上面兩條命令是**二選一**，不是先後執行：指令碼只負責把郵箱 / 顯示名寫進 `.env`，而賬號本身是後端首次啟動（使用者表為空）時才建立的。
- 密碼由指令碼隨機生成（16 位字母數字，排除了 `0 O 1 l I` 這類易混淆字元），同時也寫進 `.env` 的 `SEED_ADMIN_*`，忘了隨時可查。**只有全新資料庫（資料卷不存在）時**才會把密碼列印到終端；重複部署時因為賬號已經建好，只提示一句"賬號已存在"，不再輸出密碼。
- 該賬號**只在使用者表為空時建立一次**：重複執行指令碼、後續重啟或重建容器都不會重建它，也不會改掉已生效的密碼；後端建立成功後還會把容器內 `config.yaml` 的 seed 段自動清空。
- 想換郵箱或顯示名，要在**首次啟動之前**用 `--admin-email` / `--admin-user` 指定；賬號已經建立後再改，只能在「使用者管理」頁面裡改。
- 不用指令碼、走手工部署時，等價做法仍然是把 `.env` 的 `SEED_ENABLED` 改為 `true` 並填好 `SEED_ADMIN_*` 三項。
- **忘了管理員密碼**：不必清庫，執行 `bash deploy/deploy.sh --reset-admin` 即可。它會把 `SEED_ADMIN_EMAIL` 指定的賬號重置為管理員並換一個隨機密碼（該郵箱已不存在時重新建立），其他資料不受影響，新密碼會列印在部署結果裡。對應的後端配置是 `seed.forceReset`（環境變數 `SEED_FORCE_RESET`），只在本次啟動生效——指令碼啟動後會自動把它改回 `false`，後端處理完也會清空容器內 `config.yaml` 的 seed 段。

要點：

- **宿主機埠衝突**：`HTTP_PORT`（預設 8080）是整份 compose 裡**唯一**會對映到宿主機、因而可能與既有服務衝突的埠；`SERVER_PORT` / `DB_PORT` / `REDIS_PORT` / 前端 3201 都是容器內埠，與宿主機上跑著什麼無關。衝突時改 `.env` 的 `HTTP_PORT`（只重建 nginx：`docker compose up -d nginx`），或用 `bash deploy/pick-port.sh` 自動挑一個空閒埠寫回 `.env`（`deploy/deploy.sh` 已內建這一步）。
- **後端配置由環境變數生成**：容器入口指令碼 `server/docker-entrypoint.sh` 把 `SERVER_*` / `DB_*` / `REDIS_*` / `JWT_*` / `SEED_*` / `STORAGE_*` 渲染成 `/app/config.yaml`。這麼做是因為後端只認配置檔案（`config.seed` 段沒有環境變數入口）。想完全自定義就掛載自己的 `config.yaml` 到 `/app/config.yaml`，指令碼檢測到檔案已存在會跳過生成。
- **前端介面地址是構建期內聯的**：`NEXT_PUBLIC_API_BASE` 經 compose 的 build args 傳入，預設 `/api`（同源反代，天然規避 CORS）。改了地址必須重建前端：`docker compose up -d --build web`。
- **`web/.dockerignore` 排除了 `.env*`**：本地 `.env.local`（`http://localhost:8281/api`）的優先順序高於構建引數，一旦隨構建上下文進映象就會把 localhost 寫進產物。
- 資料都在命名卷裡：`pgdata`（資料庫）、`redisdata`（AOF）、`uploads`（附件）。`docker compose down` 不刪卷，需要清空時加 `-v`。
- **資料庫密碼只在資料卷首次初始化時生效**：`POSTGRES_PASSWORD` 是 PostgreSQL 在資料卷為空時寫入的，之後改 `.env` **不會**同步到庫裡，表現是服務反覆重啟並報 `password authentication failed for user "mini-kanban" (SQLSTATE 28P01)`。庫裡沒有需要保留的資料時，最省事的修法是：
  ```bash
  bash deploy/deploy.sh --reset      # 先 docker compose down -v，再重新部署（資料庫 / Redis / 附件都會刪除）
  ```
  `deploy/deploy.sh` 檢測到「資料卷已存在 + 密碼仍是模板佔位值」時會攔下並提示這一點；若是你自己改過密碼導致對不上，則先 `docker compose down -v` 清卷，或把 `.env` 改回初始化時用的值。要在**不刪資料**的前提下改密碼：`docker compose exec postgres psql -U postgres -c "ALTER USER \"mini-kanban\" WITH PASSWORD '新密碼';"`（容器內本地連線是 trust，不需要密碼）。
- 常用命令：`docker compose ps` 看健康狀態、`docker compose logs -f server` 看後端日誌、`docker compose up -d --build <服務名>` 只重建單個服務。
- 附件上限有兩處要同步：`.env` 的 `STORAGE_MAX_SIZE_MB`（請求體上限 = 該值 × 10）與 `deploy/nginx.conf` 的 `client_max_body_size`。
- **映象拉取慢或失敗**：國內直連 Docker Hub 經常超時，報 `TLS handshake timeout` 或 `EOF`，這與本專案的配置無關，重試或配個加速器即可。Docker Desktop 在 Settings → Docker Engine 里加 `"registry-mirrors": ["https://<加速地址>"]` 後 Apply & Restart；Linux 伺服器改 `/etc/docker/daemon.json` 後 `systemctl restart docker`。也可以先單獨把基礎映象拉好再啟動：`postgres:16-alpine`、`redis:7-alpine`、`nginx:1.27-alpine`、`golang:1.22-alpine`、`node:20-alpine`、`alpine:3.20`。

### 方式二：手工部署前端（僅在需要把前端單獨跑時）

釋出包已經把前端頁面目錄（`web-dist/`）一起帶著，正常部署用不到這一節。只有下面這些情況才需要單獨跑前端：
本地 `npm run dev` 聯調、或用 Docker Compose 那種多服務部署。

```bash
# macOS / Linux
cd web
# NEXT_PUBLIC_* 是構建期內聯的：必須先寫好 .env.production 再 build，構建後改環境變數無效
echo "NEXT_PUBLIC_API_BASE=/api" > .env.production
npm ci
npm run build
pm2 start npm --name mini-kanban-web -- run start    # 或 systemd
```

```powershell
# Windows PowerShell
cd web
# 注意別用 > 重定向：PowerShell 會寫成 UTF-16，Next 讀出來是亂碼。
# 這裡內容本身是純 ASCII，用 ascii 編碼不會帶 BOM。
Set-Content -Path .env.production -Value "NEXT_PUBLIC_API_BASE=/api" -Encoding ascii
npm ci
npm run build
# pm2 在 Windows 上啟動 npm 指令碼不穩定，建議改用 NSSM：
#   nssm install mini-kanban-web "C:\Program Files\nodejs\node.exe" "node_modules\next\dist\bin\next start -p 3201"
```

- `NEXT_PUBLIC_API_BASE=/api` 配合同域反代，可順帶規避 CORS
- 生產執行需要 `.next/` + `node_modules/`（生產依賴）+ `package.json`
- 建議在目標機器上執行 `npm ci && npm run build`：`.next` 含構建期產物，跨 Node 版本 / 架構搬運可能不相容
- 想改成靜態產物（交給任意靜態伺服器，或由後端從目錄讀取）：`bash build-web-static.sh`

### 方式三：手工部署後端

```bash
# 頁面同源提供（釋出包形態）：先出前端靜態產物到 server/web-dist，再編譯；
# 執行時把它與二進位制放在一起（或用 WEB_DIST_DIR 指定）
bash build-web-static.sh
cd server && go build -o mini-kanban .

# 只提供介面（頁面由前端服務負責，或本地聯調）
cd server && go build -o mini-kanban .
```

```powershell
cd server
# 若在 Windows 上交叉編譯出 Linux 二進位制
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o mini-kanban .
```

- 一併部署 `config.yaml` 與附件目錄（`storage.dir`，預設 `data/uploads`）
- `storage.dir` 的相對路徑按**後端工作目錄**解析，注意啟動目錄
- 生產環境不要誤開 `SETUP_FORCE`

### 前面掛閘道器（可選）

一體化部署本身就是完整入口，不需要 nginx。要加 TLS、gzip 或多域名，再在前面掛一層閘道器即可 —— 
此時只有一個上游埠：

```nginx
server {
    listen 80;
    server_name mini-kanban.example.com;

    # 頁面與介面同一個二進位制、同一個埠，全部轉發過去即可
    location / {
        proxy_pass http://127.0.0.1:8281;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        client_max_body_size 100m;             # 需不小於 storage.maxSizeMB × 10（預設 10 → 100MB）
    }
}
```

### 上線前檢查

- `jwt.secret` 換成隨機長字串，`server.mode` 設為 `release`
- `config.yaml` 含資料庫明文密碼，檔案許可權收緊；它已被 `.gitignore` 忽略，不要入庫
- 後端 CORS 中介軟體會回顯請求的 `Origin`（便於本地聯調），同域反代部署時不受影響；如前後端分域部署且需收緊來源，需自行調整 `internal/middleware/cors.go`
- `/setup` 嚮導不應暴露到公網（配置就緒後訪問只會顯示反饋頁，但仍建議在閘道器層限制）
- 附件下載是「票據 + 隨機檔名」兩道防線：URL 必須帶上登入後換取的下載票據（`?ticket=`，見「附件儲存」），
  落盤名是 32 位隨機 hex；不要把檔名改成可預測命名，也不要去掉 `middleware.FileTicketRequired()`

---

[MIT](LICENSE) © 2026 kite88
