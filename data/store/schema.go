package store

import (
	"github.com/warmans/dbr"
)

func InitializeSchema(sess *dbr.Session) error {

	_, err := sess.Exec(`
		CREATE TABLE IF NOT EXISTS event (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			venue_id INTEGER,
			date DATETIME,
			type TEXT NULL,
			description TEXT NULL,
			deleted BOOLEAN DEFAULT 0
		);
	`)
	if err != nil {
		return err
	}
	_, err = sess.Exec(`
		CREATE TABLE IF NOT EXISTS venue (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			address TEXT NULL
		);
	`)
	if err != nil {
		return err
	}
	_, err = sess.Exec(`
		CREATE TABLE IF NOT EXISTS venue_extra (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			venue_id INTEGER,
			link TEXT,
			link_type TEXT NULL,
			link_description TEXT NULL
		);
	`)
	if err != nil {
		return err
	}
	_, err = sess.Exec(`
		CREATE TABLE IF NOT EXISTS performer (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			info TEXT,
			genre TEXT,
			home TEXT,
			img TEXT,
			listen_url TEXT
		);
	`)
	if err != nil {
		return err
	}
	_, err = sess.Exec(`
		CREATE TABLE IF NOT EXISTS performer_extra (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			performer_id INTEGER,
			link TEXT,
			link_type TEXT NULL,
			link_description TEXT NULL
		);
	`)
	if err != nil {
		return err
	}
	_, err = sess.Exec(`
		CREATE TABLE IF NOT EXISTS event_performer (
			event_id INTEGER,
			performer_id INTEGER,
			PRIMARY KEY (event_id, performer_id)
		);
	`)
	if err != nil {
		return err
	}
	_, err = sess.Exec(`
		CREATE TABLE IF NOT EXISTS user (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT,
			password TEXT,
			CONSTRAINT username_uniq UNIQUE (username)
		);
	`)
	if err != nil {
		return err
	}
	return err
}