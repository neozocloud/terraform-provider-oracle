// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/sijms/go-ora/v2"
)

const (
	createTableSQL = "CREATE TABLE %s (id NUMBER)"
	dropTableSQL   = "DROP TABLE %s"
)

func setupTestTable(t *testing.T, tableName string) {
	db, err := getTestDB()
	if err != nil {
		t.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf(createTableSQL, tableName))
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}
}

func tearDownTestTable(t *testing.T, tableName string) {
	db, err := getTestDB()
	if err != nil {
		t.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf(dropTableSQL, tableName))
	if err != nil {
		t.Logf("Failed to drop test table %s: %v", tableName, err)
	}
}

func getTestDB() (*sql.DB, error) {
	return sql.Open("oracle", getDBConnectionString())
}

func getDBConnectionString() string {
	return fmt.Sprintf("oracle://%s:%s@%s:%s/%s",
		os.Getenv("ORACLE_USERNAME"),
		os.Getenv("ORACLE_PASSWORD"),
		os.Getenv("ORACLE_HOST"),
		os.Getenv("ORACLE_PORT"),
		os.Getenv("ORACLE_SERVICE"))
}
