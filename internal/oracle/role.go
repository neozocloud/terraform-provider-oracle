// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oracle

import "fmt"

// Role represents an Oracle database role.
type Role struct {
	Name string // The name of the role.
}

// CreateRole creates a new role in the Oracle database.
//
// Parameters:
//
//	role: A Role struct containing the details of the role to be created.
//
// Returns:
//
//	An error if the role creation fails.
func (c *Client) CreateRole(role Role) error {
	sql := fmt.Sprintf("CREATE ROLE %s", role.Name)
	_, err := c.DB.Exec(sql)
	return err
}

// DropRole drops a role from the Oracle database.
//
// Parameters:
//
//	roleName: The name of the role to be dropped.
//
// Returns:
//
//	An error if the role drop fails.
func (c *Client) DropRole(roleName string) error {
	sql := fmt.Sprintf("DROP ROLE %s", roleName)
	_, err := c.DB.Exec(sql)
	return err
}

// RoleExists checks if a role exists in the database.
//
// Parameters:
//
//	roleName: The name of the role to check.
//
// Returns:
//
//	A boolean indicating whether the role exists, and an error if the check fails.
func (c *Client) RoleExists(roleName string) (bool, error) {
	var count int
	sql := "SELECT COUNT(*) FROM dba_roles WHERE role = UPPER(:1)"
	err := c.DB.QueryRow(sql, roleName).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ReadRole reads a role's details from the database.
//
// Parameters:
//
//	roleName: The name of the role to read.
//
// Returns:
//
//	A Role struct containing the role's details and an error if the read fails.
func (c *Client) ReadRole(roleName string) (*Role, error) {
	role := &Role{}
	sql := "SELECT role FROM dba_roles WHERE role = UPPER(:1)"
	err := c.DB.QueryRow(sql, roleName).Scan(&role.Name)
	if err != nil {
		return nil, err
	}
	return role, nil
}
