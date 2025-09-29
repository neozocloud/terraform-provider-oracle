// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oracle

import (
	"database/sql"
	"fmt"

	goOra "github.com/sijms/go-ora/v2"
)

// Client is a client for interacting with an Oracle database.
// It encapsulates the database connection and provides methods for managing users, roles, grants, and directories.

type Client struct {
	DB *sql.DB
}

// NewClient creates and returns a new Oracle client.
// It takes the database connection details as parameters and establishes a connection.
// It also pings the database to verify that the connection is active.
//
// Parameters:
//
//	host: The hostname or IP address of the database server.
//	port: The port number on which the database is listening.
//	serviceName: The service name of the database.
//	user: The username to connect with.
//	password: The password for the specified user.
//
// Returns:
//
//	A new Oracle client or an error if the connection fails.
func NewClient(host, serviceName, user, password string, port int) (*Client, error) {
	dsn := goOra.BuildUrl(host, port, serviceName, user, password, nil)
	db, err := sql.Open("oracle", dsn)
	if err != nil {
		return nil, fmt.Errorf("error creating database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	return &Client{DB: db}, nil
}
