package mysql2postgres

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
)

const schemaMigrationsTable = "schema_migrations"

type Config struct {
	MySQLDSN    string
	PostgresDSN string
	Schema      string
	Tables      string
	Exclude     string
	Truncate    bool
	Yes         bool
	DryRun      bool
	BatchSize   int
}

type postgresColumn struct {
	Name      string
	DataType  string
	UDTName   string
	Nullable  bool
	Default   sql.NullString
	Identity  bool
	Generated bool
}

type tablePlan struct {
	Name         string
	Columns      []postgresColumn
	OrderColumns []string
	RowCount     int64
}

type foreignKey struct {
	ChildTable  string
	ParentTable string
}

func DefaultConfig() Config {
	return Config{Schema: "public", BatchSize: 1000}
}

func Run(ctx context.Context, args []string) error {
	cfg, err := ParseFlags(args)
	if err != nil {
		return err
	}
	return Migrate(ctx, cfg)
}

func Migrate(ctx context.Context, cfg Config) error {
	if cfg.MySQLDSN == "" || cfg.PostgresDSN == "" {
		return errors.New("both --mysql and --postgres are required")
	}
	if cfg.BatchSize <= 0 {
		return errors.New("--batch-size must be greater than zero")
	}
	if cfg.Truncate && !cfg.Yes && !cfg.DryRun {
		return errors.New("--truncate is destructive; pass --yes to confirm")
	}

	mysqlDB, err := sql.Open("mysql", normalizeMySQLDSN(cfg.MySQLDSN))
	if err != nil {
		return fmt.Errorf("open mysql: %w", err)
	}
	defer mysqlDB.Close()
	if err := mysqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping mysql: %w", err)
	}

	postgresDB, err := sql.Open("postgres", cfg.PostgresDSN)
	if err != nil {
		return fmt.Errorf("open postgres: %w", err)
	}
	defer postgresDB.Close()
	if err := postgresDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}

	plans, err := buildPlans(ctx, mysqlDB, postgresDB, cfg)
	if err != nil {
		return err
	}
	if len(plans) == 0 {
		return errors.New("no common tables to migrate")
	}

	log.Printf("migration plan: %d table(s)", len(plans))
	for _, plan := range plans {
		log.Printf("  %s: %d row(s), %d common column(s)", plan.Name, plan.RowCount, len(plan.Columns))
	}
	if cfg.DryRun {
		log.Printf("dry run enabled; no data copied")
		return nil
	}

	if cfg.Truncate {
		if err := truncateTables(ctx, postgresDB, cfg.Schema, plans); err != nil {
			return err
		}
	}

	for _, plan := range plans {
		if err := copyTable(ctx, mysqlDB, postgresDB, cfg.Schema, plan, cfg.BatchSize); err != nil {
			return err
		}
	}

	if err := resetSequences(ctx, postgresDB, cfg.Schema); err != nil {
		return err
	}
	log.Printf("migration completed")
	return nil
}

func ParseFlags(args []string) (Config, error) {
	cfg := DefaultConfig()
	fs := flag.NewFlagSet("mysql2postgres", flag.ContinueOnError)
	fs.StringVar(&cfg.MySQLDSN, "mysql", "", "source MySQL DSN")
	fs.StringVar(&cfg.PostgresDSN, "postgres", "", "target PostgreSQL DSN")
	fs.StringVar(&cfg.Schema, "schema", cfg.Schema, "target PostgreSQL schema")
	fs.StringVar(&cfg.Tables, "tables", "", "comma-separated table allowlist")
	fs.StringVar(&cfg.Exclude, "exclude", "", "comma-separated table denylist")
	fs.BoolVar(&cfg.Truncate, "truncate", false, "truncate common target tables before copy")
	fs.BoolVar(&cfg.Yes, "yes", false, "confirm destructive operations")
	fs.BoolVar(&cfg.DryRun, "dry-run", false, "print plan without copying data")
	fs.IntVar(&cfg.BatchSize, "batch-size", cfg.BatchSize, "rows per progress log")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}
	cfg.Schema = strings.TrimSpace(cfg.Schema)
	if cfg.Schema == "" {
		return Config{}, errors.New("--schema cannot be empty")
	}
	return cfg, nil
}

func normalizeMySQLDSN(dsn string) string {
	cfg, err := mysqlDriver.ParseDSN(dsn)
	if err != nil {
		return dsn
	}
	cfg.ParseTime = true
	if cfg.Params == nil {
		cfg.Params = make(map[string]string)
	}
	if _, ok := cfg.Params["charset"]; !ok {
		cfg.Params["charset"] = "utf8mb4"
	}
	return cfg.FormatDSN()
}

func buildPlans(ctx context.Context, mysqlDB, postgresDB *sql.DB, cfg Config) ([]tablePlan, error) {
	sourceTables, err := listMySQLTables(ctx, mysqlDB)
	if err != nil {
		return nil, err
	}
	targetTables, err := listPostgresTables(ctx, postgresDB, cfg.Schema)
	if err != nil {
		return nil, err
	}

	allow := parseTableSet(cfg.Tables)
	exclude := parseTableSet(cfg.Exclude)
	names := make([]string, 0, len(targetTables))
	for name := range targetTables {
		if name == schemaMigrationsTable {
			continue
		}
		if len(allow) > 0 {
			if _, ok := allow[name]; !ok {
				continue
			}
		}
		if _, ok := exclude[name]; ok {
			continue
		}
		if _, ok := sourceTables[name]; !ok {
			log.Printf("skip %s: source table does not exist", name)
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)

	plans := make([]tablePlan, 0, len(names))
	for _, name := range names {
		targetCols, err := listPostgresColumns(ctx, postgresDB, cfg.Schema, name)
		if err != nil {
			return nil, err
		}
		sourceCols, err := listMySQLColumns(ctx, mysqlDB, name)
		if err != nil {
			return nil, err
		}
		commonCols := make([]postgresColumn, 0, len(targetCols))
		for _, col := range targetCols {
			if col.Generated {
				continue
			}
			if _, ok := sourceCols[col.Name]; ok {
				commonCols = append(commonCols, col)
			}
		}
		if len(commonCols) == 0 {
			log.Printf("skip %s: no common columns", name)
			continue
		}
		rowCount, err := countMySQLRows(ctx, mysqlDB, name)
		if err != nil {
			return nil, err
		}
		orderColumns, err := listMySQLPrimaryKeyColumns(ctx, mysqlDB, name)
		if err != nil {
			return nil, err
		}
		plans = append(plans, tablePlan{
			Name:         name,
			Columns:      commonCols,
			OrderColumns: orderColumns,
			RowCount:     rowCount,
		})
	}
	dependencies, err := listPostgresForeignKeys(ctx, postgresDB, cfg.Schema)
	if err != nil {
		return nil, err
	}
	return sortPlansByDependencies(plans, dependencies), nil
}

func parseTableSet(input string) map[string]struct{} {
	result := make(map[string]struct{})
	for _, item := range strings.Split(input, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			result[item] = struct{}{}
		}
	}
	return result
}

func listMySQLTables(ctx context.Context, db *sql.DB) (map[string]struct{}, error) {
	rows, err := db.QueryContext(ctx, `
SELECT table_name
FROM information_schema.tables
WHERE table_schema = DATABASE()
  AND table_type = 'BASE TABLE'
ORDER BY table_name`)
	if err != nil {
		return nil, fmt.Errorf("list mysql tables: %w", err)
	}
	defer rows.Close()
	return scanNameSet(rows)
}

func listPostgresTables(ctx context.Context, db *sql.DB, schema string) (map[string]struct{}, error) {
	rows, err := db.QueryContext(ctx, `
SELECT table_name
FROM information_schema.tables
WHERE table_schema = $1
  AND table_type = 'BASE TABLE'
ORDER BY table_name`, schema)
	if err != nil {
		return nil, fmt.Errorf("list postgres tables: %w", err)
	}
	defer rows.Close()
	return scanNameSet(rows)
}

func scanNameSet(rows *sql.Rows) (map[string]struct{}, error) {
	result := make(map[string]struct{})
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		result[name] = struct{}{}
	}
	return result, rows.Err()
}

func listMySQLColumns(ctx context.Context, db *sql.DB, table string) (map[string]struct{}, error) {
	rows, err := db.QueryContext(ctx, `
SELECT column_name
FROM information_schema.columns
WHERE table_schema = DATABASE()
  AND table_name = ?
ORDER BY ordinal_position`, table)
	if err != nil {
		return nil, fmt.Errorf("list mysql columns for %s: %w", table, err)
	}
	defer rows.Close()
	return scanNameSet(rows)
}

func listMySQLPrimaryKeyColumns(ctx context.Context, db *sql.DB, table string) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
SELECT column_name
FROM information_schema.key_column_usage
WHERE table_schema = DATABASE()
  AND table_name = ?
  AND constraint_name = 'PRIMARY'
ORDER BY ordinal_position`, table)
	if err != nil {
		return nil, fmt.Errorf("list mysql primary keys for %s: %w", table, err)
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		result = append(result, name)
	}
	return result, rows.Err()
}

func listPostgresColumns(ctx context.Context, db *sql.DB, schema, table string) ([]postgresColumn, error) {
	rows, err := db.QueryContext(ctx, `
SELECT column_name,
       data_type,
       udt_name,
       is_nullable,
       column_default,
       is_identity,
       is_generated
FROM information_schema.columns
WHERE table_schema = $1
  AND table_name = $2
ORDER BY ordinal_position`, schema, table)
	if err != nil {
		return nil, fmt.Errorf("list postgres columns for %s: %w", table, err)
	}
	defer rows.Close()

	var result []postgresColumn
	for rows.Next() {
		var col postgresColumn
		var nullable, identity, generated string
		if err := rows.Scan(&col.Name, &col.DataType, &col.UDTName, &nullable, &col.Default, &identity, &generated); err != nil {
			return nil, err
		}
		col.Nullable = nullable == "YES"
		col.Identity = identity == "YES"
		col.Generated = generated != "NEVER"
		result = append(result, col)
	}
	return result, rows.Err()
}

func listPostgresForeignKeys(ctx context.Context, db *sql.DB, schema string) ([]foreignKey, error) {
	rows, err := db.QueryContext(ctx, `
SELECT child.relname AS child_table,
       parent.relname AS parent_table
FROM pg_constraint c
JOIN pg_class child ON child.oid = c.conrelid
JOIN pg_namespace child_ns ON child_ns.oid = child.relnamespace
JOIN pg_class parent ON parent.oid = c.confrelid
JOIN pg_namespace parent_ns ON parent_ns.oid = parent.relnamespace
WHERE c.contype = 'f'
  AND child_ns.nspname = $1
  AND parent_ns.nspname = $1
ORDER BY child.relname, parent.relname`, schema)
	if err != nil {
		return nil, fmt.Errorf("list postgres foreign keys: %w", err)
	}
	defer rows.Close()

	var result []foreignKey
	for rows.Next() {
		var item foreignKey
		if err := rows.Scan(&item.ChildTable, &item.ParentTable); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func sortPlansByDependencies(plans []tablePlan, dependencies []foreignKey) []tablePlan {
	if len(plans) <= 1 {
		return plans
	}
	planByName := make(map[string]tablePlan, len(plans))
	remaining := make(map[string]struct{}, len(plans))
	for _, plan := range plans {
		planByName[plan.Name] = plan
		remaining[plan.Name] = struct{}{}
	}

	deps := make(map[string]map[string]struct{}, len(plans))
	for _, dep := range dependencies {
		if dep.ChildTable == dep.ParentTable {
			continue
		}
		if _, ok := planByName[dep.ChildTable]; !ok {
			continue
		}
		if _, ok := planByName[dep.ParentTable]; !ok {
			continue
		}
		if deps[dep.ChildTable] == nil {
			deps[dep.ChildTable] = make(map[string]struct{})
		}
		deps[dep.ChildTable][dep.ParentTable] = struct{}{}
	}

	ordered := make([]tablePlan, 0, len(plans))
	for len(remaining) > 0 {
		var ready []string
		for name := range remaining {
			blocked := false
			for parent := range deps[name] {
				if _, ok := remaining[parent]; ok {
					blocked = true
					break
				}
			}
			if !blocked {
				ready = append(ready, name)
			}
		}
		if len(ready) == 0 {
			for name := range remaining {
				ready = append(ready, name)
			}
			sort.Strings(ready)
			log.Printf("foreign key cycle detected among %d table(s); copying remaining tables in lexical order", len(ready))
		} else {
			sort.Strings(ready)
		}
		for _, name := range ready {
			ordered = append(ordered, planByName[name])
			delete(remaining, name)
		}
	}
	return ordered
}

func countMySQLRows(ctx context.Context, db *sql.DB, table string) (int64, error) {
	var count int64
	query := "SELECT COUNT(*) FROM " + quoteMySQLIdent(table)
	if err := db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("count mysql table %s: %w", table, err)
	}
	return count, nil
}

func truncateTables(ctx context.Context, db *sql.DB, schema string, plans []tablePlan) error {
	names := make([]string, 0, len(plans))
	for _, plan := range plans {
		names = append(names, quotePGIdent(schema)+"."+quotePGIdent(plan.Name))
	}
	query := "TRUNCATE TABLE " + strings.Join(names, ", ") + " RESTART IDENTITY CASCADE"
	log.Printf("truncate %d target table(s)", len(names))
	if _, err := db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("truncate target tables: %w", err)
	}
	return nil
}

func copyTable(ctx context.Context, mysqlDB, postgresDB *sql.DB, schema string, plan tablePlan, batchSize int) error {
	log.Printf("copy %s: start", plan.Name)
	cols := make([]string, len(plan.Columns))
	for i, col := range plan.Columns {
		cols[i] = col.Name
	}

	query := "SELECT " + quoteMySQLIdentList(cols) + " FROM " + quoteMySQLIdent(plan.Name)
	if len(plan.OrderColumns) > 0 {
		query += " ORDER BY " + quoteMySQLIdentList(plan.OrderColumns)
	}
	rows, err := mysqlDB.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("query mysql table %s: %w", plan.Name, err)
	}
	defer rows.Close()

	tx, err := postgresDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin postgres transaction for %s: %w", plan.Name, err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	stmt, err := tx.PrepareContext(ctx, pq.CopyInSchema(schema, plan.Name, cols...))
	if err != nil {
		return fmt.Errorf("prepare postgres copy for %s: %w", plan.Name, err)
	}
	stmtClosed := false
	defer func() {
		if !stmtClosed {
			_ = stmt.Close()
		}
	}()

	raw := make([]any, len(cols))
	dest := make([]any, len(cols))
	for i := range raw {
		dest[i] = &raw[i]
	}

	var copied int64
	for rows.Next() {
		if err := rows.Scan(dest...); err != nil {
			return fmt.Errorf("scan mysql row from %s: %w", plan.Name, err)
		}
		values := make([]any, len(raw))
		for i, value := range raw {
			converted, err := convertValue(value, plan.Columns[i])
			if err != nil {
				return fmt.Errorf("convert %s.%s: %w", plan.Name, plan.Columns[i].Name, err)
			}
			values[i] = converted
		}
		if _, err := stmt.ExecContext(ctx, values...); err != nil {
			return fmt.Errorf("copy row into %s: %w", plan.Name, err)
		}
		copied++
		if copied%int64(batchSize) == 0 {
			log.Printf("copy %s: %d/%d row(s)", plan.Name, copied, plan.RowCount)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate mysql rows from %s: %w", plan.Name, err)
	}
	if _, err := stmt.ExecContext(ctx); err != nil {
		return fmt.Errorf("flush postgres copy for %s: %w", plan.Name, err)
	}
	if err := stmt.Close(); err != nil {
		return fmt.Errorf("close postgres copy for %s: %w", plan.Name, err)
	}
	stmtClosed = true
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit postgres copy for %s: %w", plan.Name, err)
	}
	committed = true
	log.Printf("copy %s: done, %d row(s)", plan.Name, copied)
	return nil
}

func convertValue(value any, col postgresColumn) (any, error) {
	if value == nil {
		return nil, nil
	}
	switch v := value.(type) {
	case []byte:
		return convertString(string(v), col)
	case string:
		return convertString(v, col)
	case time.Time:
		if v.IsZero() {
			return nil, nil
		}
		return v, nil
	case bool:
		if isBoolColumn(col) {
			return v, nil
		}
		if isIntegerColumn(col) {
			if v {
				return int64(1), nil
			}
			return int64(0), nil
		}
		return v, nil
	case int64, int32, int, uint64, uint32, uint, float64, float32:
		if isBoolColumn(col) {
			return numericToBool(v)
		}
		return v, nil
	default:
		return value, nil
	}
}

func convertString(value string, col postgresColumn) (any, error) {
	if isZeroDate(value) {
		return nil, nil
	}
	if isBoolColumn(col) {
		return stringToBool(value)
	}
	if isIntegerColumn(col) {
		if value == "" {
			return nil, nil
		}
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, err
		}
		return n, nil
	}
	if isTimestampColumn(col) {
		if value == "" {
			return nil, nil
		}
		t, err := parseTimestamp(value)
		if err != nil {
			return value, nil
		}
		return t, nil
	}
	return value, nil
}

func isBoolColumn(col postgresColumn) bool {
	return col.DataType == "boolean" || col.UDTName == "bool"
}

func isIntegerColumn(col postgresColumn) bool {
	switch col.DataType {
	case "smallint", "integer", "bigint":
		return true
	default:
		return false
	}
}

func isTimestampColumn(col postgresColumn) bool {
	return strings.Contains(col.DataType, "timestamp") || col.DataType == "date"
}

func isZeroDate(value string) bool {
	return strings.HasPrefix(value, "0000-00-00")
}

func stringToBool(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "t", "true", "yes", "y", "on":
		return true, nil
	case "", "0", "f", "false", "no", "n", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value %q", value)
	}
}

func numericToBool(value any) (bool, error) {
	switch v := value.(type) {
	case int64:
		return v != 0, nil
	case int32:
		return v != 0, nil
	case int:
		return v != 0, nil
	case uint64:
		return v != 0, nil
	case uint32:
		return v != 0, nil
	case uint:
		return v != 0, nil
	case float64:
		return v != 0, nil
	case float32:
		return v != 0, nil
	default:
		return false, fmt.Errorf("unsupported numeric boolean type %T", value)
	}
}

func parseTimestamp(value string) (time.Time, error) {
	layouts := []string{
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05.999999",
		"2006-01-02 15:04:05",
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported timestamp %q", value)
}

func resetSequences(ctx context.Context, db *sql.DB, schema string) error {
	rows, err := db.QueryContext(ctx, `
SELECT table_name, column_name
FROM information_schema.columns
WHERE table_schema = $1
  AND (is_identity = 'YES' OR column_default LIKE 'nextval(%')
ORDER BY table_name, ordinal_position`, schema)
	if err != nil {
		return fmt.Errorf("list postgres sequences: %w", err)
	}
	defer rows.Close()

	type sequenceColumn struct {
		table  string
		column string
	}
	var columns []sequenceColumn
	for rows.Next() {
		var item sequenceColumn
		if err := rows.Scan(&item.table, &item.column); err != nil {
			return err
		}
		columns = append(columns, item)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, item := range columns {
		var sequence sql.NullString
		tableName := quotePGIdent(schema) + "." + quotePGIdent(item.table)
		if err := db.QueryRowContext(ctx, `SELECT pg_get_serial_sequence($1, $2)`, tableName, item.column).Scan(&sequence); err != nil {
			return fmt.Errorf("get sequence for %s.%s: %w", item.table, item.column, err)
		}
		if !sequence.Valid || sequence.String == "" {
			continue
		}

		query := fmt.Sprintf("SELECT COALESCE(MAX(%s), 0) FROM %s", quotePGIdent(item.column), tableName)
		var maxID int64
		if err := db.QueryRowContext(ctx, query).Scan(&maxID); err != nil {
			return fmt.Errorf("get max id for %s.%s: %w", item.table, item.column, err)
		}
		if maxID > 0 {
			if _, err := db.ExecContext(ctx, `SELECT setval($1::regclass, $2, true)`, sequence.String, maxID); err != nil {
				return fmt.Errorf("set sequence for %s.%s: %w", item.table, item.column, err)
			}
		} else {
			if _, err := db.ExecContext(ctx, `SELECT setval($1::regclass, 1, false)`, sequence.String); err != nil {
				return fmt.Errorf("reset empty sequence for %s.%s: %w", item.table, item.column, err)
			}
		}
	}
	log.Printf("reset %d postgres sequence(s)", len(columns))
	return nil
}

func quoteMySQLIdent(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

func quoteMySQLIdentList(names []string) string {
	quoted := make([]string, len(names))
	for i, name := range names {
		quoted[i] = quoteMySQLIdent(name)
	}
	return strings.Join(quoted, ", ")
}

func quotePGIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
