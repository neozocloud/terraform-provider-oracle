// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oracle

import (
	"fmt"
)

// User represents an Oracle database user.
// It contains the information needed to create, modify, or manage a user.
type User struct {
	Username              string // The name of the user.
	Password              string // The user's password. Only used for authentication type "password".
	DefaultTablespace     string // The default tablespace for the user.
	DefaultTempTablespace string // The default temporary tablespace for the user.
	Profile               string // The user's profile.
	AuthenticationType    string // The authentication type, e.g., "password", "external", "global".
	State                 string // The desired state of the user account, e.g., "locked", "unlocked".
}

// CreateUser creates a new user in the Oracle database.
//
// Parameters:
//
//	user: A User struct containing the details of the user to be created.
//
// Returns:
//
//	An error if the user creation fails.
func (c *Client) CreateUser(user User) error {
	sql := fmt.Sprintf("CREATE USER %s", user.Username)

	switch user.AuthenticationType {
	case "password":
		sql += fmt.Sprintf(" IDENTIFIED BY \"%s\"", user.Password)
	case "external":
		sql += " IDENTIFIED EXTERNALLY"
	case "global":
		sql += " IDENTIFIED GLOBALLY"
	}

	if user.DefaultTablespace != "" {
		sql += fmt.Sprintf(" DEFAULT TABLESPACE %s", user.DefaultTablespace)
	}

	if user.DefaultTempTablespace != "" {
		sql += fmt.Sprintf(" TEMPORARY TABLESPACE %s", user.DefaultTempTablespace)
	}

	if user.Profile != "" {
		sql += fmt.Sprintf(" PROFILE %s", user.Profile)
	}

	if user.State == "locked" {
		sql += " ACCOUNT LOCK"
	}

	_, err := c.DB.Exec(sql)
	return err
}

// ModifyUser modifies an existing user in the Oracle database.
//
// Parameters:
//
//	user: A User struct containing the details of the user to be modified.
//
// Returns:
//
//	An error if the user modification fails.
func (c *Client) ModifyUser(user User) error {
	sql := fmt.Sprintf("ALTER USER %s", user.Username)

	if user.Password != "" {
		sql += fmt.Sprintf(" IDENTIFIED BY \"%s\"", user.Password)
	}

	if user.DefaultTablespace != "" {
		sql += fmt.Sprintf(" DEFAULT TABLESPACE %s", user.DefaultTablespace)
	}

	if user.DefaultTempTablespace != "" {
		sql += fmt.Sprintf(" TEMPORARY TABLESPACE %s", user.DefaultTempTablespace)
	}

	if user.Profile != "" {
		sql += fmt.Sprintf(" PROFILE %s", user.Profile)
	}

	switch user.State {
	case "locked":
		sql += " ACCOUNT LOCK"
	case "unlocked":
		sql += " ACCOUNT UNLOCK"
	}

	_, err := c.DB.Exec(sql)
	return err
}

// DropUser drops a user from the Oracle database.
//
// Parameters:
//
//	username: The name of the user to be dropped.
//
// Returns:
//
//	An error if the user drop fails.
func (c *Client) DropUser(username string) error {
	sql := fmt.Sprintf("DROP USER %s CASCADE", username)
	_, err := c.DB.Exec(sql)
	return err
}

// UserExists checks if a user exists in the database.
//
// Parameters:
//
//	username: The name of the user to check.
//
// Returns:
//
//	A boolean indicating whether the user exists, and an error if the check fails.
func (c *Client) UserExists(username string) (bool, error) {
	var count int
	sql := "SELECT COUNT(*) FROM dba_users WHERE username = UPPER(:1)"
	err := c.DB.QueryRow(sql, username).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ReadUser reads a user's details from the database.
//
// Parameters:
//
//	username: The name of the user to read.
//
// Returns:
//
//	A User struct containing the user's details, and an error if the read fails.
func (c *Client) ReadUser(username string) (*User, error) {
	user := &User{}
	sql := "SELECT username, default_tablespace, temporary_tablespace, profile, authentication_type, account_status FROM dba_users WHERE username = UPPER(:1)"
	err := c.DB.QueryRow(sql, username).Scan(&user.Username, &user.DefaultTablespace, &user.DefaultTempTablespace, &user.Profile, &user.AuthenticationType, &user.State)
	if err != nil {
		return nil, err
	}
	return user, nil
}
