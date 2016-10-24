package performer

import (
	"database/sql"

	"fmt"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/warmans/dbr"
	"github.com/warmans/fakt-api/server/data/service/common"
	"github.com/warmans/fakt-api/server/data/service/utag"
)

type PerformerFilter struct {
	PerformerID []int  `json:"performers"`
	Name        string `json:"name"`
	Genre       string `json:"name"`
	Home        string `json:"name"`
}

type PerformerService struct {
	DB          *dbr.Session
	UTagService *utag.UTagService
	Logger      log.Logger
}

func (s *PerformerService) FindPerformers(filter *PerformerFilter) ([]*common.Performer, error) {

	q := s.DB.
		Select("id", "name", "info", "genre", "home", "img", "listen_url").
		From("performer p").
		OrderBy("p.name")

	if len(filter.PerformerID) > 0 {
		q.Where("p.id IN ?", filter.PerformerID)
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
	performers := make([]*common.Performer, 0)
	if _, err := q.Load(&performers); err != nil && err != dbr.ErrNotFound {
		return nil, err
	}

	for k, performer := range performers {
		links, err := s.FindPerformerLinks(performer.ID)
		if err != nil {
			return nil, err
		}
		performers[k].Links = links

		//append the utags
		tags, err := s.UTagService.FindPerformerUTags(performer.ID, &common.UTagsFilter{})
		if err != nil {
			return nil, err
		}
		performers[k].UTags = tags

		//append normal tags
		if performers[k].Tags, err = s.FindPerformerTags(performer.ID); err != nil {
			return nil, err
		}

		//images
		if performers[k].Images, err = s.FindPerformerImages(performer.ID); err != nil {
			return nil, err
		}
	}

	return performers, nil
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

func (ts *PerformerService) FindPerformerTags(performerID int64) ([]string, error) {

	tags := []string{}

	res, err := ts.DB.Query("SELECT t.tag FROM performer_tag pt LEFT JOIN tag t ON pt.tag_id = t.id WHERE pt.performer_id = ?", performerID)
	if err != nil {
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

func (ts *PerformerService) FindPerformerImages(performerID int64) (map[string]string, error) {

	tags := make(map[string]string)

	res, err := ts.DB.Query("SELECT pi.usage, pi.src FROM performer_image pi WHERE pi.performer_id = ?", performerID)
	if err != nil {
		return tags, fmt.Errorf("Failed to fetch tags at query because %s", err.Error())
	}

	for res.Next() {
		var usage, src string
		if err := res.Scan(&usage, &src); err != nil {
			return tags, err
		}
		tags[usage] = src
	}

	return tags, nil
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
			"INSERT INTO performer (name, info, genre, home, img, listen_url) VALUES (?, ?, ?, ?, ?, ?)",
			performer.Name,
			performer.Info,
			performer.Genre,
			performer.Home,
			performer.Img,
			performer.ListenURL,
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
			"UPDATE performer SET info=?, home=?, img=?, listen_url=? WHERE id=?",
			performer.Info,
			performer.Home,
			performer.Img,
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
