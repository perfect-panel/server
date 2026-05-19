package system

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func WhereKey(key string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(clause.Eq{
			Column: clause.Column{Name: "key"},
			Value:  key,
		})
	}
}

func WhereCategoryKey(category, key string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(clause.Eq{
			Column: clause.Column{Name: "category"},
			Value:  category,
		}).Where(clause.Eq{
			Column: clause.Column{Name: "key"},
			Value:  key,
		})
	}
}
