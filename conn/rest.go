package connection

import (
    "fmt"
    "net/http"
    "strings"
//    "log"
	"io/ioutil"
	"encoding/json"
	"strconv"


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




func debug_show_request(r *http.Request) {
    fmt.Println(r)

	fmt.Println("Method", r.Method)
	fmt.Println("URL", r.URL)
	fmt.Println("Proto", r.Proto)
	fmt.Println("ProtoMajor", r.ProtoMajor)
	fmt.Println("ProtoMinor", r.ProtoMinor)
	fmt.Println("Header", r.Header)
	fmt.Println("Body", r.Body)
	//var p []byte = make([]byte, 10)
	//rv, err := r.Body.Read(p)
	hah, err := ioutil.ReadAll(r.Body);
	fmt.Println("Body Read", hah, err)
	fmt.Println("ContentLength", r.ContentLength)
	fmt.Println("TransferEncoding", r.TransferEncoding)
	fmt.Println("Close", r.Close)
	fmt.Println("Host", r.Host)
	fmt.Println("Form", r.Form)
	fmt.Println("PostForm", r.PostForm)
	fmt.Println("MultipartForm", r.MultipartForm)
	fmt.Println("Trailer", r.Trailer)
	fmt.Println("RemoteAddr", r.RemoteAddr)
	fmt.Println("RequestURI", r.RequestURI)
	fmt.Println("TLS", r.TLS)


    r.ParseForm()  //解析参数，默认是不会解析的
    fmt.Println(r)
    fmt.Println(r.Form)  //这些信息是输出到服务器端的打印信息
    fmt.Println("path", r.URL.Path)
    fmt.Println("scheme", r.URL.Scheme)
    fmt.Println(r.Form["url_long"])
    for k, v := range r.Form {
        fmt.Println("key:", k)
        fmt.Println("val:", strings.Join(v, ""))
    }


}


func writeRestErr(w http.ResponseWriter, err string) {
	js, _ := json.Marshal(&RestReturn{Code: 1, Err: err})
	fmt.Fprintf(w, "%s", js)

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


// Method: POST
// Uri: /push/CLIENT_ID/ZIPTYPE/DATATYPE
// Data: push data
func push(w http.ResponseWriter, r *http.Request) {
	fun := "rest.push"
	//debug_show_request(r)
	if r.Method != "POST" {
		writeRestErr(w, "method err")
		return
	}

	slog.Infof("%s %s", fun, r.URL.Path)
	path := strings.Split(r.URL.Path, "/")
	//slog.Info("%q", path)

	if len(path) != 5 {
		writeRestErr(w, "uri err")
		return
	}

	// path[0] "", path[1] push
	clientid := path[2]

	ziptype, err := strconv.Atoi(path[3])
	if err != nil {
		writeRestErr(w, "ziptype err")
		return
	}

	datatype, err := strconv.Atoi(path[4])
	if err != nil {
		writeRestErr(w, "datatype err")
		return
	}

	data, err := ioutil.ReadAll(r.Body);
	if err != nil {
		writeRestErr(w, "data err")
		return
	}

	if len(data) == 0 {
		writeRestErr(w, "data empty")
		return
	}


	msgid, link := ConnManager.Send(clientid, int32(ziptype), int32(datatype), data)
	slog.Debugf("%s msgid:%d link:%s", fun, msgid, link)
	js, _ := json.Marshal(&RestReturn{Code: 0, Msgid: msgid, Link: link})
	fmt.Fprintf(w, "%s", js)


}


func setoffline(w http.ResponseWriter, r *http.Request) {
	fun := "rest.setoffline"

	slog.Infof("%s %s", fun, r.URL.Path)

	ConnManager.setOffline()
	js, _ := json.Marshal(&RestReturn{Code: 0})
	fmt.Fprintf(w, "%s", js)

}


func setonline(w http.ResponseWriter, r *http.Request) {
	fun := "rest.setonline"

	slog.Infof("%s %s", fun, r.URL.Path)

	ConnManager.setOnline()
}



func StartHttp(httpport string) {
	go func() {
		http.HandleFunc("/push/", push)
		http.HandleFunc("/route1", route)
		http.HandleFunc("/installid1", installid)
		http.HandleFunc("/setoffline", setoffline)
		http.HandleFunc("/setonline", setonline)

		err := http.ListenAndServe(httpport, nil) //设置监听的端口
		if err != nil {
			slog.Panicf("StartHttp ListenAndServe: %s", err)
		}
	}()
}
