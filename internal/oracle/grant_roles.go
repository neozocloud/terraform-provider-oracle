// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oracle

import (
	"fmt"
	"strings"
)

// GrantRole represents a role to be granted to a user.
type GrantRole struct {
	Principal  string   // The user to whom the roles should be granted.
	Roles      []string // A list of roles to grant.
	GrantsMode string   // The grant mode, either "enforce" or "append".
}

// GrantRoles grants roles to a user.
//
// Parameters:
//
//	grant: A GrantRole struct containing the details of the roles to be granted.
//
// Returns:
//
//	An error if the grant operation fails.
func (c *Client) GrantRoles(grant GrantRole) error {
	if grant.GrantsMode == "enforce" {
		currentRoles, err := c.GetCurrentRoles(grant.Principal)
		if err != nil {
			return err
		}

		// Revoke roles that are not in the desired list
		for _, currentRole := range currentRoles {
			found := false
			for _, desiredRole := range grant.Roles {
				if strings.EqualFold(currentRole, desiredRole) {
					found = true
					break
				}
			}
			if !found {
				revokeSQL := fmt.Sprintf("REVOKE %s FROM %s", currentRole, grant.Principal)
				if _, err := c.DB.Exec(revokeSQL); err != nil {
					return err
				}
			}
		}
	}

	// Grant the desired roles
	if len(grant.Roles) > 0 {
		roles := strings.Join(grant.Roles, ",")
		grantSQL := fmt.Sprintf("GRANT %s TO %s", roles, grant.Principal)
		_, err := c.DB.Exec(grantSQL)
		return err
	}
	return nil
}

// RevokeRoles revokes roles from a user.
//
// Parameters:
//
//	grant: A GrantRole struct containing the details of the roles to be revoked.
//
// Returns:
//
//	An error if the revoke operation fails.
func (c *Client) RevokeRoles(grant GrantRole) error {
	if len(grant.Roles) > 0 {
		roles := strings.Join(grant.Roles, ",")
		revokeSQL := fmt.Sprintf("REVOKE %s FROM %s", roles, grant.Principal)
		_, err := c.DB.Exec(revokeSQL)
		return err
	}
	return nil
}

// GetCurrentRoles returns the current roles for a user.
//
// Parameters:
//
//	principal: The name of the user to check.
//
// Returns:
//
//	A slice of strings containing the current roles and an error if the check fails.
func (c *Client) GetCurrentRoles(principal string) ([]string, error) {
	var roles []string
	sql := "SELECT granted_role FROM dba_role_privs WHERE grantee = UPPER(:1)"
	rows, err := c.DB.Query(sql, principal)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, strings.ToLower(role))
	}
	return roles, nil
}
