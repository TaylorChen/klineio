# klineio

## 项目简介

这是一个基于 Go 语言开发的加密货币价格监控与通知系统，名为 `klineio`。它能够自动从 Binance 和 OKEX 等主流交易所获取指定热门币种的实时价格数据，与过去一定时间（例如 30 天）的平均价格进行对比。如果当前价格相较于平均价格下跌超过预设阈值，系统将通过钉钉（DingTalk）发送实时警报通知。所有获取到的价格数据都会持久化存储到数据库中，并确保每天每个交易所每个币种只有一条记录。

## 核心功能

*   **多交易所支持**: 支持从 Binance 和 OKEX 获取市场数据。
*   **热门币种监控**: 自动抓取交易所交易量排名前 N 的币种进行监控（目前配置为前 50）。
*   **价格下跌预警**: 实时计算当前价格与过去 30 天平均价格的跌幅，若超过设定阈值（例如 20%），则触发警报。
*   **钉钉通知**: 通过钉钉自定义机器人发送 Markdown 格式的价格下跌警报通知。
*   **数据持久化**: 将每日币种价格数据存储到 MySQL (或其他 GORM 支持的数据库) 中，确保每日每个币种每个交易所只有一条最新的价格记录。
*   **代理支持**: 支持通过 HTTP 代理进行 API 调用，以应对网络限制。
*   **优雅停机**: 支持通过操作系统信号 (如 `SIGINT`, `SIGTERM`) 实现程序的优雅关闭。

## 技术栈

*   **Go 语言**: 项目主要开发语言。
*   **GORM**: 强大的 Go ORM 库，用于数据库操作。
*   **Google Wire**: 依赖注入工具，管理项目组件依赖关系。
*   **gocron**: Go 语言的定时任务调度库。
*   **Zap Logger**: 结构化日志库，提供高效的日志记录。
*   **Viper**: 用于配置管理。
*   **Gin (可选，未启用但项目结构支持)**: 用于构建 Web API (当前任务应用程序未使用)。
*   **MySQL**: 推荐的数据库（也支持 PostgreSQL, SQLite 等）。

## 环境准备

1.  **Go 环境**: 确保你已经安装 Go 1.17+ 版本。
2.  **MySQL 数据库**: 准备一个 MySQL 数据库实例，并创建用于存储数据的数据库。
3.  **钉钉自定义机器人**:
    *   在钉钉群中添加自定义机器人。
    *   获取机器人的 Webhook URL。
    *   在 `config/local.yml` (或 `prod.yml`) 中配置 `dingtalk.webhook_url`。
4.  **交易所 API Keys (可选但推荐)**:
    *   虽然获取热门币种列表和 K 线数据通常不需要 API Keys，但如果你需要进行更高级的认证操作，建议在 `config/local.yml` 中配置 `exchange.binance.api_key` 和 `exchange.binance.secret_key`。
    *   OKEX 也类似配置 `exchange.okex.api_key` 和 `exchange.okex.secret_key`。
5.  **HTTP 代理 (可选)**: 如果你的网络环境需要通过代理访问交易所 API，请在 `config/local.yml` 中配置 `proxy.http`，例如 `http://127.0.0.1:7890`。

## 配置说明

项目主要通过 `config/local.yml` 或 `config/prod.yml` 进行配置。以下是关键配置项的说明：

```yaml
# config/local.yml
env: local
http:
  host: 127.0.0.1
  port: 8000
security:
  api_sign:
    app_key: 123456
    app_security: 123456
  jwt:
    key: QQYnRFerJTSEcrfB89fw8prOaObmrch8
data:
  db:
    user:
      driver: mysql
      dsn: root:@tcp(127.0.0.1:3306)/notify?charset=utf8mb4&parseTime=True&loc=Local
  redis:
    addr: 127.0.0.1:6350
    password: ""
    db: 0
    read_timeout: 0.2s
    write_timeout: 0.2s
  mongo:
    uri: mongodb://root:123456@localhost:27017
log:
  log_level: debug              # 日志级别：debug, info, warn, error, fatal
  mode: both                    # 日志输出模式：file, console, both
  encoding: console             # 日志编码格式：json, console
  log_file_name: "./storage/logs/server.log"
  max_backups: 30
  max_age: 7
  max_size: 1024
  compress: true

dingtalk:
  webhook_url: "YOUR_DINGTALK_WEBHOOK_URL" # 替换为你的钉钉机器人 Webhook URL

exchange:
  binance:
    api_key: "YOUR_BINANCE_API_KEY"
    secret_key: "YOUR_BINANCE_SECRET_KEY"
  okex:
    api_key: "YOUR_OKEX_API_KEY"
    secret_key: "YOUR_OKEX_SECRET_KEY"

price_monitor:
  default_threshold: 0.20       # 默认价格下跌阈值，例如 0.20 表示 20%
  top_n_symbols: 50             # 监控交易量前 N 的币种
  timeout_seconds: 120          # 价格监控任务的整体超时时间（秒）
  api_request_delay_ms: 1000    # 每个 API 请求之间的延迟（毫秒），用于避免速率限制

proxy:
  http: "http://127.0.0.1:7890" # HTTP 代理地址，如果不需要请留空或注释
```

## 运行项目

1.  **安装依赖**:
    ```bash
    go mod tidy
    ```

2.  **生成 Wire 依赖注入代码**:
    ```bash
    go generate ./...
    ```

3.  **数据库迁移**:
    首次运行或修改数据库模型后，需要执行数据库迁移来创建或更新表结构。请确保你的 `config/local.yml` 中的数据库连接信息正确。

    ```bash
    go run cmd/migration/main.go -conf config/local.yml
    ```
    运行此命令后，你会在终端看到 GORM 生成的 SQL 语句，这有助于确认表结构变更。

4.  **启动任务应用程序**:
    ```bash
    go run cmd/task/main.go -conf config/local.yml
    ```
    应用程序将启动一个定时任务调度器，定期执行价格监控。

## 项目结构

```
notify/
├── api/                       # API 定义 (通常用于 HTTP API，当前任务应用未直接使用)
├── cmd/                       # 应用入口
│   ├── migration/             # 数据库迁移工具
│   ├── server/                # HTTP 服务启动 (当前任务应用未直接使用)
│   └── task/                  # 定时任务启动入口 (我们主要运行的程序)
├── config/                    # 配置文件
├── deploy/                    # 部署相关配置 (如 Docker Compose)
├── docs/                      # Swagger 文档
├── internal/                  # 内部实现 (核心业务逻辑)
│   ├── handler/               # HTTP 请求处理 (当前任务应用未直接使用)
│   ├── job/                   # 定时任务的实现 (如 PriceMonitorJob)
│   ├── middleware/            # 中间件
│   ├── model/                 # 数据库模型定义 (如 ExchangePrice, MonitorConfig)
│   ├── repository/            # 数据库操作接口及实现
│   ├── server/                # 服务器逻辑 (如 TaskServer 负责任务调度)
│   ├── service/               # 业务逻辑服务 (如 PriceMonitorService)
│   └── task/                  # 任务定义 (如 UserTask, PriceMonitorTask)
├── pkg/                       # 公共库和通用工具
│   ├── app/                   # 应用生命周期管理
│   ├── config/                # 配置加载
│   ├── jwt/                   # JWT 工具
│   ├── log/                   # 自定义日志封装
│   ├── notifier/              # 通知器接口及实现 (如 DingTalk)
│   ├── exchange/              # 交易所 API 客户端 (如 Binance, OKEX)
│   ├── server/                # 通用服务组件
│   ├── sid/                   # ID 生成器
│   └── zapgorm2/              # GORM 的 Zap 日志适配器
└── storage/                   # 存储目录 (如 SQLite 数据库文件, 日志)
```

## 常见问题排查

*   **`context deadline exceeded`**:
    *   **原因**: 通常是 API 调用或数据库操作耗时过长，超出了为其分配的上下文超时时间。
    *   **解决方案**: 检查并增加 `config/local.yml` 中 `price_monitor.timeout_seconds` 的值。同时，确保 `pkg/exchange` 中 HTTP 客户端的 `Timeout` 设置也足够大。

*   **`429 Too Many Requests`**:
    *   **原因**: 向交易所 API 发送请求过于频繁，触发了其速率限制。
    *   **解决方案**: 增加 `config/local.yml` 中 `price_monitor.api_request_delay_ms` 的值，在每次 API 调用之间引入适当的延迟。

*   **数据库数据重复**:
    *   **原因**: 数据库迁移未正确执行，或者 `Upsert` 逻辑未生效。
    *   **解决方案**: 确保已经成功运行了 `go run cmd/migration/main.go -conf config/local.yml`，并且 `exchange_prices` 表的 `date` 字段和 `idx_symbol_exchange_date` 唯一索引已创建。确认 `internal/repository/crypto.go` 中的 `UpsertExchangePrice` 逻辑正确使用了 `Date` 字段进行冲突判断。

*   **`year is not in the range [1, 9999]`**:
    *   **原因**: 时间戳转换错误，将毫秒级时间戳错误地解释为秒级，导致生成了超出数据库日期范围的年份。
    *   **解决方案**: 确保 `internal/repository/crypto.go` 中 `UpsertExchangePrice` 方法内部，`time.UnixMilli(price.Timestamp).UTC()` 正确地将毫秒级时间戳转换为 `time.Time`。

*   **日志中出现大量 `record not found` 错误**:
    *   **原因**: GORM 默认将 `gorm.ErrRecordNotFound` 记录为 `error` 级别。
    *   **解决方案**: 我们已经通过自定义 `pkg/zapgorm2/custom_logger.go` 解决了这个问题。现在 `record not found` 应该以 `debug` 级别记录。如果仍然出现，请检查 `log.log_level` 配置。
