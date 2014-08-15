package util

import (
	"os"
	"net"
	"io/ioutil"
	"encoding/binary"
	"strings"
	"errors"

)


func PackdataPad(data []byte, pad byte) []byte {
	sendbuff := make([]byte, 0)
	// no pad
	var pacLen uint64 = uint64(len(data))
	buff := make([]byte, 20)
	rv := binary.PutUvarint(buff, pacLen)

	sendbuff = append(sendbuff, buff[:rv]...) // len
	sendbuff = append(sendbuff, data...) //data
	sendbuff = append(sendbuff, pad) //pad

	return sendbuff

}

func Packdata(data []byte) []byte {
	return PackdataPad(data, 0)
}



func GetFile(cfgFile string) ([]byte, error){
	fin, err := os.Open(cfgFile)

	if err != nil {
		return nil, err
	}

	defer fin.Close()

	data, err := ioutil.ReadAll(fin)


	return data, err
}

func GetLocalIp() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}


	for _, addr := range addrs {
		//fmt.Printf("Inter %v\n", addr)
		ip := addr.String()
		if "10." == ip[:3] {
			return strings.Split(ip, "/")[0], nil
		} else if "172." == ip[:4] {
			return strings.Split(ip, "/")[0], nil
		} else if "196." == ip[:4] {
			return strings.Split(ip, "/")[0], nil
		}



	}

	return "", errors.New("no inter ip")
}


