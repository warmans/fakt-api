package utag

import (
	"strings"
	"github.com/warmans/dbr"
	"github.com/warmans/dbr/dialect"
	"github.com/warmans/fakt-api/pkg/server/data/service/common"
)

type UTagService struct {
	DB *dbr.Session
}

func (s *UTagService) findUTags(q *dbr.SelectBuilder) ([]common.UTags, error) {

	sql, vals := q.ToSql()
	interpolated, err := dbr.InterpolateForDialect(sql, vals, dialect.SQLite3)
	if err != nil {
		return nil, err
	}

	result, err := s.DB.Query(interpolated)
	if err != nil && err != dbr.ErrNotFound {
		return nil, err
	}

	defer result.Close()

	utags := make([]common.UTags, 0)
	var username, tagString string
	for result.Next() {
		if err := result.Scan(&username, &tagString); err != nil {
			return nil, err
		}
		utags = append(utags, common.UTags{Username: username, Values: strings.Split(tagString, ";")})
	}
	return utags, nil
}

func (us *UTagService) FindEventUTags(eventID int64, filter *common.UTagsFilter) ([]common.UTags, error) {
	q := us.DB.Select("user.username", "GROUP_CONCAT(event_user_tag.tag, ';')").
	From("event_user_tag").
	Where("event_id = ?", eventID).
	LeftJoin("user", "event_user_tag.user_id = user.id").
	GroupBy("event_user_tag.event_id", "event_user_tag.user_id")

	if filter.Username != "" {
		q.Where("user.username = ?", filter.Username)
	}
	return us.findUTags(q)
}

func (s *UTagService) StoreEventUTags(eventID int64, userID int64, tags []string) error {
	for _, tag := range tags {
		_, err := s.DB.Exec("INSERT OR IGNORE INTO event_user_tag (event_id, user_id, tag) VALUES (?, ?, ?)", eventID, userID, tag)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *UTagService) RemoveEventUTags(eventID int64, userID int64, tags []string) error {
	for _, tag := range tags {
		_, err := s.DB.Exec("DELETE FROM event_user_tag WHERE event_id=? AND user_id=? AND tag=?", eventID, userID, tag)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *UTagService) StorePerformerUTags(performerID int64, userID int64, tags []string) error {
	for _, tag := range tags {
		_, err := s.DB.Exec("INSERT OR IGNORE INTO performer_user_tag (performer_id, user_id, tag) VALUES (?, ?, ?)", performerID, userID, tag)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *UTagService) RemovePerformerUTags(performerID int64, userID int64, tags []string) error {
	for _, tag := range tags {
		_, err := s.DB.Exec("DELETE FROM performer_user_tag WHERE performer_id=? AND user_id=? AND tag=?", performerID, userID, tag)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *UTagService) FindPerformerUTags(performerID int64, filter *common.UTagsFilter) ([]common.UTags, error) {
	q := s.DB.Select("coalesce(user.username, '')", "GROUP_CONCAT(performer_user_tag.tag, ';')").
	From("performer_user_tag").
	Where("performer_id = ?", performerID).
	LeftJoin("user", "performer_user_tag.user_id = user.id").
	GroupBy("performer_user_tag.performer_id", "performer_user_tag.user_id")

	if filter.Username != "" {
		q.Where("user.username = ?", filter.Username)
	}
	return s.findUTags(q)
}
