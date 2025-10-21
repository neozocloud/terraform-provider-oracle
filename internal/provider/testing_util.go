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
	createUserSQL  = "CREATE USER %s IDENTIFIED BY %s"
	dropUserSQL    = "DROP USER %s CASCADE"
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

func setupTestUser(t *testing.T, username, password string) {
	db, err := getTestDB()
	if err != nil {
		t.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf(createUserSQL, username, password))
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
}

func tearDownTestUser(t *testing.T, username string) {
	db, err := getTestDB()
	if err != nil {
		t.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf(dropUserSQL, username))
	if err != nil {
		t.Logf("Failed to drop test user %s: %v", username, err)
	}
}

func setupTestTableForOwner(t *testing.T, tableName, owner string) {
	setupTestUser(t, owner, "password")
	db, err := getTestDB()
	if err != nil {
		t.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("GRANT CREATE SESSION, CREATE TABLE TO %s", owner))
	if err != nil {
		t.Fatalf("Failed to grant privileges to owner: %v", err)
	}

	dbOwner, err := sql.Open("oracle", fmt.Sprintf("oracle://%s:%s@%s:%s/%s", owner, "password", os.Getenv("ORACLE_HOST"), os.Getenv("ORACLE_PORT"), os.Getenv("ORACLE_SERVICE")))
	if err != nil {
		t.Fatalf("Failed to connect to the database as owner: %v", err)
	}
	defer dbOwner.Close()

	_, err = dbOwner.Exec(fmt.Sprintf(createTableSQL, tableName))
	if err != nil {
		t.Fatalf("Failed to create test table for owner: %v", err)
	}
}

func tearDownTestTableForOwner(t *testing.T, tableName, owner string) {
	tearDownTestUser(t, owner)
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
