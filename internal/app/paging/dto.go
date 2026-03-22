package paging

const (
	DefaultPageSize = 20
)

type PagedDTO struct {
	Items      any           `json:"items"`
	Pagination PaginationDTO `json:"pagination"`
}

type PaginationDTO struct {
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
}

func NewPagedDTO(total, totalPage, pageSize, page int, items any) PagedDTO {
	return PagedDTO{
		Items: items,
		Pagination: PaginationDTO{
			TotalItems: total,
			TotalPages: totalPage,
			PageSize:   pageSize,
			Page:       page,
		},
	}
}

type Query struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

func (p *Query) normalize() {
	if p.PageSize <= 0 {
		p.PageSize = DefaultPageSize
	}
	if p.Page <= 0 {
		p.Page = 1
	}
}

func (p *Query) getLimitOffset() (limit, offset int) {
	limit = p.PageSize
	offset = (p.Page - 1) * p.PageSize
	return
}

func (p *Query) ToPaging() Paging {
	p.normalize()
	limit, offset := p.getLimitOffset()
	return Paging{Limit: limit, Offset: offset}
}
