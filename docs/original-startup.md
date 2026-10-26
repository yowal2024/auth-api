# 原版启动行为核对

核对对象：用户提供的原版 Linux amd64 二进制 `auth-api-linux-amd64`，构建版本 Go 1.25.11。

SHA-256：`12854376494b47426da3c3ccd1abdc8da48dd161bc2e35e0b1c859516c0dbcc8`。

虽然 ELF 符号表已剥离，Go 的 `.gopclntab` 仍包含函数名与地址。使用该表定位函数，再反汇编对应代码，得到以下结论：

- `main.main`（`0x1043960`）注册 `-config`、`-port`、`-addr`。默认配置路径为 `configs/config.yaml`，空监听地址回退到 `:8080`。
- `main.main` 中 `0x1043c38` 起的分支在端口大于零时构造 `":" + strconv.Itoa(port)`。因此原版 `-port 8443` 监听 `:8443`，不会保留配置中的 `127.0.0.1`。
- `main.main.func1`（`0x10449e0`）读取系统设置的 `https_enabled`；开启时读取加密私钥，使用主密钥解密，并调用 `tlsmgr.(*Manager).ServerTLSConfig`。HTTP 和 HTTPS 分支都把同一个捕获的服务地址写入 `http.Server.Addr`。没有使用 `Settings.HTTPSAddr` 另开监听器。
- `main.main.func1.1`（`0x104fe20`）调用 `net/http.(*Server).ListenAndServeTLS`（`0x708c20`）；`main.main.func1.2`（`0x104fe60`）调用 `net/http.(*Server).ListenAndServe`（`0x707240`）。这两个分支是协议选择，并非同时接收两种协议。
- `main.main.func1.3`（`0x1044dc0`）先对旧服务调用 `Shutdown`，再启动选中的 HTTP/HTTPS 服务。所以启用 HTTPS 后仍只有一个监听端口。
- `main.buildFullEngine`（`0x1045120`）在 `0x10453b6` 调用 `auth-api/internal/install.(*Migrator).Run`（`0x103d300`），之后才加载系统设置等组件。因此原版正常启动会自动执行数据库迁移。
- `auth-api/internal/auth.(*SessionStore).Create`（`0x8c5ce0`）使用 `sess:admin:` 前缀构造管理员会话索引，读取并删除旧令牌的会话记录，然后在 `0x8c61f2` 调用 Redis `SET` 保存新令牌字符串，因此同一管理员的新登录会使旧登录失效。此前重写版对同名键执行 `SADD`，遇到保留的原版索引会报 `WRONGTYPE`，表现为“签发会话失败”。当前版按用户要求恢复单登录，兼容原版字符串和上一版集合索引，原子保存新会话并撤销全部旧会话；退出和管理员会话撤销也兼容两种格式。

本次重写版对应修复：

1. 使用同一监听地址切换 HTTP/HTTPS，忽略旧字段 `https_addr` 对监听端口的影响。
2. 保留 `-port` 绑定所有网卡的行为。按照用户要求，两个地址参数同时提供时以 `-port` 为最高优先级；原版反汇编显示 `-addr` 在 `-port` 之后覆盖，这是本次明确保留的用户指定差异。
3. 优先查找程序旁的配置，并按配置位置确定运行目录，避免从 `/root` 启动时读取错误配置或主密钥。这是对部署兼容性的修复，不宣称原版已经提供该搜索顺序。
4. 启动时同步报告证书和端口绑定错误；运行中证书校验失败保留现有监听，证书轮换不重开监听端口。
5. 已安装实例在创建业务处理器、启动后台任务及监听端口之前自动应用待执行迁移。`-migrate` 仅保留为执行后退出的维护命令，正常启动不需要它。迁移失败直接退出；不自动重放有中断记录的迁移。

验证覆盖端口优先级、配置路径、HTTP→HTTPS→HTTP 切换、切换请求完成、证书轮换、旧 `https_addr` 被占用、缺失主密钥、无效证书及端口绑定失败。测试只使用临时配置、生成的证书和隔离的数据库。

Linux 实际运行对照也已通过：原版和新版本连接同一临时 MySQL/Redis，使用同一套证书和主密钥。当数据库中的 `https_addr` 为 `:8443`、`-port` 为另一端口时，两者均只监听 `-port` 指定的端口；`https_enabled` 决定提供 HTTP 还是 HTTPS。新产物还通过了跨目录启动、安装、47 份迁移、内嵌前端、协议热切换、重启、缺失主密钥、端口占用，以及数据库生命周期和 mTLS 回归测试。运行下方验证命令后会生成本地报告 `dist/STARTUP-VERIFICATION.json`；产物和测试报告不提交到源码仓库。

自动迁移也已在临时数据库中验证：原版启动会补上待执行的 0041；新版正常使用 `-port` 启动即可从 0041 升至 0047，保留已有管理员、用户、卡密、已配置的地域服务、HTTPS 设置和主密钥。再次启动不重跑迁移，不重置后续更新的活跃时间。SQL 执行失败时新版在监听端口之前退出，后续启动不会自动重放中断的迁移。`installed: false` 的初次启动不会提前创建数据库表，安装向导仍正常工作。

会话兼容性验证先运行原版完成管理员登录，确认 Redis 索引为字符串，然后保留该 Redis 切换到新版。新版登录、鉴权读取管理员信息和退出均通过，新登录会删除旧令牌记录，旧令牌请求返回业务码 401。还模拟上一版集合索引，确认新登录恢复单个字符串索引并撤销集合内的全部旧令牌。并发登录只保留一个会话，失败登录不会挤掉当前会话，旧会话退出或更新不会复活旧令牌或误删新会话。会话和索引同步续期，兼容缩短有效期和绝对过期；不同管理员及 SDK 用户会话互不影响。

本版内嵌前端还包含自动生成的 `build-info.json`，页脚读取同一次构建的版本与时间。Linux 验证会对照实际提供的 JavaScript 和构建元数据，并确认重启后时间保持不变。

可在 Linux/WSL 中运行 `python3 scripts/verify-startup-linux.py --binary dist/auth-api-linux-amd64 --original /path/to/original --report dist/STARTUP-VERIFICATION.json` 复现对照测试，将 `--binary` 替换为实际构建产物路径。脚本使用 Docker 创建并清理专用临时数据库和 Redis，不读取或连接生产配置；`--original` 可省略。
