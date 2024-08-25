package entx

import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	"modernc.org/sqlite"
)

// sqliteDriver is a wrapper around the sqlite to register as sqlite3 driver
type sqliteDriver struct {
	*sqlite.Driver
}

// Open opens a new connection to the database with foreign keys enabled.
func (d sqliteDriver) Open(name string) (driver.Conn, error) {
	conn, err := d.Driver.Open(name)
	if err != nil {
		return conn, err
	}

	c := conn.(interface {
		Exec(stmt string, args []driver.Value) (driver.Result, error)
	})

	if _, err := c.Exec("PRAGMA foreign_keys = on;", nil); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to enable enable foreign keys: %w", err)
	}

	return conn, nil
}

// init registers the sqlite3 driver
func init() {
	sql.Register("sqlite3", sqliteDriver{Driver: &sqlite.Driver{}})
}
