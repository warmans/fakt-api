package tag

import (
	"net/http"

	"github.com/warmans/dbr"
	"github.com/warmans/fakt-api/pkg/server/data/service/common"
)

func TagFilterFromRequest(r *http.Request) *TagFilter {
	f := &TagFilter{
		WithEvents:  common.StringToBool(r.Form.Get("with_events")),
		WithPerformers: common.StringToBool(r.Form.Get("with_performers")),
	}
	f.Populate(r)
	return f
}

type TagFilter struct {
	common.CommonFilter
	WithEvents     bool `json:"with_events"`
	WithPerformers bool `json:"with_performers"`
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
		Select(
		"t.id",
		"t.tag",
		"COUNT(DISTINCT performer_tag.performer_id) AS stat_performers",
		"COUNT(DISTINCT event.id) AS stat_events",
		).
		From("tag t").
		LeftJoin("performer_tag", "performer_tag.tag_id = t.id").
		LeftJoin("event_tag", "event_tag.tag_id = t.id").
		LeftJoin("event", "event_tag.event_id = event.id AND event.deleted = 0").
		GroupBy("t.id").
		Limit(uint64(filter.PageSize))

	if filter.PageSize != 0 {
		q.Limit(uint64(filter.PageSize)).Offset(uint64((filter.PageSize * page) - filter.PageSize))
	}

	if len(filter.IDs) > 0 {
		q.Where("t.id IN ?", filter.IDs)
	}
	if filter.WithEvents {
		q.Having("stat_events > 0")
		q.OrderDesc("stat_events")
	}
	if filter.WithPerformers {
		q.Having("stat_performers > 0")
		q.OrderDesc("stat_performers")
	}

	tags := []*common.Tag{}

	if _, err := q.Load(&tags); err != nil && err != dbr.ErrNotFound {
		return tags, err
	}

	return tags, nil
}
