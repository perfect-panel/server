# PPanel Configuration Guide

This document provides a comprehensive guide to the configuration file for the PPanel application. The configuration
file is in YAML format and defines settings for the server, logging, database, Redis, and admin access.

## 1. Configuration File Overview

- **Default Path**: `./etc/ppanel.yaml`
- **Custom Path**: Specify a custom path using the `--config` startup parameter.
- **Format**: YAML, supports comments, and must be named with a `.yaml` extension.

## 2. Configuration File Structure

Below is an example of the configuration file with default values and explanations:

```yaml
# PPanel Configuration
Host: "0.0.0.0"                     # Server listening address
Port: 8080                          # Server listening port
Debug: false                        # Enable debug mode (disables background logging)
JwtAuth: # JWT authentication settings
  AccessSecret: ""                  # Access token secret (randomly generated if empty)
  AccessExpire: 604800              # Access token expiration (seconds)
Logger: # Logging configuration
  ServiceName: ""                   # Service name for log identification
  Mode: "console"                   # Log output mode (console, file, volume)
  Encoding: "json"                  # Log format (json, plain)
  TimeFormat: "2006-01-02T15:04:05.000Z07:00"  # Custom time format
  Path: "logs"                      # Log file directory
  Level: "info"                     # Log level (info, error, severe)
  Compress: false                   # Enable log compression
  KeepDays: 7                       # Log retention period (days)
  StackCooldownMillis: 100          # Stack trace cooldown (milliseconds)
  MaxBackups: 3                     # Maximum number of log backups
  MaxSize: 50                       # Maximum log file size (MB)
  Rotation: "daily"                 # Log rotation strategy (daily, size)
MySQL: # MySQL database configuration
  Addr: ""                          # MySQL address (required)
  Username: ""                      # MySQL username (required)
  Password: ""                      # MySQL password (required)
  Dbname: ""                        # MySQL database name (required)
  Config: "charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"  # MySQL connection parameters
  MaxIdleConns: 10                  # Maximum idle connections
  MaxOpenConns: 100                 # Maximum open connections
  LogMode: "info"                   # Log level (debug, error, warn, info)
  LogZap: true                      # Enable Zap logging for SQL
  SlowThreshold: 1000               # Slow query threshold (milliseconds)
Redis: # Redis configuration
  Host: "localhost:6379"            # Redis address
  Pass: ""                          # Redis password
  DB: 0                             # Redis database index
Administer: # Admin login configuration
  Email: "admin@ppanel.dev"         # Admin login email
  Password: "password"              # Admin login password
```

## 3. Configuration Details

### 3.1 Server Settings

- **`Host`**: Address the server listens on.
  - Default: `0.0.0.0` (all network interfaces).
- **`Port`**: Port the server listens on.
  - Default: `8080`.
- **`Debug`**: Enables debug mode, disabling background logging.
  - Default: `false`.

### 3.2 JWT Authentication (`JwtAuth`)

- **`AccessSecret`**: Secret key for access tokens.
  - Default: Randomly generated if not specified.
- **`AccessExpire`**: Token expiration time in seconds.
  - Default: `604800` (7 days).

### 3.3 Logging (`Logger`)

- **`ServiceName`**: Identifier for logs, used as the log filename in `volume` mode.
  - Default: `""`.
- **`Mode`**: Log output destination.
  - Options: `console` (stdout/stderr), `file` (to a directory), `volume` (Docker volume).
  - Default: `console`.
- **`Encoding`**: Log format.
  - Options: `json` (structured JSON), `plain` (plain text with colors).
  - Default: `json`.
- **`TimeFormat`**: Custom time format for logs.
  - Default: `2006-01-02T15:04:05.000Z07:00`.
- **`Path`**: Directory for log files.
  - Default: `logs`.
- **`Level`**: Log filtering level.
  - Options: `info` (all logs), `error` (error and severe), `severe` (severe only).
  - Default: `info`.
- **`Compress`**: Enable compression for log files (only in `file` mode).
  - Default: `false`.
- **`KeepDays`**: Retention period for log files (in days).
  - Default: `7`.
- **`StackCooldownMillis`**: Cooldown for stack trace logging to prevent log flooding.
  - Default: `100`.
- **`MaxBackups`**: Maximum number of log backups (for `size` rotation).
  - Default: `3`.
- **`MaxSize`**: Maximum log file size in MB (for `size` rotation).
  - Default: `50`.
- **`Rotation`**: Log rotation strategy.
  - Options: `daily` (rotate daily), `size` (rotate by size).
  - Default: `daily`.

### 3.4 MySQL Database (`MySQL`)

- **`Addr`**: MySQL server address.
  - Required.
- **`Username`**: MySQL username.
  - Required.
- **`Password`**: MySQL password.
  - Required.
- **`Dbname`**: MySQL database name.
  - Required.
- **`Config`**: MySQL connection parameters.
  - Default: `charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai`.
- **`MaxIdleConns`**: Maximum idle connections.
  - Default: `10`.
- **`MaxOpenConns`**: Maximum open connections.
  - Default: `100`.
- **`LogMode`**: SQL logging level.
  - Options: `debug`, `error`, `warn`, `info`.
  - Default: `info`.
- **`LogZap`**: Enable Zap logging for SQL queries.
  - Default: `true`.
- **`SlowThreshold`**: Threshold for slow query logging (in milliseconds).
  - Default: `1000`.

### 3.5 Redis (`Redis`)

- **`Host`**: Redis server address.
  - Default: `localhost:6379`.
- **`Pass`**: Redis password.
  - Default: `""` (no password).
- **`DB`**: Redis database index.
  - Default: `0`.

### 3.6 Admin Login (`Administer`)

- **`Email`**: Admin login email.
  - Default: `admin@ppanel.dev`.
- **`Password`**: Admin login password.
  - Default: `password`.

## 4. Environment Variables

The following environment variables can be used to override configuration settings:

| Environment Variable | Configuration Section | Example Value                                |
|----------------------|-----------------------|----------------------------------------------|
| `PPANEL_DB`          | MySQL                 | `root:password@tcp(localhost:3306)/vpnboard` |
| `PPANEL_REDIS`       | Redis                 | `redis://localhost:6379`                     |

## 5. Best Practices

- **Security**: Avoid using default `Administer` credentials in production. Update `Email` and `Password` to secure
  values.
- **Logging**: Use `file` or `volume` mode for production to persist logs. Adjust `Level` to `error` or `severe` to
  reduce log volume.
- **Database**: Ensure `MySQL` and `Redis` credentials are secure and not exposed in version control.
- **JWT**: Specify a strong `AccessSecret` for `JwtAuth` to enhance security.

For further assistance, refer to the official PPanel documentation or contact support.