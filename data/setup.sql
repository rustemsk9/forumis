DROP TABLE IF EXISTS dislikes;
DROP TABLE IF EXISTS likedposts;
DROP TABLE IF EXISTS threaddislikes;
DROP TABLE IF EXISTS threadlikes;
DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS threads;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;


CREATE TABLE users (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  uuid       varchar(64) not null unique,
  name       varchar(64),
  email      varchar(64) not null unique,
  password   varchar(128) not null,
  created_at timestamp not null
);

CREATE TABLE sessions (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  uuid          varchar(64) not null unique,
  email         varchar(64),
  user_id       integer references users(id),
  created_at    timestamp not null,
  cookie_string varchar(255),
  active_last   integer default 0
);

CREATE TABLE threads (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  uuid       varchar(64) not null unique,
  topic      text,
  user_id    integer references users(id),
  created_at timestamp not null,
  category1  varchar(255) default '',
  category2  varchar(255) default ''
);

CREATE TABLE posts (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  uuid       varchar(64) not null unique,
  body       text,
  user_id    integer references users(id),
  thread_id  integer references threads(id),
  created_at timestamp not null
);

CREATE TABLE threadlikes (
  id        INTEGER PRIMARY KEY AUTOINCREMENT,
  type      varchar(50),
  user_id   integer references users(id),
  thread_id integer references threads(id)
);

CREATE TABLE threaddislikes (
  id        INTEGER PRIMARY KEY AUTOINCREMENT,
  type      varchar(50),
  user_id   integer references users(id),
  thread_id integer references threads(id)
);

CREATE TABLE likedposts (
  id      INTEGER PRIMARY KEY AUTOINCREMENT,
  type    varchar(50),
  user_id integer references users(id),
  post_id integer references posts(id)
);

CREATE TABLE dislikes (
  id      INTEGER PRIMARY KEY AUTOINCREMENT,
  type    varchar(50),
  user_id integer references users(id),
  post_id integer references posts(id)
);