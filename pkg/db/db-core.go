package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

type User struct {
	ID           int
	Username     string
	Words        []Word
	LanguageFrom string
	LanguageTo   string
}

func NewUser(id int, username string, words []Word, languageFrom string, languageTo string) User {
	return User{ID: id, Username: username, Words: words, LanguageFrom: languageFrom, LanguageTo: languageTo}
}

type Word struct {
	ID         int
	Original   string
	Translated []string
}

func New(path string) (*Storage, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		f, err := os.Create(path)
		if err != nil {
			return nil, err
		}
		f.Close()
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &Storage{db}, nil
}

func (storage *Storage) InitTable(dbSchema string) error {
	_, err := storage.db.Exec(dbSchema)
	return err
}

func (storage *Storage) Save(user User) error {
	_, err := storage.db.Exec(
		`INSERT OR REPLACE INTO users (id, username, lang_from, lang_to)
		 VALUES (?, ?, ?, ?)`,
		user.ID, user.Username, user.LanguageFrom, user.LanguageTo,
	)
	if err != nil {
		return fmt.Errorf("could not insert user: %w", err)
	}

	for _, w := range user.Words {
		data, err := json.Marshal(w.Translated)
		if err != nil {
			return fmt.Errorf("could not marshal translations: %w", err)
		}

		_, err = storage.db.Exec(
			`INSERT OR IGNORE INTO words (user_id, original, translated)
			 VALUES (?, ?, ?)`,
			user.ID, w.Original, string(data),
		)

		if err != nil {
			return fmt.Errorf("could not insert word '%s': %w", w.Original, err)
		}
	}

	return nil
}

func (storage *Storage) Get(userID int) (*User, error) {
	row := storage.db.QueryRow(
		`SELECT id, username, lang_from, lang_to FROM users WHERE id = ?`,
		userID,
	)

	var u User
	if err := row.Scan(&u.ID, &u.Username, &u.LanguageFrom, &u.LanguageTo); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("could not scan user: %w", err)
	}

	rows, err := storage.db.Query(`SELECT id, original, translated FROM words WHERE user_id = ?`, userID)
	if err != nil {
		return nil, fmt.Errorf("could not get words: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var w Word
		var translatedStr string
		if err := rows.Scan(&w.ID, &w.Original, &translatedStr); err != nil {
			return nil, fmt.Errorf("could not scan word: %w", err)
		}

		json.Unmarshal([]byte(translatedStr), &w.Translated)
		u.Words = append(u.Words, w)
	}

	return &u, nil
}
