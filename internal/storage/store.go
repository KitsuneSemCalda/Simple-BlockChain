package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"KitsuneSemCalda/SBC/internal/blockchain"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

func NewStore(dataDir string) (*Store, error) {
	if dataDir == "" {
		dataDir = "."
	}

	err := os.MkdirAll(dataDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "blockchain.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	query := `
	CREATE TABLE IF NOT EXISTS blocks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		block_index INTEGER NOT NULL,
		timestamp TEXT NOT NULL,
		bpm INTEGER NOT NULL,
		hash TEXT NOT NULL UNIQUE,
		prev_hash TEXT NOT NULL
	);`

	_, err = db.Exec(query)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Save(bc *blockchain.Blockchain) error {
	blocks := bc.GetAllBlocks()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT OR IGNORE INTO blocks(block_index, timestamp, bpm, hash, prev_hash) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, b := range blocks {
		_, err = stmt.Exec(b.Index, b.Timestamp.Format(time.RFC3339Nano), b.BPM, b.Hash, b.PrevHash)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) Load(bc *blockchain.Blockchain) error {
	query := "SELECT block_index, timestamp, bpm, hash, prev_hash FROM blocks ORDER BY block_index ASC"
	rows, err := s.db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var b blockchain.Block
		var ts string
		err = rows.Scan(&b.Index, &ts, &b.BPM, &b.Hash, &b.PrevHash)
		if err != nil {
			return err
		}
		b.Timestamp, _ = time.Parse(time.RFC3339Nano, ts)
		bc.ProcessBlock(&b)
	}

	return nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
