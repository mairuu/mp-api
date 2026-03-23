package service

import (
	"github.com/mairuu/mp-api/internal/app/paging"
)

type HistoryListQuery struct {
	PagingQuery
}

type PagingQuery struct {
	paging.Query
}
