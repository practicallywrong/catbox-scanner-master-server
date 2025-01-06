package database

import (
	"catbox-scanner-master/internal/config"
	"database/sql"
	"log"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	conn  *sql.DB
	queue chan Entry
	wg    sync.WaitGroup
	quit  chan struct{}
}

type Entry struct {
	ID  string
	Ext string
}

func InitDatabase() (*Database, error) {
	db, err := sql.Open("sqlite3", config.AppConfig.DBPath)
	if err != nil {
		return nil, err
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS found_ids (
		id CHAR(6) NOT NULL UNIQUE,
		ext VARCHAR(10) NOT NULL
	);`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		return nil, err
	}

	database := &Database{
		conn:  db,
		queue: make(chan Entry, 10000),
		quit:  make(chan struct{}),
	}

	database.startWorker()

	return database, nil
}

func (db *Database) startWorker() {
	db.wg.Add(1)

	go func() {
		defer db.wg.Done()

		for {
			select {
			case entry := <-db.queue:
				if err := db.insertEntryToDB(entry.ID, entry.Ext); err != nil {
					log.Printf("Error inserting entry into database: %v", err)
				}
			case <-db.quit:
				log.Println("Worker shutting down after processing remaining entries.")
				return
			}
		}
	}()
}

func (db *Database) insertEntryToDB(id, ext string) error {
	statement := "INSERT OR IGNORE INTO found_ids (id, ext) VALUES (?, ?);"
	_, err := db.conn.Exec(statement, id, ext)
	return err
}

func (db *Database) InsertEntry(id, ext string) {
	db.queue <- Entry{ID: id, Ext: ext}
}

func (db *Database) GetTotalRows() (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM found_ids;"
	err := db.conn.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (db *Database) Stop() {
	close(db.quit)
	db.wg.Wait()
}
