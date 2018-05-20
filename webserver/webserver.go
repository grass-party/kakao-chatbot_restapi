package main

import (
	//	"encoding/json"
	"fmt"
	//	"io/ioutil"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const mongoURI = "mongodb://localhost:27017/firefly"

var renderer *render.Render

var session *mgo.Session

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

func init() {
	var err error

	session, err = mgo.Dial(mongoURI)
	if err != nil {
		log.Fatalln(err)
	}
	defer session.Close()

	// 렌더러 생성
	renderer = render.New()
}

func IndexHandler(w http.ResponseWriter, req *http.Request) {
	var aPoll Poll

	var err error
	session, err = mgo.Dial(mongoURI)
	if err != nil {
		log.Fatalln(err)
	}
	defer session.Close()

	collection := session.DB("").C("poll")

	err = collection.Find(nil).Sort("-timestamp").One(&aPoll)
	if err != nil {
		panic(err)
	}

	renderer.HTML(w, http.StatusOK, "index", aPoll)
}

func KakaoLinkHandler(w http.ResponseWriter, req *http.Request) {
	var aPoll Poll

	var err error
	session, err = mgo.Dial(mongoURI)
	if err != nil {
		log.Fatalln(err)
	}
	defer session.Close()

	collection := session.DB("").C("poll")

	err = collection.Find(nil).Sort("-timestamp").One(&aPoll)
	if err != nil {
		panic(err)
	}

	aPoll.ShrCnt += 1
	err = collection.UpdateId(aPoll.ID, aPoll)
	if err != nil {
		fmt.Println("Update Fail")
		log.Fatalln(err)
	}

	renderer.HTML(w, http.StatusOK, "kakaolink", aPoll)
}

func WriteAgendaHandler(w http.ResponseWriter, req *http.Request) {
	// agenda 보여주기
	renderer.HTML(w, http.StatusOK, "writeagenda", nil)
}

func MakeAgendaHandler(w http.ResponseWriter, req *http.Request) {
	var aPoll Poll

	var err error
	session, err = mgo.Dial(mongoURI)
	if err != nil {
		log.Fatalln(err)
	}
	defer session.Close()

	req.ParseForm()

	aPoll.ID = bson.NewObjectId()
	aPoll.TimeStamp = time.Now()
	aPoll.LimitTime, _ = strconv.Atoi(req.Form.Get("limittime"))
	aPoll.Title = req.Form.Get("title")
	aPoll.Description = req.Form.Get("description")
	aPoll.Msg4Vote = req.Form.Get("msg4vote")
	aPoll.Msg4Shr = req.Form.Get("msg4shr")
	aPoll.ImgUrl = template.URL(req.Form.Get("imgurl"))
	aPoll.Link = template.URL(req.Form.Get("link"))

	aPoll.BtnTitle = req.Form.Get("btntitle")
	aPoll.BtnUrl = template.URL(req.Form.Get("btnurl"))

	ReactTitles := strings.Split(req.Form.Get("reacttitles"), ",")
	ReactSize := len(ReactTitles)

	aPoll.ReactTitles = make([]string, ReactSize)
	aPoll.ReactCnt = make([]int, ReactSize)
	for i := 0; i < ReactSize; i++ {
		aPoll.ReactTitles[i] = ReactTitles[i]
		aPoll.ReactCnt[i] = 0
	}
	aPoll.ReactUsers = make(map[string]int)

	ReactTargets := strings.Split(req.Form.Get("reacttargets"), ",")
	TargetSize := len(ReactTargets)
	aPoll.ReactTargets = make([]string, TargetSize)
	for i := 0; i < TargetSize; i++ {
		aPoll.ReactTargets[i] = ReactTargets[i]
	}

	NumCollection := session.DB("").C("numberCnt")

	type Cnt struct {
		ID  string `bson:"_id"`
		Seq uint   `bson:"seq"`
	}
	var aCnt Cnt
	changeInDocument := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"seq": 1}},
		ReturnNew: true,
	}
	NumCollection.Find(nil).Apply(changeInDocument, &aCnt)
	aPoll.Num = aCnt.Seq

	collection := session.DB("").C("poll")

	collection.Insert(aPoll)

	// agenda 보여주기
	renderer.HTML(w, http.StatusOK, "agenda", aPoll)
}

func ShowAgendaHandler(w http.ResponseWriter, req *http.Request) {
	var aPoll Poll

	var err error
	session, err = mgo.Dial(mongoURI)
	if err != nil {
		log.Fatalln(err)
	}
	defer session.Close()

	collection := session.DB("").C("poll")

	id := req.URL.Query()["id"]

	err = collection.FindId(bson.ObjectIdHex(id[0])).One(&aPoll)
	if err != nil {
		panic(err)
	}

	// agenda 보여주기(뭘 기준으로?)
	renderer.HTML(w, http.StatusOK, "agenda", aPoll)
}

func DelAgendaHandler(w http.ResponseWriter, req *http.Request) {
	var err error
	session, err = mgo.Dial(mongoURI)
	if err != nil {
		log.Fatalln(err)
	}
	defer session.Close()

	collection := session.DB("").C("poll")

	id := req.URL.Query()["id"]

	var aPoll Poll
	err = collection.FindId(bson.ObjectIdHex(id[0])).One(&aPoll)
	if err != nil {
		fmt.Println("Find Fail")
		panic(err)
	}

	err = collection.Remove(&aPoll)
	if err != nil {
		fmt.Println("Remove Fail")
		panic(err)
	}

	http.Redirect(w, req.WithContext(req.Context()), "/showagendalist", http.StatusFound)
	// agenda 보여주기(뭘 기준으로?)
	//	renderer.HTML(w, http.StatusOK, "agenda", aPoll)
}

func ShowAgendaListHandler(w http.ResponseWriter, req *http.Request) {
	var Polls []Poll
	var LimitSize = 10

	var err error
	session, err = mgo.Dial(mongoURI)
	if err != nil {
		log.Fatalln(err)
	}
	defer session.Close()

	collection := session.DB("").C("poll")

	fmt.Printf("starts\n")

	pagestrlist := req.URL.Query()["page"]

	var Page int
	if len(pagestrlist) == 0 {
		Page = 1
	} else {
		fmt.Printf("pagestr : ")
		fmt.Println(pagestrlist[0])

		Page, err = strconv.Atoi(pagestrlist[0])
		if err != nil {
			panic(err)
		}
	}

	fmt.Printf("Page : %d\n", Page)

	var SkipSize = (Page - 1) * LimitSize

	var TotalCnt int
	TotalCnt, err = collection.Count()
	if err != nil {
		panic(err)
	}
	fmt.Printf("TotalCnt : %d\n", TotalCnt)

	err = collection.Find(nil).Sort("-timestamp").Skip(SkipSize).Limit(LimitSize).All(&Polls)
	if err != nil {
		panic(err)
	}

	Pagination := 1 + (TotalCnt / LimitSize)

	fmt.Printf("Pagination : %d\n", Pagination)

	var sPages []string
	sPages = make([]string, Pagination)
	for i := 1; i <= Pagination; i++ {
		sPages[i-1] = strconv.Itoa(i)
	}

	fmt.Printf("sPages : ")
	fmt.Println(sPages)

	// agenda 보여주기(뭘 기준으로?)
	renderer.HTML(w, http.StatusOK, "agendalist", struct {
		Datas []Poll
		Pages []string
	}{
		Datas: Polls,
		Pages: sPages,
	})
}

func main() {
	// 라우터 생성
	mux := mux.NewRouter()

	// 핸들러 정의
	mux.HandleFunc("/", IndexHandler).Methods("GET")
	mux.HandleFunc("/kakaolink", KakaoLinkHandler).Methods("GET")
	mux.HandleFunc("/writeagenda", WriteAgendaHandler).Methods("GET")
	mux.HandleFunc("/makeagenda", MakeAgendaHandler).Methods("POST")
	mux.HandleFunc("/showagenda", ShowAgendaHandler).Methods("GET")
	mux.HandleFunc("/delagenda", DelAgendaHandler).Methods("GET")
	mux.HandleFunc("/showagendalist", ShowAgendaListHandler).Methods("GET")

	// negroni 미들웨어 생성
	n := negroni.Classic()

	// file serve
	n.Use(negroni.NewStatic(http.Dir("./wwwroot")))

	// negroni에 router를 핸들러로 등록
	n.UseHandler(mux)

	// 웹 서버 실행
	n.Run(":5000")
}
