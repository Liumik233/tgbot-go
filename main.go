package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2/google"
	"io/ioutil"
	"strings"

	//"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	//"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	tb "gopkg.in/tucnak/telebot.v2"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func cuser(srv *drive.Service, email string, fileid string) int {
	per := drive.Permission{
		EmailAddress: email,
		Type:         "user", Role: "reader",
	}
	p1 := srv.Permissions.Create(fileid, &per)
	p1.SupportsTeamDrives(true)
	_, err := p1.Do()
	if err != nil {
		log.Println(err)
		return 1
	}
	return 0
}

func main() {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)
	srv, err := drive.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}
	var fileid, token string
	fmt.Println("Input Teamdrive ID:")
	fmt.Scan(&fileid)
	fmt.Println("Input Tgbot Token:")
	fmt.Scan(&token)
	tb1, err := tb.NewBot(tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
		return
	}
	/*tb1.Handle(tb.OnCallback, func(m *tb.Message) {
		tb1.Send(m.Sender, "欢迎加入新番计划，在群里发送/join 邮箱即可加入团队盘")
	})*/
	tb1.Handle(tb.OnText, func(m *tb.Message) {
		if strings.HasPrefix(m.Text, "/join") {
			str := strings.Replace(m.Text, " ", "", -1)
			if cuser(srv, strings.TrimPrefix(str, "/join"), fileid) == 0 {
				tb1.Send(m.Sender, "添加成功")
				log.Println(err)
			} else {
				tb1.Send(m.Sender, "添加失败")
			}
		}
	})
	tb1.Start()
}
