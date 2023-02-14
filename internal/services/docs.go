package services

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/option"
	"log"
	"net/http"
	"os"
)

// Retrieves a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Requests a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	defer f.Close()
	if err != nil {
		log.Fatalf("Unable to cache OAuth token: %v", err)
	}
	json.NewEncoder(f).Encode(token)
}

type Docs struct {
	Ctx    context.Context
	Config *oauth2.Config
}

func (d *Docs) Init() error {
	ctx := context.Background()
	d.Ctx = ctx
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Printf("Unable to read client secret file: %v", err)
		return err
	}
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/documents")
	if err != nil {
		log.Printf("Unable to parse client secret file to config: %v", err)
		return err
	}
	d.Config = config
	return nil
}

func (d *Docs) Start(sc *Scraper) error {
	client := getClient(d.Config)
	srv, err := docs.NewService(d.Ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Printf("Unable to retrieve Docs client: %v", err)
		return err
	}
	title := viper.GetString("title")
	doc, err := srv.Documents.Create(&docs.Document{Title: title}).Do()
	if err != nil {
		log.Printf("Unable to create document: %v", err)
		return err
	}
	req := &docs.BatchUpdateDocumentRequest{
		Requests: []*docs.Request{
			{
				InsertTable: &docs.InsertTableRequest{
					Rows:    7,
					Columns: 2,
					Location: &docs.Location{
						Index: 1,
					},
				},
			},
		},
	}
	
	var lastIndex int64 = 37
	arrLen := len(sc.Data)
	myReq := docs.Request{InsertText: &docs.InsertTextRequest{
		Text:     sc.Data[arrLen-1],
		Location: &docs.Location{Index: lastIndex},
	}}
	req.Requests = append(req.Requests, &myReq)

	for i := arrLen - 2; i > -1; i-- {
		if i%2 != 0 {
			insertReq := docs.Request{InsertText: &docs.InsertTextRequest{
				Text:     sc.Data[i],
				Location: &docs.Location{Index: lastIndex - 3},
			}}
			lastIndex -= 3
			req.Requests = append(req.Requests, &insertReq)
		} else {
			insertReq := docs.Request{InsertText: &docs.InsertTextRequest{
				Text:     sc.Data[i],
				Location: &docs.Location{Index: lastIndex - 2},
			}}
			lastIndex -= 2
			req.Requests = append(req.Requests, &insertReq)
		}

	}

	secReq := docs.Request{InsertText: &docs.InsertTextRequest{
		Text:     "Описание",
		Location: &docs.Location{Index: 7},
	}}
	req.Requests = append(req.Requests, &secReq)

	frstReq := docs.Request{InsertText: &docs.InsertTextRequest{
		Text:     "HTTP код ответа",
		Location: &docs.Location{Index: 5},
	}}
	req.Requests = append(req.Requests, &frstReq)
	_, errSaving := srv.Documents.BatchUpdate(doc.DocumentId, req).Do()

	if errSaving != nil {
		log.Printf("Unable to save changes to document: %v", err)
		return errSaving
	}

	return nil
}
