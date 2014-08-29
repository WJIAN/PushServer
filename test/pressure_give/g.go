package main


import (
	"os"
	"strconv"
	//"time"
	"fmt"
	"bytes"
    "net/http"
	"io/ioutil"
	"crypto/sha1"
	"crypto/md5"



	"PushServer/slog"

)


var pushHost string

// 业务数据包发送
func restPush(clientid string, sendData []byte) string {
	fun := "restPush"

    //sendTime := time.Now().Format("2006-01-02 15:04:05")
	//sendData := []byte(sendTime+"您您您您您您您您")

	slog.Infof("%s cid:%s len:%d data:%s", fun, clientid, len(sendData), sendData)
	
	client := &http.Client{}
	//url := fmt.Sprintf("http://42.120.4.112:9090/push/%s/0/0", clientid)
	url := fmt.Sprintf("http://%s/push/%s/0/0", pushHost, clientid)
	reqest, err := http.NewRequest("POST", url, bytes.NewReader(sendData))
	if err != nil {
		return fmt.Sprintf("push newreq err:%s", err)
	}

	reqest.Header.Set("Connection","Keep-Alive")

	response, err := client.Do(reqest)

	if err != nil {
		return fmt.Sprintf("push doreq err:%s", err)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		slog.Errorf("%s cid:%s Push return ERROR %s", fun, clientid, err)
		return fmt.Sprintf("push err:%s", err)
	}


	if response.StatusCode == 200 {
		slog.Infof("%s cid:%s Push return %s", fun, clientid, body)
		return string(body)

	} else {
		slog.Errorf("%s cid:%s Push ERROR", fun, clientid)
		return fmt.Sprintf("push errcode:%d body:%s", response.StatusCode, body)

	}


}


func main() {
	slog.Init("")
	pushHost = os.Args[1]


	offset, err := strconv.Atoi(os.Args[2])
	if err != nil {
		slog.Panicln("arg not offset count", err)
	}

	clientCount, err := strconv.Atoi(os.Args[3])
	if err != nil {
		slog.Panicln("arg not client count", err)
	}

	slog.Infoln(pushHost, offset, clientCount)

	for {
		for i := offset; i < offset+clientCount; i++ {
			h := sha1.Sum([]byte(fmt.Sprintf("%d", i)))
			installid := fmt.Sprintf("%x", h)

			h1 := md5.Sum([]byte("shawn"+installid+"24ffd40775b15129c3ce9211853560d1"))
			cid := fmt.Sprintf("%x", h1)

		


			rv := restPush(cid, []byte(fmt.Sprintf("牛逼不是吹的:%d", i)))
			slog.Infof("Push %d rv:%s", i, rv)

			//time.Sleep(time.Millisecond * time.Duration(0))

		}

	}
	




}
