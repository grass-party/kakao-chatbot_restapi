package main

import (
	//  "fmt"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

type Keyboard struct {
	Type    string   `json:"type"`
	Buttons []string `json:"buttons"`
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var aKey Keyboard
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

func main() {
	router := httprouter.New()
	router.GET("/keyboard", Index)

	log.Fatal(http.ListenAndServe(":3000", router))
}
