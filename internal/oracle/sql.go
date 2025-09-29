// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oracle

import (
	"database/sql"
)

// ExecuteSQL executes an arbitrary SQL statement.
func (c *Client) ExecuteSQL(sqlStatement string) (*sql.Rows, error) {
	return c.DB.Query(sqlStatement)
}
