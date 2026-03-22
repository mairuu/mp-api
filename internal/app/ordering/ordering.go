package ordering

type Field string

type Ordering struct {
	Field     Field
	Direction Direction
}

type Direction string

const (
	Asc  Direction = "ASC"
	Desc Direction = "DESC"
)

func (d Direction) IsValid() bool {
	return d == Asc || d == Desc
}
