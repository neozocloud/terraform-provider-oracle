// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oracle

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

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
// If requireNNE is true, the client verifies after connecting that the
// session is protected by Oracle Native Network Encryption and fails
// (closing the connection) if it is not. If requireNNE is false, no
// encryption check is performed.
//
// Parameters:
//
//		host: The hostname or IP address of the database server.
//		port: The port number on which the database is listening.
//		serviceName: The service name of the database.
//		user: The username to connect with.
//		password: The password for the specified user.
//	  requireNNE:  fail if the session is not encrypted via NNE
//
// Returns:
//
//	A new Oracle client or an error if the connection fails.
func NewClient(host, serviceName, user, password string, port int, requireNNE bool) (*Client, error) {
	urlOptions := map[string]string{}
	if requireNNE {
		// Fail during connection negotiation, i.e. before authentication,
		// if encryption and integrity cannot be established.
		urlOptions["encryption"] = "required"
		urlOptions["data integrity"] = "required"
	}
	dsn := goOra.BuildUrl(host, port, serviceName, user, password, urlOptions)
	db, err := sql.Open("oracle", dsn)
	if err != nil {
		return nil, fmt.Errorf("error creating database connection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	if requireNNE {
		// Defense in depth: confirm the established session really is encrypted.
		if err := verifyEncryption(ctx, db); err != nil {
			_ = db.Close()
			return nil, err
		}
	}

	return &Client{DB: db}, nil
}

// verifyEncryption checks that the current session uses an NNE encryption
// adapter, based on the network service banner of the session.
func verifyEncryption(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, `
		SELECT network_service_banner
		FROM   v$session_connect_info
		WHERE  sid = SYS_CONTEXT('USERENV', 'SID')`)
	if err != nil {
		return fmt.Errorf("error checking connection encryption: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var banner sql.NullString
		if err := rows.Scan(&banner); err != nil {
			return fmt.Errorf("error reading connection banner: %w", err)
		}
		b := strings.ToUpper(banner.String)
		if strings.Contains(b, "AES") && strings.Contains(b, "ENCRYPTION") {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error checking connection encryption: %w", err)
	}

	return fmt.Errorf("nne is enabled in the provider configuration, but the database session is not encrypted; " +
		"check the NATIVE_NETWORK_ENCRYPTION option group settings on the database (e.g. SQLNET.ENCRYPTION_SERVER and SQLNET.ENCRYPTION_TYPES_SERVER)")
}
