package user

import (
	"strings"
	"testing"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestMonthlyResetSubscribeQueryUsesStartTime(t *testing.T) {
	tests := []struct {
		name      string
		dialector gorm.Dialector
		want      []string
	}{
		{
			name: "mysql",
			dialector: mysql.New(mysql.Config{
				DSN:                       "gorm:gorm@tcp(localhost:9910)/gorm?charset=utf8&parseTime=True&loc=Local",
				SkipInitializeWithVersion: true,
			}),
			want: []string{
				"DAY(start_time) >= ?",
				"start_time <= ?",
				"(expire_time IS NULL OR expire_time = ? OR expire_time > ?)",
			},
		},
		{
			name: "postgres",
			dialector: postgres.New(postgres.Config{
				DSN:                  "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable",
				PreferSimpleProtocol: true,
			}),
			want: []string{
				"EXTRACT(day FROM start_time) >=",
				"start_time <= ",
				"expire_time IS NULL OR expire_time = ",
				" OR expire_time > ",
			},
		},
	}

	now := time.Date(2026, 4, 30, 0, 30, 0, 0, time.Local)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := gorm.Open(tt.dialector, &gorm.Config{
				DryRun:               true,
				DisableAutomaticPing: true,
			})
			if err != nil {
				t.Fatalf("open gorm db: %v", err)
			}

			var ids []int64
			stmt := monthlyResetSubscribeQuery(db, []int64{1, 2}, now).Find(&ids).Statement
			sql := stmt.SQL.String()
			for _, want := range tt.want {
				if !strings.Contains(sql, want) {
					t.Fatalf("SQL missing %q:\n%s", want, sql)
				}
			}
			for _, unwanted := range []string{"DAY(expire_time)", "EXTRACT(day FROM expire_time)", "TIMESTAMPDIFF"} {
				if strings.Contains(sql, unwanted) {
					t.Fatalf("SQL should not contain %q:\n%s", unwanted, sql)
				}
			}
		})
	}
}

func TestYearlyResetDateConditionHandlesLeapFallback(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       "gorm:gorm@tcp(localhost:9910)/gorm?charset=utf8&parseTime=True&loc=Local",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		DryRun:               true,
		DisableAutomaticPing: true,
	})
	if err != nil {
		t.Fatalf("open gorm db: %v", err)
	}

	condition, args := yearlyResetDateCondition(db, time.Date(2025, 2, 28, 0, 30, 0, 0, time.Local))
	if !strings.Contains(condition, "MONTH(start_time) = ?") || !strings.Contains(condition, "DAY(start_time) IN ?") {
		t.Fatalf("unexpected condition: %s", condition)
	}
	if len(args) != 2 {
		t.Fatalf("args len = %d, want 2: %#v", len(args), args)
	}
	days, ok := args[1].([]int)
	if !ok || len(days) != 2 || days[0] != 28 || days[1] != 29 {
		t.Fatalf("leap fallback days = %#v, want []int{28, 29}", args[1])
	}
}
