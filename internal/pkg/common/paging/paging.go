package paging

type PagingQuery struct {
	Paging   bool `form:"paging" json:"paging"`
	Page     int  `form:"page" json:"page" validate:"omitempty,gte=1"`
	PageSize int  `form:"page_size" json:"page_size" validate:"omitempty,gte=1,lte=1000"`
}

// SetDefaults 设置默认分页, 限制每页项目最大数
func (p *PagingQuery) SetDefaults(defaultPage, defaultSize, maxSize int) {
	if p.Page <= 0 {
		p.Page = defaultPage
	}
	if p.PageSize <= 0 {
		p.PageSize = defaultSize
	}
	if maxSize > 0 && p.PageSize > maxSize {
		p.PageSize = maxSize
	}
}
