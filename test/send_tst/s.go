package main


import (
	"os"
	"time"
	"fmt"

	"PushServer/slog"
	"PushServer/test/client"
)


func main() {
	slog.Init("")
	router := os.Args[1]

	democlient.SetRouterUrl(router)
	cli := democlient.NewuserClient()
	go cli.Power()

	cn := 0
	for {
		err := cli.SendBuss(0, 0, []byte(fmt.Sprintf("SN:%d 牛逼不是吹的", cn)))
		if err != nil {
			slog.Errorln("SendBuss err", err)
		}
		time.Sleep(time.Second * time.Duration(1))
		cn++
	}
}
