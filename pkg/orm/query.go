package orm

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// CommaSeparatedContains filters comma-separated string columns, such as "1,2,3".
func CommaSeparatedContains(field string, values []string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		values = removeEmpty(values)
		if len(values) == 0 {
			return db
		}

		if db.Dialector.Name() == DriverMySQL {
			conds := make([]string, len(values))
			args := make([]interface{}, len(values))
			for i, v := range values {
				conds[i] = "FIND_IN_SET(?, " + field + ")"
				args[i] = v
			}
			return db.Where("("+strings.Join(conds, " OR ")+")", args...)
		}

		conds := make([]string, len(values))
		args := make([]interface{}, len(values))
		for i, v := range values {
			conds[i] = "(',' || COALESCE(" + field + ", '') || ',') LIKE ?"
			args[i] = "%," + v + ",%"
		}
		return db.Where("("+strings.Join(conds, " OR ")+")", args...)
	}
}

func removeEmpty(values []string) []string {
	list := values[:0]
	for _, value := range values {
		if value != "" {
			list = append(list, value)
		}
	}
	return list
}

func TextColumnExpr(db *gorm.DB, field string) string {
	if db.Dialector.Name() == DriverPostgres {
		return fmt.Sprintf("CAST(%s AS TEXT)", field)
	}
	return fmt.Sprintf("CAST(%s AS CHAR)", field)
}
