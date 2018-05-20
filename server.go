package main

import (
	"encoding/json"
	"fmt"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

const mongoURI = "mongodb://localhost:27017/firefly"

type Poll struct {
	ID        bson.ObjectId `bson:"_id"`
	Num       uint          `bson:"num"`
	TimeStamp time.Time     `bson:"timestamp"`
	LimitTime int           `bson:"limittime"`

	Title       string       `bson:"title"`
	Description string       `bson:"description"`
	Msg4Vote    string       `bson:"msg4vote"`
	Msg4Shr     string       `bson:"msg4shr"`
	ImgUrl      template.URL `bson:"imgurl"`
	Link        template.URL `bson:"link"`

	ViewCnt int `bson:"viewcnt"`
	LikeCnt int `bson:"likecnt"`
	CmtCnt  int `bson:"cmtcnt"`
	ShrCnt  int `bson:"shrcnt"`

	BtnTitle string       `bson:"btntitle"`
	BtnUrl   template.URL `bson:"btnurl"`

	ReactTitles  []string       `bson:"reacttitles"`
	ReactTargets []string       `bson:"reacttargets"`
	ReactUsers   map[string]int `bson:"reactusers"`
	ReactCnt     []int          `bson:"reactcnt"`
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

func GetCurrentResult(Titles []string, Cnts []int) string {
	Size := len(Titles)
	var VoteResult string
	for i := 0; i < Size; i++ {
		VoteResult = VoteResult + Titles[i] + " : " + strconv.Itoa(Cnts[i]) + "\n"
	}
	VoteResult = VoteResult + "\n"

	return VoteResult
}

var DefaultMessage = "반갑습니다. 버튼을 눌러주세요."
var ShareMessage = "공유해서 의견 모으기"
var ShowOriginMessage = "최신 안건 보기"

func UIKeyboard(w http.ResponseWriter, r *http.Request) {

	var aKey sKeyboard
	aKey.Type = "buttons"
	aKey.Buttons = make([]string, 1)
	aKey.Buttons[0] = DefaultMessage

	jData, err := json.Marshal(aKey)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jData)
}

func UIMessage(w http.ResponseWriter, r *http.Request) {
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

	var aPoll Poll
	collection.Find(nil).Sort("-timestamp").One(&aPoll)

	var aText string
	var aKeyboard sKeyboard
	aKeyboard.Type = "buttons"

	var aPhoto sPhoto
	aPhoto.URL = string(aPoll.ImgUrl)
	aPhoto.Width = 640
	aPhoto.Height = 480
	var aMessage_Button sMessage_Button
	aMessage_Button.Label = aPoll.BtnTitle
	aMessage_Button.URL = string(aPoll.BtnUrl)

	ReactSize := len(aPoll.ReactTitles)
	TargetSize := len(aPoll.ReactTargets)

	fmt.Println(msg)
	fmt.Println(aPoll)

	VoteNum, exists := aPoll.ReactUsers[msg.User_key]

	if exists { // 투표했을때
		if msg.Content == DefaultMessage {
			aText = "당신은 " + aPoll.ReactTitles[VoteNum] + " 라고 하셨습니다.\n\n" + GetCurrentResult(aPoll.ReactTitles, aPoll.ReactCnt)
			aKeyboard.Buttons = make([]string, 2)
			aKeyboard.Buttons[0] = ShareMessage
			aKeyboard.Buttons[1] = ShowOriginMessage
		} else if msg.Content == ShowOriginMessage {
			aText = "당신은 " + aPoll.ReactTitles[VoteNum] + "라고 하셨습니다.\n\n" + GetCurrentResult(aPoll.ReactTitles, aPoll.ReactCnt)
			aKeyboard.Buttons = make([]string, 2)
			aKeyboard.Buttons[0] = ShareMessage
			aKeyboard.Buttons[1] = ShowOriginMessage
		} else if msg.Content == ShareMessage { // 공유하기 버튼
			aText = "당신은 " + aPoll.ReactTitles[VoteNum] + "라고 하셨습니다.\n\n" + GetCurrentResult(aPoll.ReactTitles, aPoll.ReactCnt)
			aKeyboard.Buttons = make([]string, 2)
			aKeyboard.Buttons[0] = ShareMessage
			aKeyboard.Buttons[1] = ShowOriginMessage

			aMessage_Button.Label = "공유하기"
			aMessage_Button.URL = "http://49.236.137.51:5000/kakaolink"
		}
	} else { // 투표 안했을 때
		if msg.Content == DefaultMessage || msg.Content == ShowOriginMessage || msg.Content == ShareMessage {
			// 투표액션 없을때 = 투표준비
			aKeyboard.Buttons = make([]string, ReactSize)
			for i := 0; i < ReactSize; i++ {
				aKeyboard.Buttons[i] = aPoll.ReactTitles[i]
			}
		} else {
			// 투표 액션
			for i := 0; i < ReactSize; i++ {
				if msg.Content == aPoll.ReactTitles[i] {
					var TargetStr string
					for j := 0; j < TargetSize; j++ {
						TargetStr = TargetStr + aPoll.ReactTargets[j]
						if j+1 < TargetSize {
							TargetStr = TargetStr + ","
						}
					}
					aPoll.ReactUsers[msg.User_key] = i
					aPoll.ReactCnt[i] += 1
					collection.UpdateId(aPoll.ID, aPoll)

					aText = "감사합니다.\n당신은 " + aPoll.ReactTitles[i] + "에 투표 하셨습니다.\n\n당신의 의견을" + TargetStr + "에게 전달하겠습니다.\n\n" + GetCurrentResult(aPoll.ReactTitles, aPoll.ReactCnt)
					aKeyboard.Buttons = make([]string, 2)
					aKeyboard.Buttons[0] = ShareMessage
					aKeyboard.Buttons[1] = ShowOriginMessage

				}
			}
		}
	}

	if msg.Content == ShareMessage || msg.Content == ShowOriginMessage {
		aText = aText + aPoll.Title + "\n\n" + aPoll.Description + "\n\n" + aPoll.Msg4Shr + "\n"
	} else {
		aText = aText + aPoll.Title + "\n\n" + aPoll.Description + "\n\n" + aPoll.Msg4Vote + "\n"
	}
	fmt.Println(aPoll)

	type sMessage struct {
		Photo          sPhoto          `json:"photo"`
		Text           string          `json:"text"`
		Message_Button sMessage_Button `json:"message_button"`
	}
	aMessage := sMessage{Text: aText, Photo: aPhoto, Message_Button: aMessage_Button}

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

func UIAddFriend(w http.ResponseWriter, r *http.Request) {
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

func UIDeleteFriend(w http.ResponseWriter, r *http.Request) {
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

func UIDeleteChatRoom(w http.ResponseWriter, r *http.Request) {
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
	mux := mux.NewRouter()
	mux.HandleFunc("/keyboard", UIKeyboard).Methods("GET")
	mux.HandleFunc("/message", UIMessage).Methods("POST")

	mux.HandleFunc("/friend", UIAddFriend).Methods("POST")
	mux.HandleFunc("/friend", UIDeleteFriend).Methods("DELETE")

	mux.HandleFunc("/chat_room/:user_key", UIDeleteChatRoom).Methods("DELETE")

	n := negroni.Classic()

	n.UseHandler(mux)

	n.Run(":3000")
}
