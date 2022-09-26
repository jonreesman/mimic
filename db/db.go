package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

type Source int

const (
	DISCORD = iota
	ELEMENT
)

type DBManager struct {
	db     *sql.DB
	dbName string
	dbUser string
	dbPwd  string
	dbPort string
	dbURL  string
	URI    string
}

func NewManager() (DBManager, error) {
	var (
		d   DBManager
		err error
	)
	d.dbUser = os.Getenv("DB_USER")
	d.dbPwd = os.Getenv("DB_PWD")
	d.dbName = os.Getenv("DB_NAME")
	d.dbURL = os.Getenv("DB_URL")
	d.dbPort = os.Getenv("DB_PORT")

	d.URI = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", d.dbUser, d.dbPwd, d.dbURL, d.dbPort, d.dbName)
	d.db, err = sql.Open("mysql", d.URI)
	if err != nil {
		log.Printf("Failed to open connection in DB NewManager(): %v", err)
		return DBManager{}, err
	}

	if err := d.db.Ping(); err != nil {
		log.Printf("Failed to Ping DB in NewManager(): %v", err)
		return DBManager{}, err
	}

	if _, err := d.db.Exec(fmt.Sprintf("USE %s", d.dbName)); err != nil {
		log.Printf("Failed to `USE %s` in NewManager(): %v", d.dbName, err)
		return DBManager{}, err
	}

	return d, nil
}

const addMessageElementQuery = `
INSERT INTO element_messages(message_id, user_id, channel_id, msg, time_stamp) ` +
	`VALUES (?, ?, ?, ?, ?)`

const addMessageDiscordQuery = `
INSERT INTO discord_messages(message_id, user_id, channel_id, msg, time_stamp) ` +
	`VALUES (?, ?, ?, ?, ?)`

func (d *DBManager) AddMessage(messageID, userID, channelID, msg string, timestamp int64, source Source) error {
	q := addMessageElementQuery
	if source == DISCORD {
		q = addMessageDiscordQuery
	}
	if _, err := d.db.Exec(q,
		messageID,
		userID,
		channelID,
		msg,
		timestamp,
	); err != nil {
		log.Print("Error in addMessage(): ", err)
		return err
	}
	return nil
}

const getAllUsersQuery = `
SELECT user_id from users where source=?`

func (d *DBManager) GetAllUsers(source Source) []string {
	rows, err := d.db.Query(getAllUsersQuery, source)
	if err != nil {
		log.Printf("Error in GetAllUsers(): %v", err)
	}
	var (
		userID string
		users  []string
	)
	for rows.Next() {
		if err := rows.Scan(&userID); err != nil {
			log.Print(err)
		}
		users = append(users, userID)
	}
	return users
}

const addUserQuery = `
INSERT INTO users (user_id, source) VALUES (?, ?)`

func (d *DBManager) AddUser(userID string, source Source) error {
	if _, err := d.db.Exec(addUserQuery, userID, source); err != nil {
		log.Printf("Error in AddUser(): %v", err)
		return err
	}
	return nil
}

const getOldestMessageDiscordQuery = `
SELECT message_id from discord_messages where channel_id=? ORDER BY message_id ASC LIMIT 1`
const getOldestMessageElementQuery = `
SELECT message_id from element_messages where channel_id=? ORDER BY message_id ASC LIMIT 1`

func (d *DBManager) GetOldestMessage(channelID string, source Source) (string, error) {
	s := getOldestMessageDiscordQuery
	if source == ELEMENT {
		s = getOldestMessageElementQuery
	}
	rows, err := d.db.Query(s, channelID)
	if err != nil {
		log.Printf("Error in GetOldestMessage(): %v", err)
		return "", err
	}
	var messageID string
	for rows.Next() {
		if err := rows.Scan(&messageID); err != nil {
			log.Print(err)
			return "", err
		}
	}
	return messageID, nil
}

const getNewestMessageDiscordQuery = `
SELECT message_id from discord_messages where channel_id=? ORDER BY message_id DESC LIMIT 1`
const getNewestMessageElementQuery = `
SELECT message_id from element_messages where channel_id=? ORDER BY message_id DESC LIMIT 1`

func (d *DBManager) GetNewestMessage(channelID string, source Source) (string, error) {
	s := getNewestMessageDiscordQuery
	if source == ELEMENT {
		s = getNewestMessageElementQuery
	}
	rows, err := d.db.Query(s, channelID)
	if err != nil {
		log.Printf("Error in GetNewesrMessage(): %v", err)
		return "", err
	}
	var messageID string
	for rows.Next() {
		if err := rows.Scan(&messageID); err != nil {
			log.Print(err)
			return "", err
		}
	}
	return messageID, nil
}
