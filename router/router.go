package router

import (
    "fmt"
    "io/ioutil"
    "time"
    "net/http"
    "strings"
	"sort"
//    "log"
//	"io/ioutil"
	"encoding/json"
//	"strconv"

	"code.google.com/p/go-uuid/uuid"


	"github.com/shawnfeng/sutil/slog"
)


// my lib
import (
	"PushServer/util"

)

type RestReturn struct {
	// must Cap, so that can get by json Marshal
	Code int  `json:"code,"`
	Err string `json:"err,omitempty"`
	Msgid uint64 `json:"msgid,omitempty"`
	Link string `json:"link,omitempty"`
}

type linkerConf struct {
	stamp int64
	//config map[string]string
	config []byte
}

var (
	linkerTable map[string]*linkerConf = make(map[string]*linkerConf)
	linkerList []*linkerConf = make([]*linkerConf, 0)
)

func isInterReq(r *http.Request) bool {
	hdr := r.Header
	hdrRealIp := hdr.Get("X-Real-Ip")
	hdrForwardedFor := hdr.Get("X-Forwarded-For")

	if hdrRealIp == "" && hdrForwardedFor == "" {
		return true
	} else {
		return false
	}

}

// Request.RemoteAddress contains port, which we want to remove i.e.:
// "[::1]:58292" => "[::1]"
func ipAddrFromRemoteAddr(s string) string {
	idx := strings.LastIndex(s, ":")
	if idx == -1 {
		return s
	}
	return s[:idx]
}

func getIpAddress(r *http.Request) string {
	fun := "getIpAddress"
	hdr := r.Header
	hdrRealIp := hdr.Get("X-Real-Ip")
	hdrForwardedFor := hdr.Get("X-Forwarded-For")

	slog.Infof("%s X-Real-Ip:%s X-Forwarded-For:%s remoteadd:%s", fun, hdrRealIp, hdrForwardedFor, r.RemoteAddr)

	if hdrRealIp == "" && hdrForwardedFor == "" {
		return ipAddrFromRemoteAddr(r.RemoteAddr)
	}
	if hdrForwardedFor != "" {
		// X-Forwarded-For is potentially a list of addresses separated with ","
		parts := strings.Split(hdrForwardedFor, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		// TODO: should return first non-local address
		return parts[0]
	}
	return hdrRealIp
}

type ByString []string

func (s ByString) Len() int {
    return len(s)
}

func (s ByString) Swap(i, j int) {
    s[i], s[j] = s[j], s[i]
}
func (s ByString) Less(i, j int) bool {
    return s[i] < s[j]
}

func linkerCheck() {
	fun := "linkerCheck"
	linkerkey := make([]string, 0)
	now := time.Now().Unix()
	slog.Infof("%s begin %d len %d", fun, now, len(linkerList))
	for k,v := range(linkerTable) {
		slog.Infof("%s %s stamp:%d conf:%s", fun, k, v.stamp, v.config)
		if now - v.stamp > 70 {
			slog.Warnf("%s %s timeout stamp:%d conf:%s", fun, k, v.stamp, v.config)
			delete(linkerTable, k)
		} else {
			linkerkey = append(linkerkey, k)
		}

	}


	sort.Sort(ByString(linkerkey))

	linkertmp := make([]*linkerConf, 0)
	for _, k := range(linkerkey) {
		linkertmp = append(linkertmp, linkerTable[k])
	}


	linkerList = linkertmp
	slog.Infof("%s end len %d", fun, len(linkerList))


}


func route(w http.ResponseWriter, r *http.Request) {
	fun := "rest.route"

	//remoteip := strings.Split(r.RemoteAddr, ":")
	remoteip := getIpAddress(r)

	slog.Infof("%s path:%s rm:%s", fun, r.URL.Path, remoteip)


	if len(linkerList) == 0 {
		http.Error(w, "linker not found", 501)
		return
	}


	h := util.Strhash(remoteip)

	cf := linkerList[h % uint32(len(linkerList))]

	//js, _ := json.Marshal(cf)
	fmt.Fprintf(w, "%s", cf.config)


}


// POST /sublinker/LINKER
// DELETE
func sublinker(w http.ResponseWriter, r *http.Request) {
	fun := "rest.route"

	if !isInterReq(r) {
		http.Error(w, "", 403)
		return
	}

	slog.Infof("%s path:%s method:%s", fun, r.URL.Path, r.Method)

	if r.Method != "POST" && r.Method != "DELETE" {
		http.Error(w, "method err", 405)
		return
	}




	path := strings.Split(r.URL.Path, "/")
	if len(path) != 3 {
		http.Error(w, "uri invalid", 400)
		return
	}
	lk := path[2]

	if r.Method == "DELETE" {
		delete(linkerTable, lk)
		return
	}


	data, err := ioutil.ReadAll(r.Body);
	if err != nil {
		er := fmt.Sprintf("body read err:%s", err)
		http.Error(w, er, 501)
		return
	}

	if len(data) == 0 {
		http.Error(w, "data empty", 400)
		return
	}


	linkerTable[lk] = &linkerConf{
		stamp: time.Now().Unix(),
		config: data,
	}


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



func StartHttp(httpport string) {
	ticker := time.NewTicker(time.Second * 10)
	go func() {
		for {
			select {
			case <-ticker.C:
				linkerCheck()
			}
		}
    }()




	http.HandleFunc("/route", route)
	http.HandleFunc("/installid/", installid)


	http.HandleFunc("/sublinker/", sublinker)

	err := http.ListenAndServe(httpport, nil) //设置监听的端口
	if err != nil {
		slog.Panicf("StartHttp ListenAndServe: %s", err)
	}

}
