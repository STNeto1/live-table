package pkg

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/libsql/libsql-client-go/libsql"
	"github.com/rs/zerolog"
	sqldblogger "github.com/simukti/sqldb-logger"
	"github.com/simukti/sqldb-logger/logadapter/zerologadapter"
	_ "modernc.org/sqlite"
)

type Record struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Value     string    `db:"value"`
	Value2    int64     `db:"value_2"`
	Value3    time.Time `db:"value_3"`
	CreatedAt time.Time `db:"created_at"`
}

func CreateConnection(withDebug bool) *sqlx.DB {
	driverName := "sqlite"
	dsn := "file:./db.sqlite3?cache=shared&mode=rwc"

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		log.Panicf("failed to open connection: %v\n", err)
	}

	if withDebug {
		// initiate zerolog
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		zlogger := zerolog.New(os.Stdout).With().Timestamp().Logger()
		// prepare logger
		loggerOptions := []sqldblogger.Option{
			sqldblogger.WithSQLQueryFieldname("sql"),
			sqldblogger.WithWrapResult(false),
			sqldblogger.WithExecerLevel(sqldblogger.LevelDebug),
			sqldblogger.WithQueryerLevel(sqldblogger.LevelDebug),
			sqldblogger.WithPreparerLevel(sqldblogger.LevelDebug),
		}
		// wrap *sql.DB to transparent logger
		db = sqldblogger.OpenDriver(dsn, db.Driver(), zerologadapter.New(zlogger), loggerOptions...)
	}
	// pass it sqlx
	sqlxDB := sqlx.NewDb(db, driverName)

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

	if _, err := sqlxDB.Exec(sql); err != nil {
		log.Panicf("failed to create table: %v\n", err)
	}

	return sqlxDB
}
