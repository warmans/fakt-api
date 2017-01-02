package performer

import (
	"database/sql"
	"net/http"

	"fmt"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/warmans/dbr"
	"github.com/warmans/fakt-api/server/data/service/common"
	"github.com/warmans/fakt-api/server/data/service/utag"
)

func PerformerFilterFromRequest(r *http.Request) *PerformerFilter {
	f := &PerformerFilter{}
	f.Populate(r)
	return f
}

type PerformerFilter struct {
	common.CommonFilter

	Name  string `json:"name"`
	Genre string `json:"name"`
	Home  string `json:"name"`
}

func (f *PerformerFilter) Populate(r *http.Request) {

	f.CommonFilter.Populate(r)

	f.Name = r.Form.Get("name")
	f.Genre = r.Form.Get("genre")
	f.Home = r.Form.Get("home")
}

type PerformerService struct {
	DB          *dbr.Session
	UTagService *utag.UTagService
	Logger      log.Logger
}

func (s *PerformerService) FindPerformers(filter *PerformerFilter) ([]*common.Performer, error) {

	//if no page is specified assume the first page
	page := filter.Page
	if page == 0 {
		page = 1
	}

	q := s.DB.
		Select("id", "name", "info", "genre", "home", "listen_url", "embed_url", "COALESCE(activity, 0) AS activity").
		From("performer p").
		OrderBy("p.name")

	if filter.PageSize != 0 {
		q.Limit(uint64(filter.PageSize)).Offset(uint64((filter.PageSize * page) - filter.PageSize))
	}

	if len(filter.IDs) > 0 {
		q.Where("p.id IN ?", filter.IDs)
	}
	if filter.Name != "" {
		q.Where("p.name = ?", filter.Name)
	}
	if filter.Home != "" {
		q.Where("p.home = ?", filter.Home)
	}
	if filter.Genre != "" {
		q.Where("p.genre = ?", filter.Genre)
	}

	found := make([]*common.Performer, 0)
	if _, err := q.Load(&found); err != nil && err != dbr.ErrNotFound {
		return nil, err
	}

	for k, performer := range found {
		links, err := s.FindPerformerLinks(performer.ID)
		if err != nil {
			return nil, err
		}
		found[k].Links = links

		//append the utags
		tags, err := s.UTagService.FindPerformerUTags(performer.ID, &common.UTagsFilter{})
		if err != nil {
			return nil, err
		}
		found[k].UTags = tags

		//append normal tags
		if found[k].Tags, err = s.FindPerformerTags(performer.ID); err != nil {
			return nil, err
		}

		//images
		if found[k].Images, err = s.FindPerformerImages(performer.ID); err != nil {
			return nil, err
		}
	}

	return found, nil
}

func (s *PerformerService) FindPerformerLinks(performerId int64) ([]*common.Link, error) {
	q := s.DB.
		Select("link", "link_type", "link_description").
		From("performer_extra").
		Where("performer_id = ?", performerId)

	links := make([]*common.Link, 0)
	if _, err := q.Load(&links); err != nil && err != dbr.ErrNotFound {
		return nil, err
	}
	return links, nil
}

func (s *PerformerService) FindPerformerTags(performerID int64) ([]string, error) {

	tags := []string{}

	res, err := s.DB.Query("SELECT coalesce(t.tag, '') FROM performer_tag pt LEFT JOIN tag t ON pt.tag_id = t.id WHERE pt.performer_id = ?", performerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return tags, nil
		}
		return tags, fmt.Errorf("Failed to fetch tags at query because %s", err.Error())
	}

	for res.Next() {
		tag := ""
		if err := res.Scan(&tag); err != nil {
			return tags, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func (s *PerformerService) FindPerformerImages(performerID int64) (map[string]string, error) {

	images := make(map[string]string)

	res, err := s.DB.Query("SELECT pi.usage, pi.src FROM performer_image pi WHERE pi.performer_id = ?", performerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return images, nil
		}
		return images, fmt.Errorf("Failed performer images query: %s", err.Error())
	}

	for res.Next() {
		var usage, src string
		if err := res.Scan(&usage, &src); err != nil {
			return images, fmt.Errorf("Failed performer image scan: %s", err.Error())
		}
		images[usage] = src
	}

	return images, nil
}

func (s *PerformerService) FindPerformerEventIDs(performerID int64) ([]int64, error) {

	eventIDs := make([]int64, 0)

	res, err := s.DB.Query("SELECT event_id FROM event_performer WHERE performer_id = ?", performerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return eventIDs, nil
		}
		return eventIDs, fmt.Errorf("Failed to fetch performer events because of SQL error %s", err.Error())
	}

	for res.Next() {
		var eventID int64
		if err := res.Scan(&eventID); err != nil {
			return eventIDs, fmt.Errorf("Failed to fetch performer events because of scan error %s", err.Error())
		}
		eventIDs = append(eventIDs, eventID)
	}

	return eventIDs, nil
}

func (s *PerformerService) FindSimilarPerformerIDs(performerID int64) ([]int64, error) {
	performerIDs := make([]int64, 0)

	res, err := s.DB.Query(`
		SELECT pt2.performer_id
		FROM performer_tag pt1
		LEFT JOIN performer_tag pt2 ON pt1.tag_id = pt2.tag_id
		WHERE pt1.performer_id = 7 AND pt2.performer_id != 7
		GROUP BY pt2.performer_id
		ORDER BY SUM(1) DESC
	`)
	if err != nil {
		if err == sql.ErrNoRows {
			return performerIDs, nil
		}
		return performerIDs, fmt.Errorf("Failed to fetch similar performer because of SQL error %s", err.Error())
	}

	for res.Next() {
		var perfID int64
		if err := res.Scan(&perfID); err != nil {
			return performerIDs, fmt.Errorf("Failed to fetch similar performers because of scan error %s", err.Error())
		}
		performerIDs = append(performerIDs, perfID)
	}

	return performerIDs, nil
}

func (ps *PerformerService) PerformerMustExist(tr *dbr.Tx, performer *common.Performer) error {

	if !performer.IsValid() {
		return nil
	}

	if performer.ID == 0 {
		//geg the performer based on their name and genre
		err := tr.QueryRow("SELECT id FROM performer WHERE name=? AND genre=?", performer.Name, performer.Genre).Scan(&performer.ID)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
	}
	if performer.ID == 0 {
		res, err := tr.Exec(
			"INSERT INTO performer (name, info, genre, home, listen_url, embed_url) VALUES (?, ?, ?, ?, ?, ?)",
			performer.Name,
			performer.Info,
			performer.Genre,
			performer.Home,
			performer.ListenURL,
			performer.EmbedURL,
		)
		if err != nil {
			return err
		}
		performer.ID, err = res.LastInsertId()
		if err != nil {
			return err
		}
	} else {
		_, err := tr.Exec(
			"UPDATE performer SET info=?, home=?, listen_url=? WHERE id=?",
			performer.Info,
			performer.Home,
			performer.ListenURL,
			performer.ID,
		)
		if err != nil {
			return err
		}
	}

	//clear existing relationships for extra data to allow links to be kept up-to-date
	_, err := tr.Exec("DELETE FROM performer_extra WHERE performer_id=?", performer.ID)
	if err != nil {
		return err
	}
	for _, link := range performer.Links {
		_, err := tr.Exec(
			"INSERT INTO performer_extra (performer_id, link, link_type, link_description) VALUES (?, ?, ?, ?)",
			performer.ID,
			link.URI,
			link.Type,
			link.Text,
		)
		if err != nil {
			ps.Logger.Log("msg", fmt.Sprintf("Failed to insert performer_extra: %s", err.Error()))
			continue
		}
	}

	//try and store additional entities but just log errors instead of failing for now
	if err := ps.StorePerformerTags(tr, performer.ID, performer.Tags); err != nil {
		ps.Logger.Log(err.Error())
	}
	if err := ps.StorePerformerImages(tr, performer.ID, performer.Images); err != nil {
		ps.Logger.Log(err.Error())
	}

	return nil
}

func (s *PerformerService) StorePerformerTags(tr *dbr.Tx, performerID int64, tags []string) error {
	//handle tags
	if _, err := tr.Exec("DELETE FROM performer_tag WHERE performer_id = ?", performerID); err != nil {
		s.Logger.Log("msg", fmt.Sprintf("Failed to delete existing performer_tag relationships (perfomer: %d) because %s", performerID, err.Error()))
	}

	//todo: move this into new tag service
	for _, tag := range tags {

		var tagId int64
		tag = strings.ToLower(tag)

		err := s.DB.QueryRow("SELECT id FROM tag WHERE tag = ?", tag).Scan(&tagId)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("Failed to find tag id for %s because %s", tag, err.Error())
		}
		if tagId == 0 {
			res, err := tr.Exec("INSERT OR IGNORE INTO tag (tag) VALUES (?)", tag)
			if err != nil {
				return fmt.Errorf("Failed to insert tag %s because %s", tag, err.Error())
			}
			//todo: does this work with OR IGNORE?
			tagId, err = res.LastInsertId()
			if err != nil {
				return fmt.Errorf("Failed to get inserted tag id because %s", err.Error())
			}
		}

		if _, err := tr.Exec("INSERT OR IGNORE INTO performer_tag (performer_id, tag_id) VALUES (?, ?)", performerID, tagId); err != nil {
			return fmt.Errorf("Failed to insert performer_tag relationship (perfomer: %d, tag: %s, tagId: %d) because %s", performerID, tag, tagId, err.Error())
		}
	}
	return nil
}

func (s *PerformerService) StorePerformerImages(tr *dbr.Tx, performerID int64, images map[string]string) error {
	for imageUseage, imageSrc := range images {
		if _, err := tr.Exec("INSERT OR IGNORE INTO performer_image (performer_id, usage, src) VALUES (?, ?, ?)", performerID, imageUseage, imageSrc); err != nil {
			return fmt.Errorf("Failed to add performer image (perfomer: %d, usage: %s, src: %s) because %s", performerID, imageUseage, imageSrc, err.Error())
		}
	}
	return nil
}
