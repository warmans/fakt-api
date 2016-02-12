package data

import "database/sql"

func InitializeSchema(conn *sql.DB) error {

	_, err := conn.Exec(`
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
	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS venue (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			address TEXT NULL
		);
	`)
	if err != nil {
		return err
	}
	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS performer (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			genre TEXT,
			home TEXT,
			listen_url TEXT
		);
	`)
	if err != nil {
		return err
	}
	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS event_performer (
			event_id INTEGER,
			performer_id INTEGER,
			PRIMARY KEY (event_id, performer_id)
		);
	`)
	if err != nil {
		return err
	}

	return err
}