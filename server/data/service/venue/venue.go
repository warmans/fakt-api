package venue

import (
	"database/sql"

	"github.com/warmans/dbr"
	"github.com/warmans/fakt-api/server/data/service/common"
	"net/http"
	"strconv"
	"strings"
)

const PageSize = 50

func VenueFilterFromRequest(r *http.Request) *VenueFilter {
	f := &VenueFilter{}
	f.Populate(r)
	return f
}

type VenueFilter struct {
	VenueIDs []int  `json:"venues"`
	Name     string `json:"name"`
	Page     int64
}

func (f *VenueFilter) Populate(r *http.Request) {

	//query to filter
	f.VenueIDs = make([]int, 0)
	f.Name = r.Form.Get("name")

	if ven := r.Form.Get("ids"); ven != "" {
		for _, idStr := range strings.Split(ven, ",") {
			if idInt, err := strconv.Atoi(idStr); err == nil {
				f.VenueIDs = append(f.VenueIDs, idInt)
			}
		}
	}

	if page := r.Form.Get("page"); page  != "" {
		if pageInt, err := strconv.Atoi(page); err == nil {
			f.Page = int64(pageInt)
		}
	}
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

	//if no page is specified assume the first page
	page := filter.Page
	if page == 0 {
		page = 1;
	}

	q := s.DB.Select("v.id", "v.name", "v.address", "v.activity").
		From("venue v").
		OrderBy("v.name").
		Limit(uint64(PageSize)).
		Offset(uint64((PageSize * page) - PageSize))

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
