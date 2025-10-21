// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oracle_test

import (
	"log"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/neozocloud/terraform-provider-oracle/internal/oracle"
)

func TestRole(t *testing.T) {
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

	testRole := oracle.Role{
		Name: "testrole",
	}

	exists, err := client.RoleExists(testRole.Name)
	assert.NoError(t, err)
	if exists {
		assert.NoError(t, client.DropRole(testRole.Name))
	}

	assert.NoError(t, client.CreateRole(testRole))

	assert.NoError(t, client.DropRole(testRole.Name))
}
