// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oracle

import "fmt"

// Directory represents an Oracle database directory.
type Directory struct {
	Name string // The name of the directory.
	Path string // The path of the directory.
}

// CreateDirectory creates a new directory in the Oracle database.
//
// Parameters:
//
//	directory: A Directory struct containing the details of the directory to be created.
//
// Returns:
//
//	An error if the directory creation fails.
func (c *Client) CreateDirectory(directory Directory) error {
	sql := fmt.Sprintf("CREATE OR REPLACE DIRECTORY %s AS '%s'", directory.Name, directory.Path)
	_, err := c.DB.Exec(sql)
	return err
}

// DropDirectory drops a directory from the Oracle database.
//
// Parameters:
//
//	directoryName: The name of the directory to be dropped.
//
// Returns:
//
//	An error if the directory drop fails.
func (c *Client) DropDirectory(directoryName string) error {
	sql := fmt.Sprintf("DROP DIRECTORY %s", directoryName)
	_, err := c.DB.Exec(sql)
	return err
}

// DirectoryExists checks if a directory exists in the database.
//
// Parameters:
//
//	directoryName: The name of the directory to check.
//
// Returns:
//
//	A boolean indicating whether the directory exists, and an error if the check fails.
func (c *Client) DirectoryExists(directoryName string) (bool, error) {
	var count int
	sql := "SELECT COUNT(*) FROM dba_directories WHERE directory_name = UPPER(:1)"
	err := c.DB.QueryRow(sql, directoryName).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ReadDirectory reads a directory's details from the database.
//
// Parameters:
//
//	directoryName: The name of the directory to read.
//
// Returns:
//
//	A Directory struct containing the directory's details, and an error if the read fails.
func (c *Client) ReadDirectory(directoryName string) (*Directory, error) {
	directory := &Directory{}
	sql := "SELECT directory_name, directory_path FROM dba_directories WHERE directory_name = UPPER(:1)"
	err := c.DB.QueryRow(sql, directoryName).Scan(&directory.Name, &directory.Path)
	if err != nil {
		return nil, err
	}
	return directory, nil
}
