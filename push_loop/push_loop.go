package main

import (
	"time"
	"strings"
	"fmt"
	"bytes"
    "net/http"
	"io/ioutil"
	"os"

	"PushServer/slog"
	"PushServer/rediscluster"
)


type sendState struct {
	// push 计数
	count int64
	// 上一次推送rest的返回内容
	rvLast string

}


var redisPool *rediscluster.RedisPool



func getClients() []string {
	fun := "updateMan"
    cmd := []interface{}{
		"keys",
		"I.*",
	}


	mcmd := make(map[string][]interface{})
    mcmd["10.241.221.106:9600"] = cmd

	rp := redisPool.Cmd(mcmd)
	slog.Debugln(fun, "evalsha1", rp)


	clientlist := []string{}
	for addr, r := range(rp) {
		cl, err := r.List()
		if err != nil {
			slog.Errorf("%s addr:%s err:%s", fun, addr, err)
			break
		}

		for _, v := range(cl) {
			clientlist = append(clientlist, strings.Split(v, ".")[1])
		}


		break
	}


	return clientlist

}

// 业务数据包发送
func restPush(clientid string, sendData []byte) string {
	fun := "restPush"

    //sendTime := time.Now().Format("2006-01-02 15:04:05")
	//sendData := []byte(sendTime+"您您您您您您您您")

	slog.Infof("%s cid:%s len:%d data:%s", fun, clientid, len(sendData), sendData)
	
	client := &http.Client{}
	url := fmt.Sprintf("http://42.120.4.112:9090/push/%s/0/0", clientid)
	reqest, _ := http.NewRequest("POST", url, bytes.NewReader(sendData))

	reqest.Header.Set("Connection","Keep-Alive")

	response,_ := client.Do(reqest)
	if response.StatusCode == 200 {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			slog.Errorf("%s cid:%s Push return ERROR %s", fun, clientid, err)
			return fmt.Sprintf("push err:%s", err)
		}

		slog.Infof("%s cid:%s Push return %s", fun, clientid, body)
		return string(body)

	} else {
		slog.Errorf("%s cid:%s Push ERROR", fun, clientid)
		return fmt.Sprintf("push errcode:%s", response.StatusCode)

	}


}

func updateMan(sendMan map[string]*sendState) {
	clis := getClients()
	slog.Infoln(clis)
	for _, c := range(clis) {
		if _, ok := sendMan[c]; !ok {
			sendMan[c] = &sendState {
				count: 0,
				rvLast: "",
			}

		}
	}

	slog.Infof("%s sendMan len:%d", "updateMan", len(sendMan))

}

func traversePush(sendMan map[string]*sendState) {
	joke := `In Go servers, each incoming request is handled in its own goroutine. Request handlers often start additional goroutines to access backends such as databases and RPC services. The set of goroutines working on a request typically needs access to request-specific values such as the identity of the end user, authorization tokens, and the request's deadline. When a request is canceled or times out, all the goroutines working on that request should exit quickly so the system can reclaim any resources they are using.`

	joke = "你好"


	for c, v := range(sendMan) {
		sendTime := time.Now().Format("01-02 15:04:05")
		sendData := fmt.Sprintf("S:%d/D:%s/LR:%s/M:[%s]", v.count, sendTime, v.rvLast, joke)
		v.count++
		v.rvLast = restPush(c, []byte(sendData))
	}




}

func loopPush() {
	sendMan := make(map[string]*sendState)
	ticker := time.NewTicker(time.Second * 60)
	ticker2 := time.NewTicker(time.Second * 60 * 5)


	// 每5分更新一次client id
	updateMan(sendMan)
	go func() {
		for {
			select {
			case <-ticker2.C:
				updateMan(sendMan)
			}
		}
    }()


	// 循环遍历推送
	traversePush(sendMan)
	for {
		select {
		case <-ticker.C:
			traversePush(sendMan)
		}
	}

}



func main() {
    slog.Init(os.Stdout)
	redisPool = rediscluster.NewRedisPool()


	

	//restPush()

	loopPush()
}
