// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oracle

import (
	"log"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGrantRoles(t *testing.T) {
	dbUser := os.Getenv("ORACLE_USERNAME")
	dbPassword := os.Getenv("ORACLE_PASSWORD")
	dbHost := os.Getenv("ORACLE_HOST")
	dbPortStr := os.Getenv("ORACLE_PORT")
	dbServiceName := os.Getenv("ORACLE_SERVICE")

	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		log.Fatalf("Error converting port to integer: %v", err)
	}

	client, err := NewClient(dbHost, dbServiceName, dbUser, dbPassword, dbPort)
	if err != nil {
		log.Fatalf("Error creating Oracle client: %v", err)
	}
	defer client.DB.Close()

	_, err = client.DB.Exec("CREATE USER test_user IDENTIFIED BY MyPassword123")
	assert.NoError(t, err)
	defer client.DB.Exec("DROP USER test_user")

	_, err = client.DB.Exec("CREATE ROLE test_role")
	assert.NoError(t, err)
	defer client.DB.Exec("DROP ROLE test_role")

	grant := GrantRole{
		Principal:  "test_user",
		Roles:      []string{"test_role"},
		GrantsMode: "enforce",
	}
	err = client.GrantRoles(grant)
	assert.NoError(t, err)

	currentRoles, err := client.GetCurrentRoles("test_user")
	assert.NoError(t, err)
	assert.Equal(t, []string{"test_role"}, currentRoles)

	grant.Roles = []string{"test_role"}
	err = client.RevokeRoles(grant)
	assert.NoError(t, err)

	currentRoles, err = client.GetCurrentRoles("test_user")
	assert.NoError(t, err)
	assert.Empty(t, currentRoles)
}
