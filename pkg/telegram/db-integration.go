package telegram

import "fmt"

const UsersTable string = `
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    username TEXT,
    lang_from TEXT,
    lang_to TEXT
);`

const WordsTable = `
CREATE TABLE IF NOT EXISTS words (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER,
	original TEXT,
	translated TEXT,
	FOREIGN KEY(user_id) REFERENCES users(id),
	UNIQUE(user_id, original)
);`

func (b *Bot) getUserInfo(userID int) (*string, error) {
	user, err := b.db.Get(userID)
	if err != nil {
		return nil, err
	}
	userInfo := "User ID: " + fmt.Sprint(user.ID) +
		"\nUsername: " + user.Username +
		"\nTranslate from: " + user.LanguageFrom +
		"\nTranslate to: " + user.LanguageTo +
		"\nNumber of saved words: " + fmt.Sprint(len(user.Words))
	return &userInfo, nil
}
