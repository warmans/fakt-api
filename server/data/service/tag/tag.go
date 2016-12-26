package tag

import (
	"net/http"

	"github.com/warmans/dbr"
	"github.com/warmans/fakt-api/server/data/service/common"
)

func TagFilterFromRequest(r *http.Request) *TagFilter {
	f := &TagFilter{}
	f.Populate(r)
	return f
}

type TagFilter struct {
	common.CommonFilter
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
		Select("t.id", "t.tag", "SUM(1) AS stat_performers").
		From("tag t").
		LeftJoin("performer_tag", "performer_tag.tag_id = t.id").
		GroupBy("t.id").
		Limit(uint64(filter.PageSize))
	if filter.PageSize != 0 {
		q.Limit(uint64(filter.PageSize)).Offset(uint64((filter.PageSize * page) - filter.PageSize))
	}

	if len(filter.IDs) > 0 {
		q.Where("t.id IN ?", filter.IDs)
	}

	tags := []*common.Tag{}

	if _, err := q.Load(&tags); err != nil && err != dbr.ErrNotFound {
		return tags, err
	}

	return tags, nil
}
