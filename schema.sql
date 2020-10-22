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
