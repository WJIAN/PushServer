package router

import (
    "fmt"
    "net/http"
//    "strings"
//    "log"
//	"io/ioutil"
	"encoding/json"
//	"strconv"

	"code.google.com/p/go-uuid/uuid"
)


// my lib
import (
	"PushServer/slog"

)

type RestReturn struct {
	// must Cap, so that can get by json Marshal
	Code int  `json:"code,"`
	Err string `json:"err,omitempty"`
	Msgid uint64 `json:"msgid,omitempty"`
	Link string `json:"link,omitempty"`
}



func route(w http.ResponseWriter, r *http.Request) {
	fun := "rest.route"
	slog.Infof("%s %s", fun, r.URL.Path)

	js, _ := json.Marshal(&map[string]string{"heart": "300", "ip": "42.120.4.112", "port":"9988"})
	fmt.Fprintf(w, "%s", js)


}

func installid(w http.ResponseWriter, r *http.Request) {
	fun := "rest.installid"

	slog.Infof("%s %s", fun, r.URL.Path)

	uuidgen := uuid.NewUUID()
	installid := uuidgen.String()


	js, _ := json.Marshal(&map[string]string{"installid": installid})
	fmt.Fprintf(w, "%s", js)

}


func StartHttp(httpport string) {
	http.HandleFunc("/route1", route)
	http.HandleFunc("/installid1", installid)

	err := http.ListenAndServe(httpport, nil) //设置监听的端口
	if err != nil {
		slog.Panicf("StartHttp ListenAndServe: %s", err)
	}

}
