package pkg

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/libsql/libsql-client-go/libsql"
	_ "modernc.org/sqlite"
)

func CreateConnection() *sqlx.DB {
	conn, err := sqlx.Connect("sqlite", "file:./db.sqlite3?cache=shared&mode=rwc")
	if err != nil {
		log.Panicf("failed to connect to database: %v\n", err)
	}

	sql := `
        CREATE TABLE IF NOT EXISTS records (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            value TEXT NOT NULL,
            value_2 INTEGER NOT NULL,
            value_3 DATETIME NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );
    `

	if _, err := conn.Exec(sql); err != nil {
		log.Panicf("failed to create table: %v\n", err)
	}

	return conn
}
