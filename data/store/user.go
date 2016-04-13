package store

import (
	"database/sql"
	"golang.org/x/crypto/bcrypt"
	"github.com/warmans/dbr"
	"time"
)

type User struct {
	ID       int64  `json:"id",db:"id"`
	Username string `json:"username",db:"username"`
	Password string `json:"-",db:"password"`
	Admin    bool   `json:"-",db:"admin"`
}

type UserStore struct {
	DB *dbr.Session
}

func (a *UserStore) Authenticate(username, password string) (*User, error) {

	time.Sleep(time.Second)

	var paswordHash string
	user := &User{}

	err := a.DB.QueryRow(
		"SELECT id, username, password FROM user WHERE username = ?",
		username,
	).Scan(
		&user.ID,
		&user.Username,
		&paswordHash,
	)

	//username not found
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	//password invalid
	if err := bcrypt.CompareHashAndPassword([]byte(paswordHash), []byte(password)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return nil, nil
		}
		return nil, err
	}

	//success
	return user, nil
}

func (a *UserStore) Register(username, password string) (*User, error) {

	time.Sleep(time.Second)

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{Username: username, Password: string(hash)}

	res, err := a.DB.InsertInto("user").Columns("username", "password").Record(user).Exec()
	if err != nil {
		return nil, err //rely on the unique username constraint
	}

	insertID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	//cleanup record
	user.Password = ""
	user.ID = insertID

	return user, nil
}

func (a *UserStore) GetUser(id int64) (*User, error) {
	user := &User{}
	err := a.DB.QueryRow("SELECT id, username FROM user WHERE id = ?", id).Scan(&user.ID, &user.Username)
	if err != nil {
		return nil, err
	}
	return user, nil
}