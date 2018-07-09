-- +migrate Up

CREATE TABLE IF NOT EXISTS event (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  venue_id INTEGER,
  date DATETIME,
  type TEXT NULL,
  description TEXT NULL,
  deleted BOOLEAN DEFAULT 0,
  source TEXT NULL
);

CREATE TABLE IF NOT EXISTS venue (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT,
  address TEXT NULL,
  activity REAL
);

CREATE TABLE IF NOT EXISTS venue_extra (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  venue_id INTEGER,
  link TEXT,
  link_type TEXT NULL,
  link_description TEXT NULL
);

CREATE TABLE IF NOT EXISTS performer (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT,
  info TEXT,
  genre TEXT,
  home TEXT,
  listen_url TEXT,
  activity REAL,
  popularity REAL,
  embed_url TEXT
);

CREATE TABLE IF NOT EXISTS performer_extra (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  performer_id INTEGER,
  link TEXT,
  link_type TEXT NULL,
  link_description TEXT NULL
);

--a tag is a random attribute to associate performers
CREATE TABLE IF NOT EXISTS performer_tag (
  performer_id INTEGER,
  tag_id INTEGER,
  PRIMARY KEY (performer_id, tag_id)
);

CREATE TABLE IF NOT EXISTS performer_image (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  performer_id INTEGER,
  usage TEXT,
  src TEXT,
  CONSTRAINT image_uniq UNIQUE (performer_id, usage)
);

CREATE TABLE IF NOT EXISTS event_performer (
  event_id INTEGER,
  performer_id INTEGER,
  PRIMARY KEY (event_id, performer_id)
);

CREATE TABLE IF NOT EXISTS event_tag (
  event_id INTEGER,
  tag_id INTEGER,
  PRIMARY KEY (event_id, tag_id)
);

CREATE TABLE IF NOT EXISTS tag (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  tag TEXT,
  CONSTRAINT tag_uniq UNIQUE (tag)
);

-- +migrate Down

DROP TABLE tag;
DROP TABLE event_tag;
DROP TABLE event_performer;
DROP TABLE performer_image;
DROP TABLE performer_tag;
DROP TABLE performer_extra;
DROP TABLE performer;
DROP TABLE venue_extra;
DROP TABLE venue;
DROP TABLE event;