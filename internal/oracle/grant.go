// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oracle

import (
	"fmt"
	"strings"
)

// Grant represents a system privilege to be granted to a user or role.
type Grant struct {
	Principal  string   // The user or role to whom the privileges should be granted.
	Privileges []string // A list of system privileges to grant.
	GrantsMode string   // The grants mode, either "enforce" or "append".
}

// ObjectPrivilege represents a privilege on a specific database object.
type ObjectPrivilege struct {
	Principal  string   // The user or role to whom the privileges should be granted.
	Object     string   // The database object on which to grant the privileges.
	Owner      string   // The owner of the database object.
	Privileges []string // A list of object privileges to grant.
	GrantsMode string   // The grants mode, either "enforce" or "append".
}

// DirectoryPrivilege represents a privilege on a specific database directory.
type DirectoryPrivilege struct {
	Principal  string   // The user or role to whom the privileges should be granted.
	Directory  string   // The database directory on which to grant the privileges.
	Privileges []string // A list of directory privileges to grant.
	GrantsMode string   // The grants mode, either "enforce" or "append".
}

// GrantSystemPrivileges grants system privileges to a user or role.
//
// Parameters:
//
//	grant: A Grant struct containing the details of the privileges to be granted.
//
// Returns:
//
//	An error if the grant operation fails.
func (c *Client) GrantSystemPrivileges(grant Grant) error {
	if grant.GrantsMode == "enforce" {
		currentPrivs, err := c.GetCurrentSystemPrivileges(grant.Principal)
		if err != nil {
			return err
		}

		// Revoke privileges that are not in the desired list
		for _, currentPriv := range currentPrivs {
			found := false
			for _, desiredPriv := range grant.Privileges {
				if strings.EqualFold(currentPriv, desiredPriv) {
					found = true
					break
				}
			}
			if !found {
				revokeSQL := fmt.Sprintf("REVOKE %s FROM %s", currentPriv, grant.Principal)
				if _, err := c.DB.Exec(revokeSQL); err != nil {
					return err
				}
			}
		}
	}

	// Grant the desired privileges
	if len(grant.Privileges) > 0 {
		privs := strings.Join(grant.Privileges, ",")
		grantSQL := fmt.Sprintf("GRANT %s TO %s", privs, grant.Principal)
		_, err := c.DB.Exec(grantSQL)
		return err
	}
	return nil
}

// GrantObjectPrivileges grants object privileges to a user or role.
//
// Parameters:
//
//	privilege: An ObjectPrivilege struct containing the details of the privileges to be granted.
//
// Returns:
//
//	An error if the grant operation fails.
func (c *Client) GrantObjectPrivileges(privilege ObjectPrivilege) error {
	object := fmt.Sprintf("%s.%s", privilege.Owner, privilege.Object)
	if privilege.Owner == "" {
		object = privilege.Object
	}
	if privilege.GrantsMode == "enforce" {
		currentPrivs, err := c.GetCurrentObjectPrivileges(privilege.Principal, privilege.Owner, privilege.Object)
		if err != nil {
			return err
		}

		// Revoke privileges that are not in the desired list
		for _, currentPriv := range currentPrivs {
			found := false
			for _, desiredPriv := range privilege.Privileges {
				if strings.EqualFold(currentPriv, desiredPriv) {
					found = true
					break
				}
			}
			if !found {
				revokeSQL := fmt.Sprintf("REVOKE %s ON %s FROM %s", currentPriv, object, privilege.Principal)
				if _, err := c.DB.Exec(revokeSQL); err != nil {
					return err
				}
			}
		}
	}

	// Grant the desired privileges
	if len(privilege.Privileges) > 0 {
		for _, priv := range privilege.Privileges {
			grantSQL := fmt.Sprintf("GRANT %s ON %s TO %s", priv, object, privilege.Principal)
			if strings.Contains(strings.ToUpper(priv), "WITH GRANT OPTION") {
				grantSQL = fmt.Sprintf("GRANT %s ON %s TO %s WITH GRANT OPTION", strings.Replace(priv, " WITH GRANT OPTION", "", 1), object, privilege.Principal)
			}
			_, err := c.DB.Exec(grantSQL)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// GrantDirectoryPrivileges grants directory privileges to a user or role.
//
// Parameters:
//
//	privilege: a DirectoryPrivilege struct containing the details of the privileges to be granted.
//
// Returns:
//
//	An error if the grant operation fails.
func (c *Client) GrantDirectoryPrivileges(privilege DirectoryPrivilege) error {
	if privilege.GrantsMode == "enforce" {
		currentPrivs, err := c.GetCurrentDirectoryPrivileges(privilege.Principal, privilege.Directory)
		if err != nil {
			return err
		}

		// Revoke privileges that are not in the desired list
		for _, currentPriv := range currentPrivs {
			found := false
			for _, desiredPriv := range privilege.Privileges {
				if strings.EqualFold(currentPriv, desiredPriv) {
					found = true
					break
				}
			}
			if !found {
				revokeSQL := fmt.Sprintf("REVOKE %s ON DIRECTORY %s FROM %s", currentPriv, privilege.Directory, privilege.Principal)
				if _, err := c.DB.Exec(revokeSQL); err != nil {
					return err
				}
			}
		}
	}

	// Grant the desired privileges
	if len(privilege.Privileges) > 0 {
		for _, priv := range privilege.Privileges {
			grantSQL := fmt.Sprintf("GRANT %s ON DIRECTORY %s TO %s", priv, privilege.Directory, privilege.Principal)
			if strings.Contains(strings.ToUpper(priv), "WITH GRANT OPTION") {
				grantSQL = fmt.Sprintf("GRANT %s ON DIRECTORY %s TO %s WITH GRANT OPTION", strings.Replace(priv, " WITH GRANT OPTION", "", 1), privilege.Directory, privilege.Principal)
			}
			_, err := c.DB.Exec(grantSQL)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// GetCurrentSystemPrivileges returns the current system privileges for a user.
//
// Parameters:
//
//	principal: The name of the user or role to check.
//
// Returns:
//
//	A slice of strings containing the current system privileges, and an error if the check fails.
func (c *Client) GetCurrentSystemPrivileges(principal string) ([]string, error) {
	var privileges []string
	sql := "SELECT privilege, admin_option FROM dba_sys_privs WHERE grantee = UPPER(:1)"
	rows, err := c.DB.Query(sql, principal)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var privilege string
		var adminOption string
		if err := rows.Scan(&privilege, &adminOption); err != nil {
			return nil, err
		}
		privileges = append(privileges, grantOption(privilege, adminOption))
	}
	return privileges, nil
}

// GetCurrentObjectPrivileges returns the current object privileges for a user.
//
// Parameters:
//
//	principal: The name of the user or role to check.
//	object: The name of the object to check.
//
// Returns:
//
//	A slice of strings containing the current object privileges, and an error if the check fails.
func (c *Client) GetCurrentObjectPrivileges(principal, owner, object string) ([]string, error) {
	var privileges []string
	sql := "SELECT privilege, grantable FROM dba_tab_privs WHERE grantee = UPPER(:1) AND owner = UPPER(:2) AND table_name = UPPER(:3)"
	if owner == "" {
		sql = "SELECT privilege, grantable FROM dba_tab_privs WHERE grantee = UPPER(:1) AND table_name = UPPER(:2)"
	}
	rows, err := c.DB.Query(sql, principal, owner, object)
	if owner == "" {
		rows, err = c.DB.Query(sql, principal, object)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var privilege string
		var grantable string
		if err := rows.Scan(&privilege, &grantable); err != nil {
			return nil, err
		}
		privileges = append(privileges, grantOption(privilege, grantable))
	}
	return privileges, nil
}

// GetCurrentDirectoryPrivileges returns the current directory privileges for a user.
//
// Parameters:
//
//	principal: The name of the user or role to check.
//	directory: The name of the directory to check.
//
// Returns:
//
//	A slice of strings containing the current directory privileges and an error if the check fails.
func (c *Client) GetCurrentDirectoryPrivileges(principal, directory string) ([]string, error) {
	var privileges []string
	sql := "SELECT privilege, grantable FROM all_tab_privs WHERE grantee = UPPER(:1) AND table_name = UPPER(:2) AND type = 'DIRECTORY'"
	rows, err := c.DB.Query(sql, principal, directory)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var privilege string
		var grantable string
		if err := rows.Scan(&privilege, &grantable); err != nil {
			return nil, err
		}
		privileges = append(privileges, grantOption(privilege, grantable))
	}
	return privileges, nil
}

func grantOption(privilege, option string) string {
	if option == "YES" {
		return fmt.Sprintf("%s WITH ADMIN OPTION", privilege)
	}
	return privilege
}
