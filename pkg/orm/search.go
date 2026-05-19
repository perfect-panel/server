package orm

import (
	"strings"

	"gorm.io/gorm"
)

func LikePrefixPattern(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return escapeLike(value) + "%"
}

func LikeContainsPattern(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return "%" + escapeLike(value) + "%"
}

func PrefixLike(fields []string, value string) func(db *gorm.DB) *gorm.DB {
	return likeSearch(fields, LikePrefixPattern(value))
}

func ContainsLike(fields []string, value string) func(db *gorm.DB) *gorm.DB {
	return likeSearch(fields, LikeContainsPattern(value))
}

func likeSearch(fields []string, pattern string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(fields) == 0 || pattern == "" {
			return db
		}

		conds := make([]string, 0, len(fields))
		args := make([]interface{}, 0, len(fields))
		for _, field := range fields {
			if field == "" {
				continue
			}
			conds = append(conds, field+" LIKE ? ESCAPE '\\'")
			args = append(args, pattern)
		}
		if len(conds) == 0 {
			return db
		}
		return db.Where("("+strings.Join(conds, " OR ")+")", args...)
	}
}

func escapeLike(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(value)
}
