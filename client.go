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

type FirstSignal struct {
	Signal string `json:"signal"`
}

type Answer struct {
	Answer string `json:"answer"`
}

type RecvData struct {
	// Question
	Number   *int    `json:"number"`
	Previous *string `json:"previous"`

	// Result
	Signal  *string `json:"signal"`
	Result  *string `json:"result"`
	Score   *int    `json:"score"`
	Message *string `json:"message"`
}

var scheme = flag.String("scheme", "ws", "scheme")
var host = flag.String("host", "localhost:8080", "host address")
var path = flag.String("path", "/", "path")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")

func fizzBuzzMoi(num int) string {
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

	// log.Println("FirstSignal")
	signal := FirstSignal{Signal: "start"}
	if err := c.WriteJSON(&signal); err != nil {
		log.Fatal("FirstSignal: ", err)
	}

	// log.Println("QuizMsg")
	for {
		var msg RecvData
		if err := c.ReadJSON(&msg); err != nil {
			log.Fatal("read: ", err)
		}
		if msg.Signal != nil && *msg.Signal == "end" {
			log.Println("Result: ", *msg.Result)
			log.Println("Score: ", *msg.Score)
			log.Println("Message: ", *msg.Message)
			break
		}
		if msg.Previous != nil {
			log.Println("Result: ", *msg.Previous)
		}
		log.Println("Question: ", *msg.Number)

		var answer Answer
		answer.Answer = fizzBuzzMoi(*msg.Number)
		log.Println("Answer: ", answer.Answer)
		if err = c.WriteJSON(&answer); err != nil {
			log.Fatal("write: ", err)
		}
	}
}
