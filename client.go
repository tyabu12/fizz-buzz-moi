// +build ignore

package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"runtime/pprof"
	"strconv"

	"github.com/gorilla/websocket"
)

type FirstMsg struct {
	Signal string `json:"signal"`
}

type AnswerMsg struct {
	Answer string `json:"answer"`
}

type RecvMsg struct {
	// quiz
	Number   *int    `json:"number"`
	Previous *string `json:"previous"`

	// result
	Signal  *string `json:"signal"`
	Result  *string `json:"result"`
	Score   *int    `json:"score"`
	Message *string `json:"message"`
}

var scheme = flag.String("scheme", "ws", "scheme")
var host = flag.String("host", "localhost:8080", "host address")
var path = flag.String("path", "/", "path")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")

func FizzBuzzMoi(num int) string {
	const (
		Fizz        = 1
		Buzz        = 1 << 1
		Moi         = 1 << 2
		FizzBuzz    = Fizz | Buzz
		FizzMoi     = Fizz | Moi
		BuzzMoi     = Buzz | Moi
		FizzBuzzMoi = Fizz | Buzz | Moi
	)

	flag := 0

	if num%3 == 0 {
		flag |= Fizz
	}
	if num%5 == 0 {
		flag |= Buzz
	}
	if num%7 == 0 {
		flag |= Moi
	}

	switch flag {
	case Fizz:
		return "Fizz"
	case Buzz:
		return "Buzz"
	case Moi:
		return "Moi"
	case FizzBuzz:
		return "FizzBuzz"
	case FizzMoi:
		return "FizzMoi"
	case FizzBuzzMoi:
		return "FizzBuzzMoi"
	case BuzzMoi:
		return "BuzzMoi"
	default:
		return strconv.Itoa(num)
	}
}

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	u := url.URL{Scheme: *scheme, Host: *host, Path: *path}
	log.Println(u.String())

	c, r, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Println("dial: ", err)
		if r != nil {
			log.Printf("%+v\n", *r)
		}
		log.Fatal()
	}
	defer c.Close()

	// log.Println("FirstMsg")
	firstMsg := FirstMsg{Signal: "start"}
	if err := c.WriteJSON(&firstMsg); err != nil {
		log.Fatal("FirstMsg: ", err)
	}

	// log.Println("QuizMsg")
	for {
		var msg RecvMsg
		if err := c.ReadJSON(&msg); err != nil {
			log.Fatal("read: ", err)
		}
		if msg.Signal != nil {
			log.Println("Result: ", *msg.Result)
			log.Println("Score: ", *msg.Score)
			log.Println("Message: ", *msg.Message)
			break
		}
		// if msg.Previous != nil {
		// 	log.Println("Previous:", *msg.Previous)
		// }
		// log.Println("Number:", *msg.Number)
		var answerMsg AnswerMsg
		answerMsg.Answer = FizzBuzzMoi(*msg.Number)
		// log.Println("Answer:", answerMsg.Answer)
		if err = c.WriteJSON(&answerMsg); err != nil {
			log.Fatal("write: ", err)
		}
	}
}
