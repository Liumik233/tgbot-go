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
	p1.SupportsAllDrives(true)
	_, err := p1.Do()
	if err != nil {
		log.Println(err)
		return 1
	}
	return 0
}
func duser(srv *drive.Service, fileid string, email string) int {
	l := srv.Permissions.List(fileid)
	l.SupportsAllDrives(true)
	l1, err := l.Do()
	if err != nil {
		log.Println(err)
		return 1
	}
	for i := 0; i < len(l1.Permissions); i++ {
		log.Printf(l1.Permissions[i].EmailAddress)
		if l1.Permissions[i].EmailAddress == email {
			call := srv.Permissions.Delete(fileid, l1.Permissions[i].Id)
			call.SupportsAllDrives(true)
			err := call.Do()
			if err != nil {
				log.Println(err)
				log.Printf("here")
				return 1
			}
		}
	}
	return 0
}

func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
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
	conf1 := struct {
		Fileid string `Fileid`
		Token  string `Token`
	}{}
	if Exists("conf.json") {
		c1, err := os.OpenFile("./conf.json", os.O_RDONLY, 0600)
		defer c1.Close()
		if err != nil {
			log.Fatal("openfile error:", err)
		}
		confs, err := ioutil.ReadAll(c1)
		json.Unmarshal(confs, &conf1)
	} else {
		fmt.Println("Input Teamdrive ID:")
		fmt.Scan(&conf1.Fileid)
		fmt.Println("Input Tgbot Token:")
		fmt.Scan(&conf1.Token)
		tmp1, _ := json.Marshal(conf1)
		c1, err := os.Create("conf.json")
		defer c1.Close()
		if err != nil {
			log.Println("create conf.json error:", err)
			os.Exit(1)
		}
		c1.Write(tmp1)
	}
	tb1, err := tb.NewBot(tb.Settings{
		Token:  conf1.Token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
		return
	}
	tb1.Handle(tb.OnUserJoined, func(m *tb.Message) {

		tb1.Reply(m, "欢迎加入新番计划，在群里发送\"/join 邮箱\"即可加入团队盘")
	})
	tb1.Handle("/join", func(m *tb.Message) {
		if strings.HasSuffix(string(m.Chat.Type), "group") {
			if m.Payload == "" {
				tb1.Reply(m, "请填写邮箱！！")
			} else {
				if strings.HasSuffix(m.Payload, "gmail.com") {
					if cuser(srv, m.Payload, conf1.Fileid) == 0 {
						tb1.Reply(m, "添加成功！")
					} else {
						tb1.Reply(m, "添加失败,请检查邮箱是否填写正确！")
					}
				} else {
					tb1.Reply(m, "禁止非gmail邮箱！！")
				}
			}
		}
	})
	duser(srv, conf1.Fileid, "liumikjp@gmail.com")
	tb1.Start()
}
