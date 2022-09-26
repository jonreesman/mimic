package scraped

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/jonreesman/mimic/db"
)

type Attachment struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	URL         string `json:"url"`
	ProxyURL    string `json:"proxy_url"`
	Width       int64  `json:"width"`
	Height      int64  `json:"height"`
	ContentType string `json:"content_type"`
}

type User struct {
	ID               string `json:"id"`
	Username         string `json:"username"`
	Avatar           string `json:"avatar"`
	AvatarDecoration string `json:"avatar_decoration"`
	Discriminator    string `json:"discriminator"`
	Bot              bool   `json:"bot"`
}

type APIMsg struct {
	ID          string       `json:"id"`
	Type        int          `json:"type"`
	Content     string       `json:"content"`
	ChannelID   string       `json:"channel_id"`
	Attachments []Attachment `json:"attachments"`
	Author      User         `json:"author"`
	Mentions    []User       `json:"mentions"`
	Timestamp   time.Time    `json:"timestamp"`
}

type Scraper struct {
	token  string
	client *http.Client
}

func NewScraper(token string) Scraper {
	return Scraper{
		token:  token,
		client: &http.Client{},
	}
}

func (s Scraper) FullScrape(channel string) error {
	req, err := newRequest(s, "channels/"+channel+"/messages")
	if err != nil {
		log.Printf("Error in FullScrape() generating newRequest: %v", err)
	}
	req.Header = http.Header{
		"Authorization": []string{"Bot " + s.token},
	}
	res, err := s.client.Do(req)
	if err != nil {
		log.Printf("Scrape(): Error getting request from DiscordAPI: %v", err)
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode > 299 {
		log.Printf("Response failed with status code: %d and \nbody: %s", res.StatusCode, body)
		return nil
	}
	var msg []APIMsg
	if err := json.Unmarshal(body, &msg); err != nil {
		log.Printf("Failed to Unmarshal response: %v", err)
		return nil
	}
	d, err := db.NewManager()
	if err != nil {
		log.Printf("Error in Scrape() creating DBManager: %v", err)
	}
	userMap := getUsers(d, db.DISCORD)
	var lastMsg string
	for i, m := range msg {
		if userMap[m.Author.ID] == 0 {
			if err := d.AddUser(m.Author.ID, db.DISCORD); err != nil {
				log.Printf("Failed to add user %s, skipping message %s\n", m.Author.ID, m.Content)
				continue
			}
			userMap[m.Author.ID] = 1
		}
		if err := d.AddMessage(m.ID, m.Author.ID, m.ChannelID, m.Content, m.Timestamp.Unix(), db.DISCORD); err != nil {
			log.Printf("Failed to add message %s with error: %v", m.ID, err)
		}
		if i == len(msg)-1 {
			lastMsg = m.ID
		}
	}
	return s.ScrapeBefore(channel, lastMsg)
}

func (s Scraper) Scrape(channel string) error {
	d, err := db.NewManager()
	if err != nil {
		log.Printf("Error in Scrape() creating DBManager: %v", err)
	}
	mostRecentMsg, err := d.GetNewestMessage(channel, db.DISCORD)
	if err != nil {
		log.Printf("Error in Scrape(): Error pulling most recent Msg from DB: %v", err)
		return err
	}
	if err := s.ScrapeAfter(channel, mostRecentMsg); err != nil {
		log.Printf("Error in Scrape(): ScrapeAfter() failed: %v", err)
		return err
	}
	return nil
}

func (s Scraper) ScrapeBefore(channel, beforeID string) error {
	req, err := newRequest(s, "channels/"+channel+"/messages")
	if err != nil {
		log.Printf("Error in ScrapeBefore() generating newRequest: %v", err)
	}
	q := req.URL.Query()
	q.Add("before", beforeID)
	req.URL.RawQuery = q.Encode()
	log.Printf("Query: %s", req.URL.String())
	res, err := s.client.Do(req)
	if err != nil {
		log.Printf("Scrape(): Error getting request from DiscordAPI: %v", err)
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode > 299 {
		log.Printf("Response failed with status code: %d and \nbody: %s", res.StatusCode, body)
		return nil
	}
	var msg []APIMsg
	if err := json.Unmarshal(body, &msg); err != nil {
		log.Printf("Failed to Unmarshal response: %v", err)
		return nil
	}
	d, err := db.NewManager()
	if err != nil {
		log.Printf("Error in Scrape() creating DBManager: %v", err)
	}
	if len(msg) == 0 {
		return nil
	}
	userMap := getUsers(d, db.DISCORD)
	for _, m := range msg {
		if userMap[m.Author.ID] == 0 {
			if err := d.AddUser(m.Author.ID, db.DISCORD); err != nil {
				log.Printf("Failed to add user %s, skipping message %s\n", m.Author.ID, m.Content)
				continue
			}
			userMap[m.Author.ID] = 1
		}
		if err := d.AddMessage(m.ID, m.Author.ID, m.ChannelID, m.Content, m.Timestamp.Unix(), db.DISCORD); err != nil {
			log.Printf("Failed to add message %s with error: %v", m.ID, err)
		}
	}
	return nil
}

func (s Scraper) ScrapeAfter(channel, afterID string) error {
	req, err := newRequest(s, "channels/"+channel+"/messages")
	if err != nil {
		log.Printf("Error in ScrapeAfter() generating newRequest: %v", err)
	}
	q := req.URL.Query()
	q.Add("after", afterID)
	req.URL.RawQuery = q.Encode()
	res, err := s.client.Do(req)
	if err != nil {
		log.Printf("Scrape(): Error getting request from DiscordAPI: %v", err)
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode > 299 {
		log.Printf("Response failed with status code: %d and \nbody: %s", res.StatusCode, body)
		return nil
	}
	var msg []APIMsg
	if err := json.Unmarshal(body, &msg); err != nil {
		log.Printf("Failed to Unmarshal response: %v", err)
		return nil
	}
	d, err := db.NewManager()
	if err != nil {
		log.Printf("Error in Scrape() creating DBManager: %v", err)
	}
	if len(msg) == 0 {
		return nil
	}
	userMap := getUsers(d, db.DISCORD)
	for _, m := range msg {
		if userMap[m.Author.ID] == 0 {
			if err := d.AddUser(m.Author.ID, db.DISCORD); err != nil {
				log.Printf("Failed to add user %s, skipping message %s\n", m.Author.ID, m.Content)
				continue
			}
			userMap[m.Author.ID] = 1
		}
		if err := d.AddMessage(m.ID, m.Author.ID, m.ChannelID, m.Content, m.Timestamp.Unix(), db.DISCORD); err != nil {
			log.Printf("Failed to add message %s with error: %v", m.ID, err)
		}
	}
	return nil
}

func getUsers(d db.DBManager, source db.Source) map[string]int {
	users := d.GetAllUsers(source)
	userMap := make(map[string]int, len(users))
	for _, user := range users {
		userMap[user] = 1
	}
	return userMap
}

func (s Scraper) ScrapeChannel(d db.DBManager, channelID string, source db.Source) error {
	var temp string
	for {
		oldestMsg, err := d.GetOldestMessage(channelID, source)
		log.Printf("Oldest msg: %s", oldestMsg)
		if err != nil {
			log.Printf("Error in ScrapeChannel(): %v", err)
		}
		if temp == oldestMsg {
			break
		}
		s.ScrapeBefore(channelID, oldestMsg)
		temp = oldestMsg
		time.Sleep(time.Second)
	}
	log.Printf("Finished scraping channel: %s", channelID)
	return nil
}

func (s Scraper) GetUserAvatar(userID string) string {
	user := s.GetUser(userID)
	return user.Avatar
}

func (s Scraper) GetUserName(userID string) string {
	user := s.GetUser(userID)
	return user.Username
}

func (s Scraper) GetUser(userID string) User {
	req, err := newRequest(s, "users/"+userID)
	if err != nil {
		log.Printf("Error in GetUser(): %v", err)
	}
	res, err := s.client.Do(req)
	if err != nil {
		log.Printf("Error in GetUser() sending request: %v", err)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Error in GetUser() reading response body: %v", err)
	}
	res.Body.Close()
	if res.StatusCode > 299 {
		log.Printf("GetUser(): Response failed with status code: %d and \nbody: %s", res.StatusCode, body)
		return User{}
	}
	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		log.Printf("GetUser(): Failed to Unmarshal response: %v", err)
		return User{}
	}
	return user
}

func newRequest(s Scraper, URL string) (*http.Request, error) {
	req, err := http.NewRequest("GET", "https://discord.com/api/"+URL, nil)
	if err != nil {
		log.Printf("newRequest(): failed to build NewRequest: %v", err)
		return nil, err
	}
	req.Header = http.Header{
		"Authorization": []string{"Bot " + s.token},
	}
	return req, nil
}
