package paging

import "github.com/go-playground/validator/v10"

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

type PagingParam struct {
	Paging   bool `form:"paging,default=true" json:"paging"`     // 默认为 true
	Page     int  `form:"page,default=1" json:"page"`            // 当 paging == true 时, page 被使用. 默认为1, 最小值为1.
	PageSize int  `form:"page_size,default=10" json:"page_size"` // 当 paging == true 时, page_size 被使用. 默认为10, 最小值为1, 最大值为100.
}

// StructLevelValidationWithPagingParam PagingParam's struct-level validator.
func StructLevelValidationWithPagingParam(sl validator.StructLevel) {
	pp, ok := sl.Current().Interface().(PagingParam)
	if !ok {
		panic("receive type error of variable in pagingParamValidator")
	}

	// when paging == false, no validate for page, page_size.
	if !pp.Paging {
		return
	}

	if pp.Page < 1 {
		sl.ReportError(pp.Page, "page", "Page", "min", "1")
	}

	if pp.PageSize < 1 {
		sl.ReportError(pp.PageSize, "page_size", "PageSize", "min", "1")
	}

	if pp.PageSize > 100 {
		sl.ReportError(pp.PageSize, "page_size", "PageSize", "max", "100")
	}
}
