package handler
import (
	"net/http"
	"github.com/warmans/stressfaktor-api/api/common"
	"github.com/warmans/stressfaktor-api/entity"
	"log"
	"strings"
	"strconv"
)

func NewVenueHandler(eventStore *entity.EventStore) http.Handler {
	return &VenueHandler{eventStore: eventStore}
}

type VenueHandler struct {
	eventStore *entity.EventStore
}

func (h *VenueHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	r.ParseForm()

	//query to filter
	filter := &entity.VenueFilter{VenueIDs: make([]int, 0)}
	if venue := r.Form.Get("venue"); venue != "" {
		for _, idStr := range strings.Split(venue, ",") {
			if idInt, err := strconv.Atoi(idStr); err == nil {
				filter.VenueIDs = append(filter.VenueIDs, idInt)
			}
		}
	}

	venues, err := h.eventStore.FindVenues(filter)
	if err != nil {
		log.Print(err.Error())
		common.SendResponse(rw, &common.Response{Status: http.StatusInternalServerError, Payload: nil})
		return
	}

	common.SendResponse(rw, &common.Response{Status: http.StatusOK, Payload: venues})
}
