// Copyright (C) 2019 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build !sqlite_omit_load_extension

package sqlite3

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"sync"
	"testing"
)

// registerIfNeeded registers driver as a sql Driver if it does not already exist.
// The allows tests to run multiple times with -count.
func registerIfNeeded(name string, driver driver.Driver) {
	for _, registeredDriver := range sql.Drivers() {
		if name == registeredDriver {
			// already registered: do nothing
			return
		}
	}
	sql.Register(name, driver)
}

var uniqueCountMu sync.Mutex
var uniqueCount int

// registerNew generates a new driver name and registers the driver with it.
// The allows tests to run multiple times with -count.
// TODO: Replace with sql.OpenDB once implemented.
func registerNew(namePrefix string, driver driver.Driver) string {
	uniqueCountMu.Lock()
	count := uniqueCount
	uniqueCount++
	uniqueCountMu.Unlock()

	name := fmt.Sprintf("%s_%03d", namePrefix, count)
	sql.Register(name, driver)
	return name
}

func TestExtensionsError(t *testing.T) {
	const driverName = "sqlite3_TestExtensionsError"
	registerIfNeeded(driverName, &SQLiteDriver{
		Extensions: []string{
			"foobar",
		},
	},
	)

	db, err := sql.Open(driverName, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err == nil {
		t.Fatal("expected error loading non-existent extension")
	}

	if err.Error() == "not an error" {
		t.Fatal("expected error from sqlite3_enable_load_extension to be returned")
	}
}

func TestLoadExtensionError(t *testing.T) {
	const driverName = "sqlite3_TestLoadExtensionError"
	registerIfNeeded(driverName, &SQLiteDriver{
		ConnectHook: func(c *SQLiteConn) error {
			return c.LoadExtension("foobar", "")
		},
	},
	)

	db, err := sql.Open(driverName, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err == nil {
		t.Fatal("expected error loading non-existent extension")
	}

	if err.Error() == "not an error" {
		t.Fatal("expected error from sqlite3_enable_load_extension to be returned")
	}
}
