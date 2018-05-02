package main

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"log"
	"net/http"
)

type sUserMessage struct {
	User_key string `json:"user_key"`
	Type     string `json:"type"`
	Content  string `json:"content"`
}

type sUserADDDELFriend struct {
	User_key string `json:"user_key"`
}

type sPhoto struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type sMessage_Button struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

type sMessage struct {
	Text           string          `json:"text"`
	Photo          sPhoto          `json:"photo"`
	Message_Button sMessage_Button `json:"message_button"`
}

type sKeyboard struct {
	Type    string   `json:"type"`
	Buttons []string `json:"buttons"`
}

func UIKeyboard(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var aKey sKeyboard
	aKey.Type = "buttons"
	aKey.Buttons = make([]string, 4)
	aKey.Buttons[0] = "선택1"
	aKey.Buttons[1] = "선택2"
	aKey.Buttons[2] = "선택3"
	aKey.Buttons[3] = "선택4"

	jData, err := json.Marshal(aKey)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jData)
}

func UIMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	Body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var msg sUserMessage
	err = json.Unmarshal(Body, &msg)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Println(msg)

	var aMessage sMessage
	aMessage.Text = msg.Content + "을 선택하셨습니다."
	aMessage.Photo.URL = "http://img-cdn.ddanzi.com/files/attach/images/4258226/719/796/510/753f8d9231ccd2535584c4a4905fac50.JPG"
	aMessage.Photo.Width = 640
	aMessage.Photo.Height = 480
	aMessage.Message_Button.Label = "다음 연결"
	aMessage.Message_Button.URL = "http://daum.net"

	var aKeyboard sKeyboard
	aKeyboard.Type = "buttons"
	aKeyboard.Buttons = make([]string, 3)
	aKeyboard.Buttons[0] = "선택1"
	aKeyboard.Buttons[1] = "선택2"
	aKeyboard.Buttons[2] = "선택3"

	ReturnMessage := struct {
		Message  sMessage  `json:"message"`
		Keyboard sKeyboard `json:"keyboard"`
	}{
		Message:  aMessage,
		Keyboard: aKeyboard,
	}
	fmt.Println(ReturnMessage)

	jData, err := json.Marshal(ReturnMessage)
	if err != nil {
		fmt.Println("marshal err")
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jData)
}

func UIAddFriend(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// DB에 넣기
	/*
		ReturnMessage := struct {
			Message sText `json:"message"`
			Photo sPhoto `json:"photo"`
		}

		ReturnMessage.Message.Text = "테스트 메세지1"
		ReturnMessage.Photo.URL = "http://cfile233.uf.daum.net/image/276E593D535DE7732B04C3"
		ReturnMessage.Photo.Width = 640
		ReturnMessage.Photo.Height = 480

		jData, err := json.Marshal(ReturnMessage)
		if err != nil {
			panic(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jData)
	*/
	fmt.Fprint(w, "add friend")
}

func UIDeleteFriend(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// DB에서 지우기
	/*
		jData, err := json.Marshal(aKey)
		if err != nil {
			panic(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jData)
	*/
	fmt.Fprint(w, "delete friend")
}

func UIDeleteChatRoom(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// DB에서 지우기
	/*
		jData, err := json.Marshal(aKey)
		if err != nil {
			panic(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jData)
	*/
	fmt.Fprint(w, "delete chat room")
}

func main() {
	router := httprouter.New()
	router.GET("/keyboard", UIKeyboard)
	router.POST("/message", UIMessage)

	router.POST("/friend", UIAddFriend)
	router.DELETE("/friend", UIDeleteFriend)

	router.DELETE("/chat_room/:user_key", UIDeleteChatRoom)

	log.Fatal(http.ListenAndServe(":3000", router))
}
