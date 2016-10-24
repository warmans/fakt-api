package main

import (
	"os"

	"github.com/urfave/cli"
	"github.com/warmans/fakt-api/cmd/maint/action"
)

func main() {
	app := cli.NewApp()
	app.Commands = []cli.Command{
		{
			Name:    "performer",
			Aliases: []string{"p"},
			Usage:   "performer tasks",
			Subcommands: []cli.Command{
				{
					Name:   "unhotlink-images",
					Usage:  "convert any URLs in the performer.img field to a locally hosted path",
					Action: action.PerformerImgUnHotlink,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "db.path",
							Value: "db.sqlite3",
							Usage: "path to database file",
						},
						cli.StringFlag{
							Name:  "static.path",
							Value: "static",
							Usage: "path to stadic files dir",
						},
					},
				},
			},
		},
	}
	app.Run(os.Args)
}
