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

type FirstMsg struct {
	Signal string `json:"signal"`
}

type QuizMsg struct {
	Number   int     `json:"number"`
	Previous *string `json:"previous"`
}

type LastMsg struct {
	Signal  string `json:"signal"`
	Result  string `json:"result"`
	Score   int    `json:"score"`
	Message string `json:"message"`
}

type AnswerMsg struct {
	Answer string `json:"answer"`
}

var host = flag.String("host", "localhost:8080", "http service address")
var path = flag.String("path", "/", "http service path")

var upgrader = websocket.Upgrader{} // use default options

var success, failure = "success", "failure"

func FizzBuzzMoi(num int) (ans string) {
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
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	// log.Println("FirstMsg")
	var firstMsg FirstMsg
	if err = c.ReadJSON(&firstMsg); err != nil {
		log.Println("read:", err)
		return
	}
	if firstMsg.Signal != "start" {
		log.Println("read:", err)
		return
	}

	// log.Println("QuizMsg")
	numQuiz, numSuccess := 0, 0
	quizMsg := QuizMsg{Previous: nil}
	start := time.Now()

	for (time.Now().Sub(start)).Seconds() < 1 {
		num := rand.Int()
		numQuiz++

		quizMsg.Number = num
		if err = c.WriteJSON(&quizMsg); err != nil {
			log.Println("write:", err)
			return
		}

		ans := FizzBuzzMoi(num)

		var answerMsg AnswerMsg
		if err = c.ReadJSON(&answerMsg); err != nil {
			log.Println("read:", err)
			return
		}
		// log.Println("Answer:", answerMsg.Answer)
		if answerMsg.Answer == ans {
			numSuccess++
			quizMsg.Previous = &success
		} else {
			quizMsg.Previous = &failure
		}
	}

	// log.Println("LastMsg")
	lastMsg := LastMsg{
		Signal: "end",
		Score:  numSuccess,
	}
	if numSuccess == numQuiz {
		lastMsg.Result = success
		lastMsg.Message = fmt.Sprintf("チャレンジ成功です！記録は %d / %d でした",
			numSuccess, numQuiz)
	} else {
		lastMsg.Result = failure
		lastMsg.Message = fmt.Sprintf("チャレンジ失敗です。記録は %d / %d でした",
			numSuccess, numQuiz)
	}
	if err = c.WriteJSON(&lastMsg); err != nil {
		log.Println("write:", err)
		return
	}
}

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	http.HandleFunc(*path, hander)
	log.Fatal(http.ListenAndServe(*host, nil))
}
