package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gobuffalo/packr"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	region = "eu-west-1"

	config  *Config
	configM = &sync.Mutex{}

	awsSession *session.Session
	sessionM   = &sync.Mutex{}

	accountStates  = map[string]*AccountState{}
	accountStatesM = &sync.Mutex{}

	messages  = []Message{}
	messagesM = &sync.Mutex{}
)

var (
	configFilename  = kingpin.Flag("config-file", "Config filename").Short('c').Default("config.json").String()
	port            = kingpin.Flag("port", "Server port").Short('p').Default("8000").Int32()
	refreshInterval = kingpin.Flag("refresh-interval", "Refresh interval (seconds)").Short('r').Default("30").Int()
	openBrowser     = kingpin.Flag("open", "Open browser").Bool()
)

func main() {
	kingpin.Parse()

	region = os.Getenv("AWS_REGION")
	if region == "" {
		region = "eu-west-1"
	}

	updateConfig(*configFilename)
	go watchConfig(*configFilename)

	awsConfig := aws.Config{Region: aws.String(region)}
	awsSession = session.New(&awsConfig)

	go doEvery(time.Duration(*refreshInterval)*time.Second, updateAccounts)

	r := mux.NewRouter()

	r.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		jsonify(w, config)
	})

	r.HandleFunc("/api/overview", func(w http.ResponseWriter, r *http.Request) {
		result := []*AccountState{}
		for _, value := range accountStates {
			result = append(result, value)
		}
		jsonify(w, map[string]interface{}{
			"state":    result,
			"accounts": config.Accounts,
			"messages": messages,
		})
	})

	r.PathPrefix("/").Handler(http.FileServer(packr.NewBox("./ui/dist")))

	cors := handlers.AllowedOrigins([]string{"*"})

	if *openBrowser {
		go func() {
			<-time.After(1000 * time.Millisecond)
			exec.Command("xdg-open", fmt.Sprintf("http://localhost:%d", *port)).Run()
		}()
	}

	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), handlers.CORS(cors)(r))
	if err != nil {
		fmt.Println(err)
	}
}
