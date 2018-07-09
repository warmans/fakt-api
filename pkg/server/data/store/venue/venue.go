package venue

import (
	"database/sql"
	"net/http"

	"github.com/warmans/dbr"
	"github.com/warmans/fakt-api/pkg/server/data/store/common"
)

func FilterFromRequest(r *http.Request) *Filter {
	f := &Filter{}
	f.Populate(r)
	return f
}

type Filter struct {
	common.Filter

	Name    string `json:"name"`
	SortCol string `json:"sort_col"`
	SortAsc bool   `json:"sort_asc"`
}

func (f *Filter) Populate(r *http.Request) {

	f.Filter.Populate(r)

	//query to filter
	f.Name = r.Form.Get("name")

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

type Store struct {
	DB *dbr.Session
}

func (s *Store) VenueMustExist(tr *dbr.Tx, venue *common.Venue) error {

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

func (s *Store) FindVenues(filter *Filter) ([]*common.Venue, error) {

	//if no page is specified assume the first page
	page := filter.Page
	if page == 0 {
		page = 1
	}

	q := s.DB.Select("id", "name", "address", "COALESCE(activity, 0) AS activity").
		From("venue").
		Limit(uint64(filter.PageSize))
	if filter.PageSize != 0 {
		q.Limit(uint64(filter.PageSize)).Offset(uint64((filter.PageSize * page) - filter.PageSize))
	}

	if len(filter.IDs) > 0 {
		q.Where("id IN ?", filter.IDs)
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
