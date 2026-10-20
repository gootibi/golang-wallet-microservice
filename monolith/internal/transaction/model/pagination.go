package model

type PaginationParams struct {
	Page   int    `form:"page,default=1"`
	Limit  int    `form:"limit,default=10"`
	Sort   string `form:"sort,default=created_at"`
	Order  string `form:"order,default=desc"`
	Status string `form:"status"` // Optional filter status
}

func (p *PaginationParams) Offset() int {
	return (p.Page - 1) * p.Limit
}

type PaginationMeta struct {
	Page      int   `json:"page"`
	Limit     int   `json:"limit"`
	Total     int64 `json:"total"`
	TotalPage int   `json:"total_page"`
}

// TypeSave solution
// type PaginationResponse[T any] struct {
// 	Success bool           `json:"success"`
// 	Data    T              `json:"data"`
// 	Meta    PaginationMeta `json:"meta"`
// }

type PaginationResponse struct {
	Success bool           `json:"success"`
	Data    any            `json:"data"`
	Meta    PaginationMeta `json:"meta"`
}
