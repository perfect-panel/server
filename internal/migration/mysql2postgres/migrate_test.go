package mysql2postgres

import (
	"database/sql"
	"strings"
	"testing"
	"time"
)

func TestQuoteIdentifiers(t *testing.T) {
	if got := quoteMySQLIdent("or`der"); got != "`or``der`" {
		t.Fatalf("quoteMySQLIdent() = %q", got)
	}
	if got := quotePGIdent(`or"der`); got != `"or""der"` {
		t.Fatalf("quotePGIdent() = %q", got)
	}
}

func TestNormalizeMySQLDSN(t *testing.T) {
	got := normalizeMySQLDSN("user:pass@tcp(127.0.0.1:3306)/ppanel")
	if got == "user:pass@tcp(127.0.0.1:3306)/ppanel" {
		t.Fatalf("normalizeMySQLDSN() did not add params")
	}
	if !containsAll(got, []string{"parseTime=true", "charset=utf8mb4"}) {
		t.Fatalf("normalizeMySQLDSN() = %q, want parseTime and charset", got)
	}
}

func TestConvertValueBoolean(t *testing.T) {
	col := postgresColumn{DataType: "boolean", UDTName: "bool"}
	tests := []struct {
		name  string
		input any
		want  bool
	}{
		{name: "int64 true", input: int64(1), want: true},
		{name: "int64 false", input: int64(0), want: false},
		{name: "bytes true", input: []byte("1"), want: true},
		{name: "bytes false", input: []byte("false"), want: false},
		{name: "string true", input: "true", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertValue(tt.input, col)
			if err != nil {
				t.Fatalf("convertValue() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("convertValue() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestConvertValueInteger(t *testing.T) {
	col := postgresColumn{DataType: "bigint", UDTName: "int8"}
	got, err := convertValue([]byte("42"), col)
	if err != nil {
		t.Fatalf("convertValue() error = %v", err)
	}
	if got != int64(42) {
		t.Fatalf("convertValue() = %#v, want int64(42)", got)
	}
}

func TestConvertValueTimestamp(t *testing.T) {
	col := postgresColumn{DataType: "timestamp without time zone", UDTName: "timestamp"}
	got, err := convertValue([]byte("2026-05-21 14:30:00"), col)
	if err != nil {
		t.Fatalf("convertValue() error = %v", err)
	}
	if _, ok := got.(time.Time); !ok {
		t.Fatalf("convertValue() = %T, want time.Time", got)
	}

	got, err = convertValue([]byte("0000-00-00 00:00:00"), col)
	if err != nil {
		t.Fatalf("convertValue() zero date error = %v", err)
	}
	if got != nil {
		t.Fatalf("convertValue() zero date = %#v, want nil", got)
	}
}

func TestBuildPlansSkipsGeneratedColumns(t *testing.T) {
	cols := []postgresColumn{
		{Name: "id", DataType: "bigint"},
		{Name: "computed", DataType: "text", Generated: true},
		{Name: "missing", DataType: "text"},
		{Name: "name", DataType: "text"},
	}
	source := map[string]struct{}{
		"id":       {},
		"computed": {},
		"name":     {},
	}
	common := make([]postgresColumn, 0, len(cols))
	for _, col := range cols {
		if col.Generated {
			continue
		}
		if _, ok := source[col.Name]; ok {
			common = append(common, col)
		}
	}
	if len(common) != 2 || common[0].Name != "id" || common[1].Name != "name" {
		t.Fatalf("common columns = %#v", common)
	}
}

func TestParseTableSet(t *testing.T) {
	got := parseTableSet(" user, order ,,payment ")
	for _, name := range []string{"user", "order", "payment"} {
		if _, ok := got[name]; !ok {
			t.Fatalf("missing table %q in %#v", name, got)
		}
	}
	if len(got) != 3 {
		t.Fatalf("parseTableSet length = %d, want 3", len(got))
	}
}

func TestSortPlansByDependencies(t *testing.T) {
	plans := []tablePlan{
		{Name: "order"},
		{Name: "user_subscribe"},
		{Name: "user"},
		{Name: "subscribe"},
	}
	dependencies := []foreignKey{
		{ChildTable: "order", ParentTable: "user"},
		{ChildTable: "order", ParentTable: "subscribe"},
		{ChildTable: "user_subscribe", ParentTable: "user"},
		{ChildTable: "user_subscribe", ParentTable: "subscribe"},
	}
	got := sortPlansByDependencies(plans, dependencies)

	positions := make(map[string]int, len(got))
	for i, plan := range got {
		positions[plan.Name] = i
	}
	for _, dep := range dependencies {
		if positions[dep.ParentTable] > positions[dep.ChildTable] {
			t.Fatalf("%s should be copied before %s; got %#v", dep.ParentTable, dep.ChildTable, got)
		}
	}
}

func TestSortPlansByDependenciesIgnoresSelfReferences(t *testing.T) {
	got := sortPlansByDependencies(
		[]tablePlan{{Name: "order"}, {Name: "user"}},
		[]foreignKey{
			{ChildTable: "order", ParentTable: "order"},
			{ChildTable: "order", ParentTable: "user"},
		},
	)
	if len(got) != 2 || got[0].Name != "user" || got[1].Name != "order" {
		t.Fatalf("sortPlansByDependencies() = %#v", got)
	}
}

func TestIsBoolColumn(t *testing.T) {
	if !isBoolColumn(postgresColumn{DataType: "boolean"}) {
		t.Fatal("boolean data type should be bool")
	}
	if !isBoolColumn(postgresColumn{UDTName: "bool"}) {
		t.Fatal("bool udt should be bool")
	}
	if isBoolColumn(postgresColumn{DataType: "smallint"}) {
		t.Fatal("smallint should not be bool")
	}
}

func TestNullableNullIsPreserved(t *testing.T) {
	got, err := convertValue(nil, postgresColumn{DataType: "text", Nullable: true, Default: sql.NullString{}})
	if err != nil {
		t.Fatalf("convertValue() error = %v", err)
	}
	if got != nil {
		t.Fatalf("convertValue(nil) = %#v, want nil", got)
	}
}

func containsAll(value string, parts []string) bool {
	for _, part := range parts {
		if !strings.Contains(value, part) {
			return false
		}
	}
	return true
}
