create table if not exists state
(
	id serial not null
		constraint state_pk
			primary key,
	user_id integer not null
		constraint state_user_id_fk
			references users,
	timestamp timestamp not null,
	data jsonb
);

alter table state owner to doadmin;

create table if not exists tesla_auth
(
	id serial not null
		constraint tesla_auth_pk
			primary key,
	user_id integer not null
		constraint tesla_auth_users_id_fk
			references users,
	access_token text not null,
	token_type text not null,
	expires_in integer not null,
	refresh_token text not null,
	created_at integer not null
);

alter table tesla_auth owner to doadmin;

create unique index if not exists tesla_auth_id_uindex
	on tesla_auth (id);

create unique index if not exists tesla_auth_user_id_uindex
	on tesla_auth (user_id);

create table if not exists users
(
	id serial not null
		constraint users_pk
			primary key,
	email text not null,
	password text not null
);

alter table users owner to doadmin;

create unique index if not exists users_email_uindex
	on users (email);

create unique index if not exists users_id_uindex
	on users (id);
