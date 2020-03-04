package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"golang.org/x/net/proxy"
	tb "gopkg.in/tucnak/telebot.v2"
)

type Task struct {
	Pid     int
	CmdText string
	Cmd     *exec.Cmd
}

var (
	Bot   *tb.Bot
	Tasks []Task
)

func (t Task) String() string {
	return fmt.Sprintf("%v, %v", t.Pid, t.CmdText)
}

func init() {
	token := os.Getenv("TELEBOT_TOKEN")
	socks5 := os.Getenv("PROXY_SOCKS5")

	enabledUser, err := strconv.Atoi(os.Getenv("ENABLED_USER"))
	if err != nil {
		panic(err)
	}

	poller := &tb.LongPoller{Timeout: 10 * time.Second}
	restricted := tb.NewMiddlewarePoller(poller, func(upd *tb.Update) bool {
		if upd.Message.Sender.ID != enabledUser {
			log.Printf("Unauthorized access denied for %d.\n", upd.Message.Sender.ID)
			return false
		}
		return true
	})

	botSettings := tb.Settings{
		Token:  token,
		Poller: restricted,
	}

	if socks5 != "" {
		log.Printf("Bot with proxy: %s\n", socks5)

		dialer, err := proxy.SOCKS5("tcp", socks5, nil, proxy.Direct)
		if err != nil {
			log.Fatal("Error creating dialer, aborting.")
		}

		httpTransport := &http.Transport{}
		httpClient := &http.Client{Transport: httpTransport}
		httpTransport.Dial = dialer.Dial
		botSettings.Client = httpClient

	} else {
		log.Printf("Bot with no proxy.\n")
	}

	Bot, err = tb.NewBot(botSettings)

	if err != nil {
		log.Fatal(err)
		return
	}
}
func makeHandle() {
	Bot.Handle("/start", handleHelp)
	Bot.Handle("/help", handleHelp)

	Bot.Handle("/tasks", handleTasks)
	Bot.Handle(tb.OnText, handleExecCommand)
}

func main() {
	makeHandle()
	Bot.Start()
}
