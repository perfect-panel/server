package user

import (
	"strings"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ptr[T any](v T) *T {
	return &v
}

func TestApplyUserPageFiltersSearchSQL(t *testing.T) {
	tests := []struct {
		name       string
		dialector  gorm.Dialector
		wantSQL    []string
		wantNoSQL  []string
		wantSearch string
	}{
		{
			name: "mysql",
			dialector: mysql.New(mysql.Config{
				DSN:                       "gorm:gorm@tcp(localhost:9910)/gorm?charset=utf8&parseTime=True&loc=Local",
				SkipInitializeWithVersion: true,
			}),
			wantSQL: []string{
				"FROM `user`",
				"`user`.`id` = ?",
				"`user`.`refer_code` LIKE ? ESCAPE '='",
				"EXISTS (SELECT 1 FROM `user_auth_methods`",
				"`user_auth_methods`.`user_id` = `user`.`id`",
				"`user_auth_methods`.`auth_identifier` LIKE ? ESCAPE '='",
				"EXISTS (SELECT 1 FROM `user_subscribe`",
				"`user_subscribe`.`user_id` = `user`.`id`",
				"`user_subscribe`.`id` = ?",
				"`user_subscribe`.`subscribe_id` = ?",
				"`user_subscribe`.`status` IN (?,?)",
				"ORDER BY `user`.`id` DESC",
			},
			wantNoSQL:  []string{"LEFT JOIN", "GROUP BY"},
			wantSearch: "alice=_100=%@example.com%",
		},
		{
			name: "postgres",
			dialector: postgres.New(postgres.Config{
				DSN:                  "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable",
				PreferSimpleProtocol: true,
			}),
			wantSQL: []string{
				`FROM "user"`,
				`"user"."id" = $1`,
				`"user"."refer_code" LIKE $2 ESCAPE '='`,
				`EXISTS (SELECT 1 FROM "user_auth_methods"`,
				`"user_auth_methods"."user_id" = "user"."id"`,
				`"user_auth_methods"."auth_identifier" LIKE $3 ESCAPE '='`,
				`EXISTS (SELECT 1 FROM "user_subscribe"`,
				`"user_subscribe"."user_id" = "user"."id"`,
				`"user_subscribe"."id" = $4`,
				`"user_subscribe"."subscribe_id" = $7`,
				`"user_subscribe"."status" IN ($5,$6)`,
				`"user_subscribe"."status" IN ($8,$9)`,
				`ORDER BY "user"."id" DESC`,
			},
			wantNoSQL:  []string{"LEFT JOIN", "GROUP BY"},
			wantSearch: "alice=_100=%@example.com%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := gorm.Open(tt.dialector, &gorm.Config{
				DryRun:               true,
				DisableAutomaticPing: true,
			})
			if err != nil {
				t.Fatalf("open gorm db: %v", err)
			}

			var result []User
			filter := &UserFilterParams{
				Search:          "alice_100%@example.com",
				UserId:          ptr[int64](99),
				UserSubscribeId: ptr[int64](10),
				SubscribeId:     ptr[int64](20),
				Order:           "DESC",
			}
			stmt := applyUserPageFilters(db.Model(&User{}), filter).Find(&result).Statement
			sql := stmt.SQL.String()

			for _, want := range tt.wantSQL {
				if !strings.Contains(sql, want) {
					t.Fatalf("SQL missing %q:\n%s", want, sql)
				}
			}
			for _, unwanted := range tt.wantNoSQL {
				if strings.Contains(sql, unwanted) {
					t.Fatalf("SQL should not contain %q:\n%s", unwanted, sql)
				}
			}
			if got := stmt.Vars[1]; got != tt.wantSearch {
				t.Fatalf("refer_code search pattern = %#v, want %#v", got, tt.wantSearch)
			}
			if got := stmt.Vars[2]; got != tt.wantSearch {
				t.Fatalf("auth search pattern = %#v, want %#v", got, tt.wantSearch)
			}
		})
	}
}

func TestApplyUserPageFiltersSkipsBlankSearch(t *testing.T) {
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

	var result []User
	stmt := applyUserPageFilters(db.Model(&User{}), &UserFilterParams{Search: "   "}).Find(&result).Statement
	sql := stmt.SQL.String()
	if strings.Contains(sql, "LIKE") || strings.Contains(sql, "user_auth_methods") {
		t.Fatalf("blank search should not add search filters:\n%s", sql)
	}
	if len(stmt.Vars) != 0 {
		t.Fatalf("vars len = %d, want 0: %#v", len(stmt.Vars), stmt.Vars)
	}
}
