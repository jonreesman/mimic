package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jonreesman/mimic/db"
	"github.com/jonreesman/mimic/pb"
	"github.com/jonreesman/mimic/scraped"
	"github.com/jonreesman/mimic/tts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	token           string
	user            string
	username        string
	channel         string
	GRPC_SERVER     string
	VOICE_GUILD     string
	VOICE_CHANNEL   string
	VOICE_COMMANDER string
)

func init() {
	token = os.Getenv("TOKEN")
	user = os.Getenv("USER_TO_MIMIC")
	channel = os.Getenv("CHANNEL")
	GRPC_SERVER = os.Getenv("GRPC_SERVER")
	VOICE_GUILD = os.Getenv("VOICE_GUILD")
	VOICE_CHANNEL = os.Getenv("VOICE_CHANNEL")
	VOICE_COMMANDER = os.Getenv("VOICE_COMMANDER")
	rand.Seed(time.Now().UnixNano())
}

type Discord struct {
	Session             *discordgo.Session
	Sessions            []*discordgo.Session
	Guilds              []*discordgo.Guild
	OwnerUserID         string
	ApplicationClientID string
	ChannelToTalkOn     string
	userToMimic         string
}

func main() {
	var (
		d   Discord
		err error
	)

	log.Printf("Token: %s\nUser: %s\nChannel:%s\nPort: %s\nGuild: %s\nVoice Channel:%s\n", token, user, channel, GRPC_SERVER, VOICE_GUILD, VOICE_CHANNEL)

	d.userToMimic = user
	d.ChannelToTalkOn = channel
	d.Session, err = discordgo.New("Bot " + token)
	if err != nil {
		log.Printf("error creating Discord session %v", err)
		return
	}

	d.Session.AddHandler(messageCreate)

	d.Session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentDirectMessages | discordgo.IntentGuildVoiceStates

	if err := d.Session.Open(); err != nil {
		log.Printf("error opening connection, %v", err)
		return
	}

	fmt.Println("Bot is now running.")
	d.Guilds = d.Session.State.Guilds

	fmt.Println("Bot is: " + d.Session.State.User.Username)
	fmt.Println("Bot is supposed to be: " + username)

	s := scraped.NewScraper(token)
	username = s.GetUserName(user)
	fmt.Println("Bot is: " + d.Session.State.User.Username)
	fmt.Println("Bot is supposed to be: " + username)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	d.Session.Close()
}

func scrape(d Discord) {
	s := scraped.NewScraper(token)
	dbm, err := db.NewManager()
	if err != nil {
		log.Printf("Error Scraping channel: %v", err)
	}
	s.ScrapeChannel(dbm, "879531176166051841", db.DISCORD)
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author != nil && m.Author.Bot && m.Author.Username == username {
		log.Printf("Message from ")
		return
	}
	if m.Author != nil && m.Author.Bot {
		time.Sleep(time.Second * time.Duration(rand.Intn(20)))
	}
	if m.GuildID == "" || m.Member == nil {
		//DIRECT MESSAGE. DO DIRECT MESSAGE THINGS.
		log.Printf("Direct message...")
	}

	if isMentioned(m, s) {
		log.Printf("Mentioned")
		if strings.Contains(m.Content, "BOOT") || strings.Contains(m.Content, "VOICE") {
			guild, err := s.State.Guild(m.GuildID)
			if err != nil {
				log.Printf("Failed to get guild from message... %v", err)
			}
			time.Sleep(time.Second * time.Duration(rand.Intn(10)))
			VoiceMessage(s, guild, m.Author.ID)
			return
		}
	}
	s.UserUpdate(username, "")
	fmt.Println(m.Content)
	response, err := GrabMessage()
	if err != nil {
		log.Printf("Response error: %v", err)
		return
	}

	s.ChannelTyping(m.ChannelID)
	for i := rand.Intn(11); i > 0; i-- {
		time.Sleep(time.Second)
	}
	s.ChannelMessageSend(m.ChannelID, response)
}

func isMentioned(m *discordgo.MessageCreate, s *discordgo.Session) bool {
	if len(m.Mentions) == 0 {
		return false
	}
	if m.Mentions != nil {
		for _, m := range m.Mentions {
			if m.ID == s.State.User.ID {
				return true
			}
		}
	}
	return false
}

func VoiceMessage(s *discordgo.Session, guild *discordgo.Guild, authorID string) {
	var channel string

	for _, vs := range guild.VoiceStates {
		fmt.Println(vs)
		if vs.UserID == authorID {
			channel = vs.ChannelID
			log.Printf(channel)
		}
	}
	v, err := s.ChannelVoiceJoin(guild.ID, VOICE_CHANNEL, false, true)
	fmt.Println(v.ChannelID)
	fmt.Println(v.GuildID)
	if err != nil {
		log.Printf("Error joining voice: %v", err)
		if _, ok := s.VoiceConnections[VOICE_GUILD]; ok {
			v = s.VoiceConnections[VOICE_GUILD]
		} else {
			log.Printf("Still errors...")
		}
	}
	if !v.Ready || v == nil {
		log.Printf("Failed to send voice message.")
		//return
	}

	resp, err := GrabMessage()
	if err != nil {
		log.Printf("Error grabbing message...")
		return
	}
	tts.GetVoice(v, resp)

	v.Disconnect()
	v.Close()
}

func GrabMessage() (string, error) {
	conn, err := grpc.Dial(GRPC_SERVER, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Printf("Error connecting to Markov server: %v\n", err)
	}
	defer conn.Close()
	client := pb.NewMessagesClient(conn)
	request := pb.MsgRequest{
		Signal: "",
	}
	response, err := client.Detect(context.Background(), &request)
	resp := strings.TrimPrefix(response.Msg, "'")
	resp = strings.TrimSuffix(resp, "'")
	return resp, err
}
