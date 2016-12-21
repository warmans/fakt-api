package tag

import (
	"github.com/warmans/dbr"
	"github.com/warmans/fakt-api/server/data/service/common"
	"net/http"
	"strconv"
)

const DefaultPageSize = 50

func TagFilterFromRequest(r *http.Request) *TagFilter {
	f := &TagFilter{}
	f.Populate(r)
	return f
}

type TagFilter struct {
	Page              int64     `json:"page"`
	PageSize          int64     `json:"page_size"`
}

func (f *TagFilter)Populate(r *http.Request) {
	if page := r.Form.Get("page"); page != "" {
		if pageInt, err := strconv.Atoi(page); err == nil {
			f.Page = int64(pageInt)
		}
	}
	f.PageSize = DefaultPageSize
	if pageSize := r.Form.Get("page_size"); pageSize != "" {
		if pageSizeInt, err := strconv.Atoi(pageSize); err == nil {
			if pageSizeInt > 0 {
				f.PageSize = int64(pageSizeInt)
			}
		}
	}
}

type TagService struct {
	DB *dbr.Session
}

func (s *TagService) FindTags(filter *TagFilter) ([]*common.Tag, error) {

	page := filter.Page
	if page == 0 {
		page = 1
	}

	q := s.DB.
		Select("id", "tag").
		From("tag").
		Limit(uint64(filter.PageSize)).
		Offset(uint64((filter.PageSize * page) - filter.PageSize))

	tags := []*common.Tag{}

	if _, err := q.Load(&tags); err != nil && err != dbr.ErrNotFound  {
		return tags, err
	}

	return tags, nil
}
