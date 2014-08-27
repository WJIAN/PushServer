package main


import (
	"os"
	"strconv"

	"PushServer/slog"
	"PushServer/test/client"
)


func main() {
	slog.Init("")
	offset, err := strconv.Atoi(os.Args[1])
	if err != nil {
		slog.Panicln("arg not offset count", err)
	}

	clientCount, err := strconv.Atoi(os.Args[2])
	if err != nil {
		slog.Panicln("arg not client count", err)
	}

	democlient.SetClientOffset(offset)

	for i := 0; i < clientCount-1; i++ {
		go democlient.StartClient()
	}
	democlient.StartClient()
}
