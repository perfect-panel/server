# mysql2postgres

One-shot data migration tool for moving an existing PPanel MySQL deployment to
PostgreSQL.

The target PostgreSQL database must already be initialized by the same PPanel
server version. This tool copies data only; it does not create or migrate the
target schema.

## Workflow

1. Stop PPanel server, queue, scheduler, and nodes that can write traffic data.
2. Back up MySQL.
3. Create an empty PostgreSQL database.
4. Start the latest PPanel server once with PostgreSQL config so built-in
   PostgreSQL migrations create the schema, then stop it.
5. Run this tool with `--truncate --yes`.
6. Point `etc/ppanel.yaml` to PostgreSQL and start PPanel.

## Usage

```bash
go run ./tools/mysql2postgres \
  --mysql 'user:pass@tcp(127.0.0.1:3306)/ppanel?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai' \
  --postgres 'postgres://ppanel:pass@127.0.0.1:5432/ppanel?sslmode=disable' \
  --truncate \
  --yes
```

Dry-run the plan first:

```bash
go run ./tools/mysql2postgres \
  --mysql 'user:pass@tcp(127.0.0.1:3306)/ppanel?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai' \
  --postgres 'postgres://ppanel:pass@127.0.0.1:5432/ppanel?sslmode=disable' \
  --dry-run
```

## Flags

- `--mysql`: source MySQL DSN.
- `--postgres`: target PostgreSQL DSN.
- `--schema`: PostgreSQL schema, defaults to `public`.
- `--truncate`: truncate common target tables before copy.
- `--yes`: required when `--truncate` is used.
- `--tables`: comma-separated allowlist.
- `--exclude`: comma-separated denylist.
- `--batch-size`: progress log interval.
- `--dry-run`: print plan without copying.

## Notes

- `schema_migrations` is never copied.
- Only tables and columns that exist on both sides are copied. New PostgreSQL
  columns keep their defaults.
- Tables are copied in PostgreSQL foreign-key dependency order, and MySQL rows
  are read by primary-key order when a primary key exists.
- PostgreSQL sequences are reset after copy.
- Use a fresh PostgreSQL database. Running without `--truncate` can fail on
  unique constraints because migrations insert default rows.
