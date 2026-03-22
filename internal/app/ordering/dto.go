package ordering

import (
	"slices"
	"strings"
)

type Query struct {
	// syntax: order=field1,asc&order=field2,desc
	Orders []string `form:"orders[]"`
}

func (o *Query) ToOrdering(validFields ...Field) []Ordering {
	var orderings []Ordering
	for _, order := range o.Orders {
		parts := strings.Split(order, ",")
		if len(parts) != 2 {
			continue
		}
		field := Field(parts[0])
		direction := parts[1]

		if !slices.Contains(validFields, field) {
			continue
		}

		dir := Asc
		switch direction {
		case "desc", "DESC":
			dir = Desc
		}

		orderings = append(orderings, Ordering{
			Field:     field,
			Direction: dir,
		})
	}

	return orderings
}
