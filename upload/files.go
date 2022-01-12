package upload

import (
	"database/sql"
	"fmt"
	"time"
)

// File holds information on our documents that we will be storing
type File struct {
	Id           string `json:"Id"`
	Name         string `json:"name"`
	Path         string `json:"path"`
	UploadedBy   string `json:"uploaded_by"`
	DateUploaded string `json:"date_uploaded"`
}

// CreateDB connects to our database, creates the files table that can store File data
func CreateTable(db *sql.DB) error {
	fmt.Println("creating table")
	_, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS files (
			    id SERIAL NOT NULL PRIMARY KEY,
			    name VARCHAR(250) NOT NULL,
				path VARCHAR NOT NULL,
				uploaded_by VARCHAR(250) NOT NULL,
			    date_uploaded DATE NOT NULL
			);
		`)
	fmt.Println("created db")
	return err
}

// GetFile Queries the db for a file record using its ID
func GetFile(db *sql.DB, id int) *sql.Row {
	row := db.QueryRow("SELECT * FROM files WHERE id = $1", id)
	return row
}

// CreateFile Inserts a file instance into the database
func CreateFile(db *sql.DB, name string, path string, uploaded_by string) error {
	now := time.Now().Local()
	currentTime := now.Format("2006-02-01")
	fmt.Println(currentTime, "currentTime")
	sqlStatement := `INSERT INTO files (name, path, uploaded_by, date_uploaded) VALUES ($1, $2, $3, $4)`
	_, err := db.Exec(sqlStatement, name, path, uploaded_by, currentTime)
	if err != nil {
		panic(err)
	}
	return err
}
