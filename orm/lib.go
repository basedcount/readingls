package orm

import (
	"database/sql"
	"net/url"
)

func New(db *sql.DB) *ORM {
	return &ORM{db}
}

type ORM struct {
	db *sql.DB
}

type ReadingEntry struct {
	Id        int       `json:"id" sql:"id"`
	Url       string    `json:"url" sql:"url"`
	Submitted []byte `json:"submitted" sql:"submitted"`
	Username  string    `json:"username" sql:"username"`
}

func (o ORM) InsertURL(link *url.URL, username string) error {
	_, err := o.db.Exec("INSERT INTO readingls (url, username) values (?,?)", link.String(), username)
	if err != nil {
		return err
	}
	return nil
}

func (o ORM) GetLinks(username string) ([]ReadingEntry, error) {
	res, err := o.db.Query("SELECT * from readingls where username = ?", username)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	result := make([]ReadingEntry, 0)
	for res.Next() {
		var entry ReadingEntry
		err := res.Scan(&entry.Id, &entry.Url, &entry.Submitted, &entry.Username)
		if err != nil {
			return nil, err
		}
		result = append(result, entry)
	}

	return result, nil
}
