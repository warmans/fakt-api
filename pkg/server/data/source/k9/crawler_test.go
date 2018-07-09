package k9

import (
	"testing"
	"github.com/warmans/fakt-api/pkg/server/data/source"
)

func TestDateFromTitle(t *testing.T) {

	examples := []struct {
		RawDateString      string
		ExpectedParsedDate string
		ExpectError        bool
	}{
		{RawDateString: "Freitag, 02.09.2016 – ab 19.00 h – Größenwahn und Leichtsinn", ExpectedParsedDate: "02-09-2016 19:00", ExpectError: false},
		{RawDateString: "Donnerstag, 15.09.2016, 18.30 h – 21.30 h, Größenwahn und Leichtsinn", ExpectedParsedDate: "15-09-2016 18:30", ExpectError: false},
		{RawDateString: "Donnerstag, 22.09.2016, 18.30 h – 21.30 h, Größenwahn und Leichtsinn", ExpectedParsedDate: "22-09-2016 18:30", ExpectError: false},
		{RawDateString: "Sonntag, 25.09.2016 – 20.00 h – Größenwahn", ExpectedParsedDate: "25-09-2016 20:00", ExpectError: false},
	}


	tz, err := source.MustMakeTimeLocation("Europe/Berlin")
	if err != nil {
		t.Fatal("Unexpected error creating timzone: %s", err.Error())
	}
	for _, ex := range examples {
		date, err := dateFromTitle(ex.RawDateString, tz)
		if (err == nil) == ex.ExpectError {
			t.Errorf("Unexpected error parsing %s: %s", ex.RawDateString, err.Error())
			return
		}
		if formatted := date.Format("02-01-2006 15:04"); formatted != ex.ExpectedParsedDate {
			t.Errorf("Date did not parse to correct value: %s (expected %s)", formatted, ex.ExpectedParsedDate)
			return
		}
	}
}
