package auth

const (
	tb_user_sql = `
CREATE TABLE IF NOT EXISTS user_info (
	id TEXT NOT NULL PRIMARY KEY,
	created_at DATETIME NOT NULL DEFAULT (datetime('now', 'localtime')),
	updated_at DATETIME NOT NULL DEFAULT (datetime('now', 'localtime')),
	passwd TEXT NOT NULL,
	nick_name TEXT NOT NULL DEFAULT '',
	kind INT NOT NULL DEFAULT 2, -- 1, admin; 2, users.
	memo TEXT NOT NULL DEFAULT ''
);`
)
