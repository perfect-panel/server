# PPanel 配置指南

本文件为 PPanel 应用程序的配置文件提供全面指南。配置文件采用 YAML 格式，定义了服务器、日志、数据库、Redis 和管理员访问的相关设置。

## 1. 配置文件概述

- **默认路径**：`./etc/ppanel.yaml`
- **自定义路径**：通过启动参数 `--config` 指定配置文件路径。
- **格式**：YAML 格式，支持注释，文件名需以 `.yaml` 结尾。

## 2. 配置文件结构

以下是配置文件示例，包含默认值和说明：

```yaml
# PPanel 配置文件
Host: "0.0.0.0"                     # 服务监听地址
Port: 8080                          # 服务监听端口
Debug: false                        # 是否开启调试模式（禁用后台日志）
JwtAuth: # JWT 认证配置
  AccessSecret: ""                  # 访问令牌密钥（为空时随机生成）
  AccessExpire: 604800              # 访问令牌过期时间（秒）
Logger: # 日志配置
  ServiceName: ""                   # 日志服务标识名称
  Mode: "console"                   # 日志输出模式（console、file、volume）
  Encoding: "json"                  # 日志格式（json、plain）
  TimeFormat: "2006-01-02T15:04:05.000Z07:00"  # 自定义时间格式
  Path: "logs"                      # 日志文件目录
  Level: "info"                     # 日志级别（info、error、severe）
  Compress: false                   # 是否压缩日志文件
  KeepDays: 7                       # 日志保留天数
  StackCooldownMillis: 100          # 堆栈日志冷却时间（毫秒）
  MaxBackups: 3                     # 最大日志备份数
  MaxSize: 50                       # 最大日志文件大小（MB）
  Rotation: "daily"                 # 日志轮转策略（daily、size）
MySQL: # MySQL 数据库配置
  Addr: ""                          # MySQL 地址（必填）
  Username: ""                      # MySQL 用户名（必填）
  Password: ""                      # MySQL 密码（必填）
  Dbname: ""                        # MySQL 数据库名（必填）
  Config: "charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"  # MySQL 连接参数
  MaxIdleConns: 10                  # 最大空闲连接数
  MaxOpenConns: 100                 # 最大打开连接数
  LogMode: "info"                   # 日志级别（debug、error、warn、info）
  LogZap: true                      # 是否使用 Zap 记录 SQL 日志
  SlowThreshold: 1000               # 慢查询阈值（毫秒）
Redis: # Redis 配置
  Host: "localhost:6379"            # Redis 地址
  Pass: ""                          # Redis 密码
  DB: 0                             # Redis 数据库索引
Administer: # 管理员登录配置
  Email: "admin@ppanel.dev"         # 管理员登录邮箱
  Password: "password"              # 管理员登录密码
```

## 3. 配置项说明

### 3.1 服务器设置

- **`Host`**：服务监听的地址。
  - 默认：`0.0.0.0`（监听所有网络接口）。
- **`Port`**：服务监听的端口。
  - 默认：`8080`。
- **`Debug`**：是否开启调试模式，开启后禁用后台日志功能。
  - 默认：`false`。

### 3.2 JWT 认证 (`JwtAuth`)

- **`AccessSecret`**：访问令牌的密钥。
  - 默认：为空时随机生成。
- **`AccessExpire`**：令牌过期时间（秒）。
  - 默认：`604800`（7天）。

### 3.3 日志配置 (`Logger`)

- **`ServiceName`**：日志的服务标识名称，在 `volume` 模式下用作日志文件名。
  - 默认：`""`。
- **`Mode`**：日志输出方式。
  - 选项：`console`（标准输出/错误输出）、`file`（写入指定目录）、`volume`（Docker 卷）。
  - 默认：`console`。
- **`Encoding`**：日志格式。
  - 选项：`json`（结构化 JSON）、`plain`（纯文本，带颜色）。
  - 默认：`json`。
- **`TimeFormat`**：日志时间格式。
  - 默认：`2006-01-02T15:04:05.000Z07:00`。
- **`Path`**：日志文件存储目录。
  - 默认：`logs`。
- **`Level`**：日志过滤级别。
  - 选项：`info`（记录所有日志）、`error`（仅错误和严重日志）、`severe`（仅严重日志）。
  - 默认：`info`。
- **`Compress`**：是否压缩日志文件（仅在 `file` 模式下生效）。
  - 默认：`false`。
- **`KeepDays`**：日志文件保留天数。
  - 默认：`7`。
- **`StackCooldownMillis`**：堆栈日志冷却时间（毫秒），防止日志过多。
  - 默认：`100`。
- **`MaxBackups`**：最大日志备份数量（仅在 `size` 轮转时生效）。
  - 默认：`3`。
- **`MaxSize`**：日志文件最大大小（MB，仅在 `size` 轮转时生效）。
  - 默认：`50`。
- **`Rotation`**：日志轮转策略。
  - 选项：`daily`（按天轮转）、`size`（按大小轮转）。
  - 默认：`daily`。

### 3.4 MySQL 数据库 (`MySQL`)

- **`Addr`**：MySQL 服务器地址。
  - 必填。
- **`Username`**：MySQL 用户名。
  - 必填。
- **`Password`**：MySQL 密码。
  - 必填。
- **`Dbname`**：MySQL 数据库名。
  - 必填。
- **`Config`**：MySQL 连接参数。
  - 默认：`charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai`。
- **`MaxIdleConns`**：最大空闲连接数。
  - 默认：`10`。
- **`MaxOpenConns`**：最大打开连接数。
  - 默认：`100`。
- **`LogMode`**：SQL 日志级别。
  - 选项：`debug`、`error`、`warn`、`info`。
  - 默认：`info`。
- **`LogZap`**：是否使用 Zap 记录 SQL 日志。
  - 默认：`true`。
- **`SlowThreshold`**：慢查询阈值（毫秒）。
  - 默认：`1000`。

### 3.5 Redis 配置 (`Redis`)

- **`Host`**：Redis 服务器地址。
  - 默认：`localhost:6379`。
- **`Pass`**：Redis 密码。
  - 默认：`""`（无密码）。
- **`DB`**：Redis 数据库索引。
  - 默认：`0`。

### 3.6 管理员登录 (`Administer`)

- **`Email`**：管理员登录邮箱。
  - 默认：`admin@ppanel.dev`。
- **`Password`**：管理员登录密码。
  - 默认：`password`。

## 4. 环境变量

以下环境变量可用于覆盖配置文件中的设置：

| 环境变量           | 配置项      | 示例值                                          |
|----------------|----------|----------------------------------------------|
| `PPANEL_DB`    | MySQL 配置 | `root:password@tcp(localhost:3306)/vpnboard` |
| `PPANEL_REDIS` | Redis 配置 | `redis://localhost:6379`                     |

## 5. 最佳实践

- **安全性**：生产环境中避免使用默认的 `Administer` 凭据，更新 `Email` 和 `Password` 为安全值。
- **日志**：生产环境中建议使用 `file` 或 `volume` 模式持久化日志，将 `Level` 设置为 `error` 或 `severe` 以减少日志量。
- **数据库**：确保 `MySQL` 和 `Redis` 凭据安全，避免在版本控制中暴露。
- **JWT**：为 `JwtAuth` 的 `AccessSecret` 设置强密钥以增强安全性。

如需进一步帮助，请参考 PPanel 官方文档或联系支持团队。