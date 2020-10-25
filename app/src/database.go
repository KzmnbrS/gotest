package main

import (
	"github.com/jmoiron/sqlx"
	"log"
)

const init_sql = `
create table image
(
	id integer
		constraint image_pk
			primary key autoincrement,
	parent integer
		constraint image_child_parent_fk
			references image
				on delete cascade,
	basename text not null,
	uri text not null,
	width integer not null,
	height integer not null
);

create unique index image_uri_unique
	on image (uri);
`

func getDatabase(path string) *sqlx.DB {
	DBWasCreated := ensurePath(path, false)
	db, err := sqlx.Connect(`sqlite3`, path)
	if err != nil {
		log.Fatalf(`database can't be opened: %v`, err)
	}

	if DBWasCreated {
		if _, err := db.Exec(init_sql); err != nil {
			log.Fatalf(`database initialization failed: %v`, err)
		}
	}

	if _, err := db.Exec(`pragma foreign_keys = on`); err != nil {
		log.Fatalf(`database didn't PONG: %v`, err)
	}

	return db
}
