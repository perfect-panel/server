# mysql2postgres

用于将现有 PPanel MySQL 部署一次性迁移到 PostgreSQL 的数据迁移工具。

目标 PostgreSQL 数据库必须先使用相同版本的 PPanel server 完成初始化。该工具只复制数据，不创建或迁移目标数据库结构。

## 迁移流程

1. 停止 PPanel server、queue、scheduler，以及可能写入流量数据的节点。
2. 备份 MySQL 数据库。
3. 创建一个空的 PostgreSQL 数据库。
4. 使用 PostgreSQL 配置启动一次最新版 PPanel server，让内置 PostgreSQL migrations 创建数据库结构，然后停止服务。
5. 使用 `--truncate --yes` 运行本工具。
6. 将 `etc/ppanel.yaml` 指向 PostgreSQL，然后启动 PPanel。

## 使用方式

推荐使用主程序内置命令：

```bash
./ppanel migrate mysql2postgres \
  --mysql 'user:pass@tcp(127.0.0.1:3306)/ppanel?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai' \
  --postgres 'postgres://ppanel:pass@127.0.0.1:5432/ppanel?sslmode=disable' \
  --truncate \
  --yes
```

开发调试用 wrapper：

```bash
go run ./tools/mysql2postgres \
  --mysql 'user:pass@tcp(127.0.0.1:3306)/ppanel?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai' \
  --postgres 'postgres://ppanel:pass@127.0.0.1:5432/ppanel?sslmode=disable' \
  --truncate \
  --yes
```

建议先 dry-run 查看迁移计划：

```bash
go run ./tools/mysql2postgres \
  --mysql 'user:pass@tcp(127.0.0.1:3306)/ppanel?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai' \
  --postgres 'postgres://ppanel:pass@127.0.0.1:5432/ppanel?sslmode=disable' \
  --dry-run
```

## 参数

- `--mysql`：源 MySQL DSN。
- `--postgres`：目标 PostgreSQL DSN。
- `--schema`：PostgreSQL schema，默认是 `public`。
- `--truncate`：复制前清空目标库中两边都存在的表。
- `--yes`：使用 `--truncate` 时必须传入，用于确认破坏性操作。
- `--tables`：逗号分隔的表白名单。
- `--exclude`：逗号分隔的表黑名单。
- `--batch-size`：进度日志输出间隔。
- `--dry-run`：只打印迁移计划，不复制数据。

## 注意事项

- `schema_migrations` 永远不会被复制。
- 只复制源库和目标库同时存在的表与列。PostgreSQL 新增列会保留默认值。
- 表会按照 PostgreSQL 外键依赖顺序复制；如果 MySQL 表存在主键，读取时会按主键排序。
- 复制完成后会重置 PostgreSQL sequences。
- 建议使用全新的 PostgreSQL 数据库。如果不使用 `--truncate`，可能因为 migrations 已插入默认数据而触发唯一约束错误。
