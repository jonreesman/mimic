package db

import (
	"math/rand"
	"sort"
	"strconv"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	d, err := NewManager()
	if err != nil {
		t.Errorf("NewManager() failed for error %v", err)
	}
	if err := d.db.Ping(); err != nil {
		t.Errorf("NewManager() failed for error %v", err)
	}
}

type testMessage struct {
	messageID string
	userID    string
	channelID string
	msg       string
	timestamp int64
	source    Source
}

func randomMessage() testMessage {
	return testMessage{
		messageID: strconv.FormatUint(rand.Uint64(), 10),
		userID:    strconv.FormatUint(rand.Uint64(), 10),
		channelID: strconv.FormatUint(rand.Uint64(), 10),
		msg:       strconv.FormatUint(rand.Uint64(), 10),
		timestamp: rand.Int63(),
		source:    DISCORD,
	}
}

func TestAddMessage(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	d, err := NewManager()
	if err != nil {
		t.Errorf("NewManager() failed for error %v", err)
	}
	var messages []testMessage
	userMap := make(map[string]int, 100)
	channelID := strconv.FormatUint(rand.Uint64(), 10)
	for i := 0; i < 100; i++ {
		msg := randomMessage()
		msg.channelID = channelID
		messages = append(messages, msg)
	}
	for _, m := range messages {
		if userMap[m.userID] == 0 {
			d.AddUser(m.userID, DISCORD)
			userMap[m.userID] = 1
		}
		if err := d.AddMessage(m.messageID, m.userID, m.channelID, m.msg, m.timestamp, m.source); err != nil {
			t.Errorf("AddMessage Failed: %v", err)
		}
	}
	userCheck := d.GetAllUsers(DISCORD)
	if len(userCheck) != len(messages) {
		t.Errorf("Incorrect number of users added. DB does not reflect input params.")
	}
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].messageID < messages[j].messageID
	})
	dbNewest, err := d.GetNewestMessage(channelID, DISCORD)
	if err != nil {
		t.Errorf("GetNewestMessage() failed: %v", err)
	}
	if dbNewest != messages[len(messages)-1].messageID {
		t.Errorf("Database newest doesn't match newest...")
		t.Errorf("DBNewest: %s		Actual Newest: %s", dbNewest, messages[0].messageID)
	}
	dbOldest, err := d.GetOldestMessage(channelID, DISCORD)
	if err != nil {
		t.Errorf("GetOldestMessage() failed: %v", err)
	}
	if dbOldest != messages[0].messageID {
		t.Errorf("Database oldest doesn't match oldest...")
		t.Errorf("DBOldest: %s		Actual Oldest: %s", dbOldest, messages[len(messages)-1].messageID)
	}
}
