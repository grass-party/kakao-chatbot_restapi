package main

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"log"
	"net/http"
)

const mongoURI = "mongodb://localhost:27017/firebug"

type Poll struct {
	ID      bson.ObjectId `bson:"_id"`
	Img     string        `bson:"img"`
	Article string        `bson:"article"`
	Link    string        `bson:"link"`
}

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

type sKeyboard struct {
	Type    string   `json:"type"`
	Buttons []string `json:"buttons"`
}

/*
type sMessage struct {
	Text           string          `json:"text"`
	Photo          sPhoto          `json:"photo"`
	Message_Button sMessage_Button `json:"message_button"`
}

type s2Kakao struct {
	Message  sMessage  `json:"message"`
	Keyboard sKeyboard `json:"keyboard"`
}
*/

func UIKeyboard(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var aKey sKeyboard
	aKey.Type = "buttons"
	aKey.Buttons = make([]string, 4)
	aKey.Buttons[0] = "초기키보드1"
	aKey.Buttons[1] = "초기키보드2"
	aKey.Buttons[2] = "초기키보드3"
	aKey.Buttons[3] = "초기키보드4"

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

	session, err := mgo.Dial(mongoURI)
	if err != nil {
		log.Fatalln(err)
	}
	defer session.Close()

	collection := session.DB("").C("poll")

	aPoll := new(Poll)

	collection.Find(nil).One(&aPoll)

	var aText string
	aText = msg.Content + "선택 : " + aPoll.Article
	var aPhoto sPhoto
	aPhoto.URL = aPoll.Img
	aPhoto.Width = 640
	aPhoto.Height = 480
	var aMessage_Button sMessage_Button
	aMessage_Button.Label = "고양고양"
	//	aMessage_Button.URL = "http://49.236.137.51:5000/kakaolink.html"
	aMessage_Button.URL = aPoll.Link

	fmt.Println(aPoll)

	//	var bMessage_Button sMessage_Button
	//	bMessage_Button.Label = "투표"
	//	bMessage_Button.URL = aPoll.Link

	type sMessage struct {
		Text           string          `json:"text"`
		Photo          sPhoto          `json:"photo"`
		Message_Button sMessage_Button `json:"message_button"`
		//		Message_Button2 sMessage_Button `json:"message_button"`
	}
	aMessage := sMessage{Text: aText, Photo: aPhoto, Message_Button: aMessage_Button} //, Message_Button2: aMessage_Button}

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
		aMessage,
		aKeyboard,
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
