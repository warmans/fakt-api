package tag

import (
	"github.com/warmans/dbr"
	"log"
	"database/sql"
	"strings"
)

type TagService struct {
	DB *dbr.Session
}

func (ts *TagService) StorePerformerTags(tr *dbr.Tx, performerID int64, tags []string) {
	//handle tags
	if _, err := tr.Exec("DELETE FROM performer_tag WHERE performer_id = ?", performerID); err != nil {
		log.Printf("Failed to delete existing performer_tag relationships (perfomer: %d) because %s", performerID, err.Error())
	}

	//todo: move all this into new tag service
	for _, tag := range tags {

		var tagId int64
		tag = strings.ToLower(tag)

		err := ts.DB.QueryRow("SELECT id FROM tag WHERE tag = ?", tag).Scan(&tagId)
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

		if _, err := tr.Exec("INSERT OR IGNORE INTO performer_tag (performer_id, tag_id) VALUES (?, ?)", performerID, tagId); err != nil {
			log.Printf("Failed to insert performer_tag relationship (perfomer: %d, tag: %s, tagId: %d) because %s", performerID, tag, tagId, err.Error())
			continue
		}
	}
}
