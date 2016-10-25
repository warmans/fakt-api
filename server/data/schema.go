package data

import (
	"github.com/warmans/dbr"
)

func InitializeSchema(sess *dbr.Session) error {

	statements := []string{
		`CREATE TABLE IF NOT EXISTS event (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			venue_id INTEGER,
			date DATETIME,
			type TEXT NULL,
			description TEXT NULL,
			deleted BOOLEAN DEFAULT 0,
			source TEXT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS venue (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			address TEXT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS venue_extra (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			venue_id INTEGER,
			link TEXT,
			link_type TEXT NULL,
			link_description TEXT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS performer (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			info TEXT,
			genre TEXT,
			home TEXT,
			listen_url TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS performer_extra (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			performer_id INTEGER,
			link TEXT,
			link_type TEXT NULL,
			link_description TEXT NULL
		);`,
		//a user tag is just a "reaction" e.g. like, dislike
		`CREATE TABLE IF NOT EXISTS performer_user_tag (
			performer_id INTEGER,
			user_id INTEGER,
			tag TEXT,
			PRIMARY KEY (performer_id, user_id, tag)
		);`,
		//a tag is a random attribute to associate performers
		`CREATE TABLE IF NOT EXISTS performer_tag (
			performer_id INTEGER,
			tag_id INTEGER,
			PRIMARY KEY (performer_id, tag_id)
		);`,
		`CREATE TABLE IF NOT EXISTS performer_image (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			performer_id INTEGER,
			usage TEXT,
			src TEXT,
			CONSTRAINT image_uniq UNIQUE (performer_id, usage)
		);`,
		`CREATE TABLE IF NOT EXISTS event_performer (
			event_id INTEGER,
			performer_id INTEGER,
			PRIMARY KEY (event_id, performer_id)
		);`,
		`CREATE TABLE IF NOT EXISTS event_user_tag (
			event_id INTEGER,
			user_id INTEGER,
			tag TEXT,
			PRIMARY KEY (event_id, user_id, tag)
		);`,
		`CREATE TABLE IF NOT EXISTS user (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT,
			password TEXT,
			admin BOOLEAN DEFAULT 0,
			CONSTRAINT username_uniq UNIQUE (username)
		);`,
		`CREATE TABLE IF NOT EXISTS tag (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tag TEXT,
			CONSTRAINT tag_uniq UNIQUE (tag)
		);`,
	}

	for _, stmnt := range statements {
		_, err := sess.Exec(stmnt)
		if err != nil {
			return err
		}
	}
	return nil
}
