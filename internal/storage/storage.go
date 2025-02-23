package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func NewStorage() (db *sql.DB, err error) {

	appPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	dbFile := filepath.Join(filepath.Dir(appPath), "scheduler.db")
	_, err = os.Stat(dbFile)
	fmt.Println("Checking file")

	var install bool
	if err != nil {
		install = true
	}
	if install {
		db, err := sql.Open("sqlite3", dbFile)
		if err != nil {
			return nil, err
		}
		fmt.Println("Connect to database")
		_, err = db.Exec("CREATE TABLE scheduler (id INTEGER PRIMARY KEY AUTOINCREMENT, date TEXT NOT NULL DEFAULT '', title VARCHAR(128) NOT NULL DEFAULT '', comment VARCHAR(256) NOT NULL DEFAULT '', repeat VARCHAR(128) NOT NULL DEFAULT '')")
		if err != nil {
			return nil, err
		}
		_, err = db.Exec("CREATE INDEX date_sort ON scheduler (date)")
		if err != nil {
			return nil, err
		}
		fmt.Println("Table created")
		//defer db.Close()
	} else {
		_, err := sql.Open("sqlite3", dbFile)
		if err != nil {
			return nil, err
		}
		//fmt.Println("Connect to database")
		//defer db.Close()
	}
	return db, nil
}
