package venue

import (
	"database/sql"

	"github.com/warmans/dbr"
	"github.com/warmans/fakt-api/server/data/service/common"
)

type VenueFilter struct {
	VenueIDs []int  `json:"venues"`
	Name     string `json:"name"`
}

type VenueService struct {
	DB *dbr.Session
}

func (vs *VenueService) VenueMustExist(tr *dbr.Tx, venue *common.Venue) error {

	//get the venue ID if it exists
	if venue.ID == 0 {
		err := tr.QueryRow("SELECT id FROM venue WHERE name=?", venue.Name, venue.Address).Scan(&venue.ID)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
	}
	//still no venue ID... create it
	if venue.ID == 0 {
		res, err := tr.Exec(
			"INSERT INTO venue (name, address) VALUES (?, ?)",
			venue.Name,
			venue.Address,
		)
		if err != nil {
			return err
		}
		venue.ID, err = res.LastInsertId()
		if err != nil {
			return err
		}
	} else {
		_, err := tr.Exec(
			"UPDATE venue SET address=? WHERE id=?",
			venue.Address,
			venue.ID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *VenueService) FindVenues(filter *VenueFilter) ([]*common.Venue, error) {

	q := s.DB.Select("v.id", "v.name", "v.address").
		From("venue v").
		OrderBy("v.name")

	if len(filter.VenueIDs) > 0 {
		q.Where("v.id IN ?", filter.VenueIDs)
	}
	if filter.Name != "" {
		q.Where("v.name = ?", filter.Name)
	}

	venues := make([]*common.Venue, 0)
	if _, err := q.Load(&venues); err != nil && err != dbr.ErrNotFound {
		return nil, err
	}

	return venues, nil
}
