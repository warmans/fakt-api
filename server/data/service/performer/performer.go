package performer

import (
	"database/sql"
	"log"

	"github.com/warmans/dbr"
	"github.com/warmans/fakt-api/server/data/service/common"
	"github.com/warmans/fakt-api/server/data/service/utag"
	"strings"
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

	//handle tags
	if _, err := tr.Exec("DELETE FROM performer_tag WHERE performer_id = ?", performer.ID); err != nil {
		log.Printf("Failed to delete existing performer_tag relationships (perfomer: %d) because %s", performer.ID, err.Error())
	}

	//todo: move all this into new tag service
	for _, tag := range performer.Tags {

		var tagId int64
		tag = strings.ToLower(tag)

		err := tr.QueryRow("SELECT id FROM tag WHERE tag = ?", tag).Scan(&tagId)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("Failed to find tag id for %s because %s", tag, err.Error())
			continue
		}
		if tagId == 0 {
			res, err := tr.Exec("INSERT OR IGNORE INTO tag (tag) VALUES (?)", tag)
			if err != nil {
				log.Printf("Failed to insert tag %s because %s", tag, err.Error())
				continue
			}
			//todo: does this work with OR IGNORE?
			tagId, err = res.LastInsertId()
			if err != nil {
				log.Printf("Failed to get inserted tag id because %s", err.Error())
				continue
			}
		}

		if _, err := tr.Exec("INSERT OR IGNORE INTO performer_tag (performer_id, tag_id) VALUES (?, ?)", performer.ID, tagId); err != nil {
			log.Printf("Failed to insert performer_tag relationship (perfomer: %d, tag: %s, tagId: %d) because %s", performer.ID, tag, tagId, err.Error())
			continue
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
			log.Print("Failed to insert performer_extra: %s", err.Error())
			continue
		}
	}

	return nil
}

func (s *PerformerService) FindPerformers(filter *PerformerFilter) ([]*common.Performer, error) {

	q := s.DB.Select("id", "name", "info", "genre", "home", "img", "listen_url").
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

		//append the tags
		tags, err := s.UTagService.FindPerformerUTags(performer.ID, &common.UTagsFilter{})
		if err != nil {
			return nil, err
		}
		performers[k].UTags = tags
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
