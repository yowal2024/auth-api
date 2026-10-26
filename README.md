# Auth API

基于 Go、Vue 3、MySQL 和 Redis 的授权管理系统，支持用户与卡密验证、应用管理、SDK 接口和后台管理。生产构建将前端静态资源和数据库迁移嵌入 Go 程序，Linux 部署只需一个可执行文件及运行配置。

当前项目版本为 **0.1.1**，使用 **Go 1.27.3** 构建。这是根据已有程序重建并持续维护的源码，原客户端及实际业务 APK 的完整兼容性仍需按业务场景验收。

## 主要功能

- 管理应用、用户、用户分组、卡密、公告、黑名单、版本和扩展文件。
- 各应用下的业务记录使用固定的应用内编号；最近操作、应用操作日志及文件列表按当前应用隔离，多标签页可分别选择应用。
- 提供 SDK 握手、登录、注册、充值、心跳、云函数及签名校验相关接口。
- 管理 APK 原包、签名、克隆与 DELTA1 差异补丁。
- 支持后台角色权限、自定义登录入口、安全事件、客户端会话管理和自动备份。
- IP 定位服务、共用 KEY 和本地 IP 库由系统设置统一管理，应用各自保留地域绑定开关与距离限制。
- 管理员账号采用单登录：新登录成功后，该账号的旧会话在下一次请求时失效；不同管理员互不影响。
- HTTP / HTTPS 在同一个指定端口上切换，支持证书管理和应用 mTLS。
- 已安装实例启动时自动应用待执行的数据库迁移。
- 管理页脚显示版本号与构建时间；构建时间由打包流程生成，显示为 UTC+8。

普通管理员可以管理用户、卡密、公告、黑名单及自己的资料；超级管理员拥有应用配置、证书、系统设置、备份和管理员管理等权限。当前角色权限覆盖所有应用，未提供按应用分配管理员的功能。

## 环境要求

| 组件 | 开发与验证环境 |
| --- | --- |
| Go | 1.27.3，`go.mod` 指定工具链 |
| Node.js | 推荐 24，与 Docker 构建环境一致 |
| MySQL | 8.4 |
| Redis | 7.4 |
| Docker Compose | 可选，用于本地环境和集成测试 |

## 项目结构

```text
backend/
  cmd/auth-api/          程序入口
  configs/              空白配置模板
  internal/             业务逻辑及测试
  migrations/           SQL 迁移及嵌入入口
frontend/
  src/                  Vue 管理前端
  scripts/              前端检查、浏览器测试
  package.json          项目版本与前端依赖
scripts/                Linux 打包、启动验证与集成测试
docker/                 容器初始配置
docs/                   启动兼容性和维护文档
```

仓库保存源码、依赖锁文件、配置模板和数据库迁移。`dist/`、`node_modules/`、编译后的前端、运行配置、上传文件、密钥和备份由 `.gitignore` 排除；本地文件会保留。

## Docker 本地启动

在仓库根目录执行：

```sh
docker compose up -d --build
```

打开 `http://127.0.0.1:18080/install`，按安装向导配置：

| 设置 | 本地示例值 |
| --- | --- |
| MySQL 主机 / 端口 | `mysql` / `3306` |
| MySQL 用户 / 密码 | `root` / `local_test_only` |
| 数据库 / 表前缀 | `auth_api` / `a_` |
| Redis 地址 / DB | `redis:6379` / `0` |
| Redis 密码 | 留空 |

安装时自行创建管理员并设置后台入口。上述示例仅用于本地开发，Compose 的服务端口绑定在 `127.0.0.1`。数据库、配置和上传文件保存在 Docker 数据卷中。

查看日志使用 `docker compose logs -f app`；停止服务使用 `docker compose down`，数据卷会保留。

## 本地开发

可先用 Docker 启动依赖：

```sh
docker compose up -d mysql redis
```

后端在一个终端中运行：

```sh
cd backend
cp configs/config.example.yaml configs/config.yaml
go mod download
go run ./cmd/auth-api -config configs/config.yaml -port 8080
```

PowerShell 中将复制命令替换为 `Copy-Item configs/config.example.yaml configs/config.yaml`。已有运行配置时直接使用，不要再次复制覆盖。

前端在另一个终端中运行：

```sh
cd frontend
npm ci
npm run dev
```

打开 `http://127.0.0.1:5173/install`，开发服务器将 `/api` 和 `/uploads` 转发到 `127.0.0.1:8080`。使用上述 Docker 数据库时，安装向导填写 MySQL `127.0.0.1:18306`、Redis `127.0.0.1:18379`，其余使用表格中的本地示例值。

修改代理地址时，将 `frontend/.env.example` 复制为 `frontend/.env.local`，设置 `API_PROXY_TARGET` 后重启前端开发服务器。

## 构建 Linux amd64 单文件

### Windows PowerShell

确保 `go`、`node`、`npm` 可用，在仓库根目录执行：

```powershell
./scripts/build-linux-amd64.ps1
```

脚本检查并构建前端、运行 Go 测试，然后使用 Go 1.27.3 交叉编译 Linux amd64 程序。产物位于：

```text
dist/auth-api-linux-amd64-v<版本>-<UTC构建时间>
```

程序包含前端和数据库迁移，`CGO_ENABLED=0`、`GOAMD64=v1`。脚本输出文件大小及 SHA-256，并拒绝覆盖同名产物。

### Linux / macOS

在仓库根目录执行：

```sh
cd frontend
npm ci
npm run check
npm run build -- --outDir ../backend/internal/webassets/dist --emptyOutDir

cd ../backend
go test -mod=readonly ./...
go test -mod=readonly -tags=embeddedweb ./cmd/auth-api ./internal/webassets
mkdir -p ../dist
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v1 \
  go build -mod=readonly -trimpath -tags=embeddedweb -ldflags='-s -w' \
  -o ../dist/auth-api-linux-amd64 ./cmd/auth-api
```

版本来自 `frontend/package.json`。每次前端构建自动生成时间，并同时写入管理页脚和嵌入的 `build-info.json`；刷新页面或重启服务不会改变构建信息。

## Linux 部署与升级

建议目录结构：

```text
/opt/auth-api/
  auth-api-linux-amd64
  configs/config.yaml
  data/
  uploads/
```

将构建产物复制到服务器并重命名为 `auth-api-linux-amd64`。首次安装可使用 `backend/configs/config.example.yaml` 作为 `configs/config.yaml`，保留 `installed: false`，再通过浏览器安装向导配置 MySQL、Redis 和管理员。

启动命令：

```sh
chmod +x /opt/auth-api/auth-api-linux-amd64
/opt/auth-api/auth-api-linux-amd64 -port 8443
```

程序默认前台运行。首次安装尚未开启 HTTPS 时，访问 `http://服务器地址:8443/install`；开启 HTTPS 后改用 `https://`。端口号本身不会启用 TLS。

需要后台运行时可以使用：

```sh
nohup /opt/auth-api/auth-api-linux-amd64 -port 8443 \
  >> /opt/auth-api/auth-api.log 2>&1 < /dev/null &
```

| 参数 | 行为 |
| --- | --- |
| `-port 8443` | 监听所有网卡的 `:8443`，优先级最高 |
| `-addr 127.0.0.1:8080` | 未指定 `-port` 时覆盖配置中的监听地址 |
| `-config /path/config.yaml` | 指定配置文件 |
| `-migrate` | 仅检查和执行数据库迁移，然后退出 |

不传地址参数时使用 YAML 中的 `server.addr`，空地址回退为 `:8080`。命令行参数不会改写配置。

省略 `-config` 时，依次查找程序旁的 `configs/config.yaml`、`config.yaml`，再查启动工作目录。运行路径按配置位置解析；配置位于 `configs/` 时，其上一级作为服务根目录。因此可以从 `/root` 启动 `/opt/auth-api/auth-api-linux-amd64`。

HTTP 和 HTTPS 共用选定的监听端口，每次只提供选中的协议。旧设置 `https_addr` 不另开监听端口。HTTPS 依赖数据库中的证书、加密私钥及原来的 `data/master.key`。

升级已有实例时，先备份并停止旧服务，替换二进制，保留 `configs/`、`data/`、`uploads/`、MySQL 和 Redis，再按原参数启动。`installed: true` 的实例会在监听前自动应用待执行迁移，无需手动执行 `-migrate`。迁移失败时程序报错退出；中断的迁移不会自动重放。

管理员会话兼容原版 Redis 字符串索引和上一版集合索引，无需清空 Redis。上一版集合会话升级后需要重新登录；成功登录会原子撤销该账号的旧会话。SDK 用户与卡密仍使用各自的会话规则。

## 检查与测试

```sh
# 后端单元测试；需要外部依赖的集成测试在未配置时跳过
cd backend
go test -mod=readonly ./...

# 前端源码检查
cd ../frontend
npm ci
npm run check
```

前端浏览器测试使用 Playwright，运行 `npm test` 或对应的 `npm run test:*` 脚本。默认浏览器路径为 Windows Chrome，其他安装位置设置 `CHROME_PATH`。真实接口回归脚本使用独立测试账号与数据，请先阅读 [测试与运行维护](docs/development-notes.md)。

Linux 单文件启动验证在 Linux / WSL 中运行，需要 Docker：

```sh
python3 scripts/verify-startup-linux.py \
  --binary dist/auth-api-linux-amd64 \
  --report dist/STARTUP-VERIFICATION.json
```

将 `--binary` 替换为实际产物路径。脚本创建并清理专用 MySQL / Redis 测试容器，覆盖安装、配置路径、端口优先级、前端嵌入、构建信息、单登录、数据库升级、HTTP / HTTPS 切换和错误退出。可选 `--original /path/to/original-binary` 用于与原程序对照。

Windows 下也可以执行 `./scripts/test-docker.ps1` 运行隔离的容器集成测试。

## 更多说明

- [原版启动与会话行为核对](docs/original-startup.md)：参数、配置目录、自动迁移和 Redis 兼容性。
- [测试与运行维护](docs/development-notes.md)：本地 IP 库、专项回归、实现差异、安全事件、备份和恢复。
- [配置模板](backend/configs/config.example.yaml)：空白数据库连接和默认运行参数。

项目保留了用于兼容 DELTA1 补丁的 Go 压缩实现，其来源和许可证见 [flate125 说明](backend/internal/buildpool/bsdiff/internal/flate125/README.md)。
