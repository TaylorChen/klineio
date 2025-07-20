# klineio

## Project Overview

This is a cryptocurrency price monitoring and notification system developed in Go. It automatically fetches real-time price data for selected popular cryptocurrencies from major exchanges like Binance and OKEX. It then compares the current price against a historical average (e.g., last 30 days). If the current price drops below a predefined threshold compared to the average, the system sends real-time alert notifications via DingTalk. All fetched price data is persistently stored in a database, ensuring only one record per cryptocurrency, per exchange, per day.

## Key Features

*   **Multi-Exchange Support**: Fetches market data from Binance and OKEX.
*   **Top Coin Monitoring**: Automatically retrieves and monitors the top N cryptocurrencies by trading volume (currently configured for the top 50).
*   **Price Drop Alert**: Real-time calculation of price drops against a 30-day average, triggering alerts if a set threshold (e.g., 20%) is exceeded.
*   **DingTalk Notifications**: Sends Markdown-formatted price drop alerts via DingTalk custom bots.
*   **Data Persistence**: Stores daily cryptocurrency price data in MySQL (or other GORM-supported databases), ensuring only one latest price record per cryptocurrency, per exchange, per day.
*   **Proxy Support**: Supports API calls through HTTP proxies to handle network restrictions.
*   **Graceful Shutdown**: Implements graceful shutdown using OS signals (e.g., `SIGINT`, `SIGTERM`).

## Technologies Used

*   **Go**: The primary programming language for the project.
*   **GORM**: A powerful ORM library for Go, used for database operations.
*   **Google Wire**: Dependency injection tool for managing component dependencies.
*   **gocron**: A Go library for scheduling periodic tasks.
*   **Zap Logger**: A structured logging library for efficient logging.
*   **Viper**: Used for configuration management.
*   **Gin (Optional, not enabled by default)**: A web framework for building HTTP APIs (not directly used by the current task application).
*   **MySQL**: Recommended database (PostgreSQL, SQLite, etc., are also supported).

## Environment Setup

1.  **Go Environment**: Ensure Go version 1.17+ is installed.
2.  **MySQL Database**: Set up a MySQL database instance and create a database for storing data.
3.  **DingTalk Custom Bot**:
    *   Add a custom bot to your DingTalk group.
    *   Obtain the bot's Webhook URL.
    *   Configure `dingtalk.webhook_url` in `config/local.yml` (or `prod.yml`).
4.  **Exchange API Keys (Optional but Recommended)**:
    *   While fetching top volume lists and K-line data usually doesn't require API Keys, if you need advanced authenticated operations, it's recommended to configure `exchange.binance.api_key` and `exchange.binance.secret_key` in `config/local.yml`.
    *   Similarly, configure `exchange.okex.api_key` and `exchange.okex.secret_key` for OKEX.
5.  **HTTP Proxy (Optional)**: If your network environment requires an HTTP proxy to access exchange APIs, configure `proxy.http` in `config/local.yml`, e.g., `http://127.0.0.1:7890`.

## Configuration

The project is primarily configured via `config/local.yml` or `config/prod.yml`. Here are the key configuration items:

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
  log_level: debug              # Log level: debug, info, warn, error, fatal
  mode: both                    # Log output mode: file, console, both
  encoding: console             # Log encoding format: json, console
  log_file_name: "./storage/logs/server.log"
  max_backups: 30
  max_age: 7
  max_size: 1024
  compress: true

dingtalk:
  webhook_url: "YOUR_DINGTALK_WEBHOOK_URL" # Replace with your DingTalk bot Webhook URL

exchange:
  binance:
    api_key: "YOUR_BINANCE_API_KEY"
    secret_key: "YOUR_BINANCE_SECRET_KEY"
  okex:
    api_key: "YOUR_OKEX_API_KEY"
    secret_key: "YOUR_OKEX_SECRET_KEY"

price_monitor:
  default_threshold: 0.20       # Default price drop threshold, e.20 for 20%
  top_n_symbols: 50             # Number of top volume symbols to monitor
  timeout_seconds: 120          # Overall timeout for price monitoring tasks (in seconds)
  api_request_delay_ms: 1000    # Delay between each API request for a symbol (in milliseconds), to avoid rate limits

proxy:
  http: "http://127.0.0.1:7890" # HTTP proxy address, leave empty or comment out if not needed
```

## How to Run

1.  **Install Dependencies**:
    ```bash
    go mod tidy
    ```

2.  **Generate Wire Dependency Injection Code**:
    ```bash
    go generate ./...
    ```

3.  **Database Migration**:
    For the first run or after modifying database models, you need to execute database migration to create or update table structures. Ensure your database connection details in `config/local.yml` are correct.

    ```bash
    go run cmd/migration/main.go -conf config/local.yml
    ```
    After running this command, you will see the SQL statements generated by GORM in the terminal, which helps confirm schema changes.

4.  **Start the Task Application**:
    ```bash
    go run cmd/task/main.go -conf config/local.yml
    ```
    The application will start a scheduled task runner that periodically performs price monitoring.

## Project Structure

```
notify/
├── api/                       # API definitions (typically for HTTP APIs, not directly used by current task app)
├── cmd/                       # Application entry points
│   ├── migration/             # Database migration tool
│   ├── server/                # HTTP server startup (not directly used by current task app)
│   └── task/                  # Scheduled task application startup (our main program)
├── config/                    # Configuration files
├── deploy/                    # Deployment-related configurations (e.g., Docker Compose)
├── docs/                      # Swagger documentation
├── internal/                  # Internal implementations (core business logic)
│   ├── handler/               # HTTP request handlers (not directly used by current task app)
│   ├── job/                   # Implementations of scheduled jobs (e.g., PriceMonitorJob)
│   ├── middleware/            # Middleware
│   ├── model/                 # Database model definitions (e.g., ExchangePrice, MonitorConfig)
│   ├── repository/            # Database operation interfaces and implementations
│   ├── server/                # Server logic (e.g., TaskServer for task scheduling)
│   ├── service/               # Business logic services (e.g., PriceMonitorService)
│   └── task/                  # Task definitions (e.g., UserTask, PriceMonitorTask)
├── pkg/                       # Public libraries and common utilities
│   ├── app/                   # Application lifecycle management
│   ├── config/                # Configuration loading
│   ├── jwt/                   # JWT utilities
│   ├── log/                   # Custom logger wrapper
│   ├── notifier/              # Notifier interfaces and implementations (e.g., DingTalk)
│   ├── exchange/              # Exchange API client (e.g., Binance, OKEX)
│   ├── server/                # Generic server components
│   ├── sid/                   # ID generator
│   └── zapgorm2/              # Zap logger adapter for GORM
└── storage/                   # Storage directory (e.g., SQLite database file, logs)
```

## Troubleshooting Common Issues

*   **`context deadline exceeded`**:
    *   **Cause**: API calls or database operations take too long, exceeding the allocated context timeout.
    *   **Solution**: Check and increase the `price_monitor.timeout_seconds` value in `config/local.yml`. Also, ensure that the `Timeout` setting for HTTP clients in `pkg/exchange` is sufficiently large.

*   **`429 Too Many Requests`**:
    *   **Cause**: Sending requests to the exchange API too frequently, triggering rate limits.
    *   **Solution**: Increase the `price_monitor.api_request_delay_ms` value in `config/local.yml` to introduce a proper delay between each API call.

*   **Duplicate Data in Database**:
    *   **Cause**: Database migration was not executed correctly, or the `Upsert` logic is not working as expected.
    *   **Solution**: Ensure that you have successfully run `go run cmd/migration/main.go -conf config/local.yml` and that the `date` field and `idx_symbol_exchange_date` unique index in the `exchange_prices` table are created. Confirm that the `UpsertExchangePrice` logic in `internal/repository/crypto.go` correctly uses the `Date` field for conflict resolution.

*   **`year is not in the range [1, 9999]`**:
    *   **Cause**: Timestamp conversion error, where a millisecond timestamp was incorrectly interpreted as seconds, leading to a year far beyond the database's supported range.
    *   **Solution**: Ensure that `time.UnixMilli(price.Timestamp).UTC()` within the `UpsertExchangePrice` method in `internal/repository/crypto.go` correctly converts the millisecond timestamp to `time.Time`.

*   **Excessive `record not found` errors in logs**:
    *   **Cause**: GORM by default logs `gorm.ErrRecordNotFound` at the `error` level.
    *   **Solution**: We have addressed this by customizing `pkg/zapgorm2/custom_logger.go`. `record not found` messages should now be logged at the `debug` level. If they still appear as errors, please check your `log.log_level` configuration.