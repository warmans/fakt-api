package venue

import (
	"database/sql"

	"net/http"
	"strconv"
	"strings"

	"github.com/warmans/dbr"
	"github.com/warmans/fakt-api/server/data/service/common"
)

const DefaultPageSize = 50

func VenueFilterFromRequest(r *http.Request) *VenueFilter {
	f := &VenueFilter{}
	f.Populate(r)
	return f
}

type VenueFilter struct {
	VenueIDs []int  `json:"venues"`
	Name     string `json:"name"`
	Page     int64  `json:"page"`
	PageSize int64  `json:"page_size"`
	SortCol  string `json:"sort_col"`
	SortAsc  bool   `json:"sort_asc"`
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

	validSortColumns := map[string]bool{"name": true, "activity": true}
	if sortCol := r.Form.Get("sort_col"); sortCol != "" {
		if validSortColumns[sortCol] {
			f.SortCol = sortCol
		}
	}

	f.SortAsc = true
	if sortDir := r.Form.Get("sort_asc"); sortDir == "false" {
		f.SortAsc = false
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
		page = 1
	}

	q := s.DB.Select("id", "name", "address", "COALESCE(activity, 0) AS activity").
		From("venue").
		Limit(uint64(filter.PageSize)).
		Offset(uint64((filter.PageSize * page) - filter.PageSize))

	if len(filter.VenueIDs) > 0 {
		q.Where("id IN ?", filter.VenueIDs)
	}
	if filter.Name != "" {
		q.Where("name = ?", filter.Name)
	}

	if filter.SortCol != "" {
		q.OrderDir(filter.SortCol, filter.SortAsc)
	}

	venues := make([]*common.Venue, 0)
	if _, err := q.Load(&venues); err != nil && err != dbr.ErrNotFound {
		return nil, err
	}

	return venues, nil
}
