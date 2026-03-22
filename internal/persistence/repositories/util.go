package repositories

import (
	"github.com/mairuu/mp-api/internal/app/ordering"
	"github.com/mairuu/mp-api/internal/app/paging"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func applyPagging(q *gorm.DB, p paging.Paging) *gorm.DB {
	q = q.Limit(p.Limit).Offset(p.Offset)
	return q
}

func applyOrderings(q *gorm.DB, os []ordering.Ordering) *gorm.DB {
	for _, o := range os {
		if o.Field == "" && !o.Direction.IsValid() {
			continue
		}
		q = q.Order(clause.OrderByColumn{
			Column: clause.Column{Name: string(o.Field), Raw: true},
			Desc:   o.Direction == ordering.Desc,
		})
	}

	return q
}
