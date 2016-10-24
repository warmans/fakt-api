package action

import (
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/urfave/cli"
	"github.com/warmans/dbr"
	"github.com/warmans/fakt-api/server/data/media"
	"github.com/warmans/fakt-api/server/data/service/performer"
)

var PerformerImgUnHotlink = func(c *cli.Context) error {

	db, err := dbr.Open("sqlite3", c.String("db.path"), nil)
	if err != nil {
		return fmt.Errorf("failed to open DB: %s", err.Error())
	}
	defer db.Close()

	dbSession := db.NewSession(nil)

	performerService := performer.PerformerService{}

	mirror := media.NewImageMirror(c.String("static.path"))

	rows, err := db.Query("SELECT id, img FROM performer WHERE img LIKE 'http%'")
	if err != nil {
		return fmt.Errorf("failed find performers because %s", err.Error())
	}

	imagesToProcess := make(map[int64]string)
	for rows.Next() {
		var performerID int64
		var oldImage string

		if err := rows.Scan(&performerID, &oldImage); err != nil {
			return fmt.Errorf("Failed scanning performers: %s", err.Error())
		}
		imagesToProcess[performerID] = oldImage
	}
	if err := rows.Close(); err != nil {
		return err
	}

	for performerID, uri := range imagesToProcess {
		images, err := mirror.Mirror(uri, fmt.Sprintf("%d", performerID))
		if err != nil {
			return err
		}

		tx, err := dbSession.Begin()
		if err != nil {
			return fmt.Errorf("Failed to start DB transaction: %s", err.Error())
		}
		if err := performerService.StorePerformerImages(tx, performerID, images); err != nil {
			return fmt.Errorf("Failed to store images for: %s", err.Error())
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("Failed commit image to db: %s", err.Error())
		}
		fmt.Printf("Mirrored %s\n", uri)
	}

	fmt.Println("Finished!")
	return nil
}
