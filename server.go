// +build ignore

package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

type FirstSignal struct {
	Signal string `json:"signal"`
}

type Question struct {
	Number   int     `json:"number"`
	Previous *string `json:"previous"`
}

type Result struct {
	Signal  string `json:"signal"`
	Result  string `json:"result"`
	Score   int    `json:"score"`
	Message string `json:"message"`
}

type Answer struct {
	Answer string `json:"answer"`
}

var host = flag.String("host", "localhost:8080", "http service address")
var path = flag.String("path", "/", "http service path")

var upgrader = websocket.Upgrader{} // use default options

var success, failure = "success", "failure"

func fizzBuzzMoi(num int) (ans string) {
	if num%3 == 0 {
		ans += "Fizz"
	}
	if num%5 == 0 {
		ans += "Buzz"
	}
	if num%7 == 0 {
		ans += "Moi"
	}
	if ans == "" {
		ans = strconv.Itoa(num)
	}
	return
}

func hander(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade: ", err)
		return
	}
	defer c.Close()

	// log.Println("FirstSignal")
	var signal FirstSignal
	if err = c.ReadJSON(&signal); err != nil {
		log.Println("read:", err)
		return
	}
	if signal.Signal != "start" {
		log.Println("read:", err)
		return
	}

	log.Println("Question Time Start.")

	numQuestions, numSuccesses := 0, 0
	question := Question{Previous: nil}
	start := time.Now()

	for (time.Now().Sub(start)).Seconds() < 1 {
		num := rand.Int()
		numQuestions++

		question.Number = num
		if err = c.WriteJSON(&question); err != nil {
			log.Println("write:", err)
			return
		}

		ans := fizzBuzzMoi(num)

		var answer Answer
		if err = c.ReadJSON(&answer); err != nil {
			log.Println("read:", err)
			return
		}
		// log.Println("Answer:", answer.Answer)
		if answer.Answer == ans {
			numSuccesses++
			question.Previous = &success
		} else {
			question.Previous = &failure
		}
	}

	// log.Println("ResultMessage")
	result := Result{
		Signal: "end",
		Score:  numSuccesses,
	}
	if numSuccesses == numQuestions {
		result.Result = success
		result.Message = fmt.Sprintf("チャレンジ成功です！記録は %d / %d でした",
			numSuccesses, numQuestions)
	} else {
		result.Result = failure
		result.Message = fmt.Sprintf("チャレンジ失敗です。記録は %d / %d でした",
			numSuccesses, numQuestions)
	}
	if err = c.WriteJSON(&result); err != nil {
		log.Println("write:", err)
		return
	}

	log.Println("Question Time Finish.")
}

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	http.HandleFunc(*path, hander)
	log.Fatal(http.ListenAndServe(*host, nil))
}
