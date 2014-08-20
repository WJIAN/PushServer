package router

import (
    "fmt"
    "net/http"
    "strings"
//    "log"
//	"io/ioutil"
	"encoding/json"
//	"strconv"

	"code.google.com/p/go-uuid/uuid"
)


// my lib
import (
	"PushServer/util"
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
	slog.Infof("%s path:%s rm:%s", fun, r.URL.Path, r.RemoteAddr)



	remoteip := strings.Split(r.RemoteAddr, ":")

	h := util.Strhash(remoteip[0])

	cf := ProxyConfig[h % uint32(len(ProxyConfig))]

	js, _ := json.Marshal(cf)
	fmt.Fprintf(w, "%s", js)


}


// 获取外网ip，返回非200的code
func installid(w http.ResponseWriter, r *http.Request) {
	fun := "rest.installid"

	slog.Infof("%s path:%s rm:%s", fun, r.URL.Path, r.RemoteAddr)

	uuidgen := uuid.NewUUID()
	installid := uuidgen.String()

	path := strings.Split(r.URL.Path, "/")

	if len(path) != 3 {
		//writeRestErr(w, "uri err")
		http.Error(w, "appid not found", 400)
		return
	}

	// path[0] "", path[1] push
	appid := path[2]

	slog.Infof("%s appid:%s", fun, appid)


	js, _ := json.Marshal(&map[string]string{"installid": installid})
	fmt.Fprintf(w, "%s", js)

}


var ProxyConfig []map[string]string
func StartHttp(httpport string, pc []map[string]string) {
	http.HandleFunc("/route", route)
	http.HandleFunc("/installid/", installid)

	ProxyConfig = pc
	err := http.ListenAndServe(httpport, nil) //设置监听的端口
	if err != nil {
		slog.Panicf("StartHttp ListenAndServe: %s", err)
	}

}
