// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oracle_test

import (
	"log"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"terraform-provider-oracle/internal/oracle"
)

func TestGrant(t *testing.T) {
	dbUser := os.Getenv("ORACLE_USERNAME")
	dbPassword := os.Getenv("ORACLE_PASSWORD")
	dbHost := os.Getenv("ORACLE_HOST")
	dbPortStr := os.Getenv("ORACLE_PORT")
	dbServiceName := os.Getenv("ORACLE_SERVICE")

	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		log.Fatalf("Error converting port to integer: %v", err)
	}

	client, err := oracle.NewClient(dbHost, dbServiceName, dbUser, dbPassword, dbPort)
	if err != nil {
		log.Fatalf("Error creating Oracle client: %v", err)
	}
	defer client.DB.Close()

	testUser := oracle.User{
		Username:           "testgrantuser",
		Password:           "testpassword",
		AuthenticationType: "password",
	}

	exists, err := client.UserExists(testUser.Username)
	assert.NoError(t, err)
	if exists {
		assert.NoError(t, client.DropUser(testUser.Username))
	}

	assert.NoError(t, client.CreateUser(testUser))

	// Create a dummy table for testing object privileges
	_, err = client.ExecuteSQL("CREATE TABLE system.test_table (id NUMBER)")
	assert.NoError(t, err)

	// Create a dummy directory for testing directory privileges
	_, err = client.ExecuteSQL("CREATE OR REPLACE DIRECTORY test_dir AS '/tmp'")
	assert.NoError(t, err)

	// Grant system privileges with enforce mode
	systemGrant := oracle.Grant{
		Principal:  testUser.Username,
		Privileges: []string{"CREATE SESSION"},
		GrantsMode: "enforce",
	}
	assert.NoError(t, client.GrantSystemPrivileges(systemGrant))

	// Grant another privilege to test enforce mode
	_, err = client.ExecuteSQL("GRANT CREATE TABLE TO " + testUser.Username)
	assert.NoError(t, err)

	// Run GrantSystemPrivileges again with enforce mode to revoke the extra privilege
	assert.NoError(t, client.GrantSystemPrivileges(systemGrant))

	// Grant object privileges with enforce mode
	objectGrant := oracle.ObjectPrivilege{
		Principal:  testUser.Username,
		Object:     "system.test_table",
		Privileges: []string{"SELECT"},
		GrantsMode: "enforce",
	}
	assert.NoError(t, client.GrantObjectPrivileges(objectGrant))

	// Grant another privilege to test enforce mode
	_, err = client.ExecuteSQL("GRANT UPDATE ON system.test_table TO " + testUser.Username)
	assert.NoError(t, err)

	// Run GrantObjectPrivileges again with enforce mode to revoke the extra privilege
	assert.NoError(t, client.GrantObjectPrivileges(objectGrant))

	// Grant directory privileges with enforce mode
	directoryGrant := oracle.DirectoryPrivilege{
		Principal:  testUser.Username,
		Directory:  "test_dir",
		Privileges: []string{"READ"},
		GrantsMode: "enforce",
	}
	assert.NoError(t, client.GrantDirectoryPrivileges(directoryGrant))

	// Grant another privilege to test enforce mode
	_, err = client.ExecuteSQL("GRANT WRITE ON DIRECTORY test_dir TO " + testUser.Username)
	assert.NoError(t, err)

	// Run GrantDirectoryPrivileges again with enforce mode to revoke the extra privilege
	assert.NoError(t, client.GrantDirectoryPrivileges(directoryGrant))

	// Drop the dummy table
	_, err = client.ExecuteSQL("DROP TABLE system.test_table")
	assert.NoError(t, err)

	// Drop the dummy directory
	_, err = client.ExecuteSQL("DROP DIRECTORY test_dir")
	assert.NoError(t, err)

	// Drop the user
	assert.NoError(t, client.DropUser(testUser.Username))
}
