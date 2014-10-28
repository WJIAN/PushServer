package util

import (
	"os"
	"net"
	"time"
	"io/ioutil"
	"encoding/binary"
	"strings"
	"errors"
	"hash/fnv"
	"compress/gzip"
	"bytes"
	"fmt"
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

func GetInterIp() (string, error) {
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

func GetLocalIp() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}


	for _, addr := range addrs {
		//fmt.Printf("Inter %v\n", addr)
		ip := addr.String()
		if "127." == ip[:4] {
			return strings.Split(ip, "/")[0], nil
		}

	}

	return "", errors.New("no local ip")
}



func GetExterIp() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}


	for _, addr := range addrs {
		//fmt.Printf("Inter %v\n", addr)
		ip := addr.String()
		if "10." != ip[:3] && "172." != ip[:4] && "196." != ip[:4] && "127." != ip[:4] {
			return strings.Split(ip, "/")[0], nil
		}

	}

	return "", errors.New("no exter ip")
}



func Strhash(s string) uint32 {
    h := fnv.New32a()
    h.Write([]byte(s))
    return h.Sum32()
}

func GzipBytes(bs []byte) ([]byte, error) {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	n, err := w.Write(bs)
	if err != nil {
		return nil, err
	}

	w.Close()

	if n != len(bs) {
		return nil, errors.New("gzip incomplete")
	}

	return b.Bytes(), nil

}


func UngzipBytes(bs []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(bs))

	if err != nil {
		return nil, err
	}
	defer r.Close()

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}


	return b, nil


}


var (
	Since2014 int64 = time.Date(2014, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano() / 1000
)


func Timestamp2014() uint64 {
	return uint64(time.Now().UnixNano()/1000 - Since2014)

}

func PackageSplit(conn net.Conn, readtimeout int, readCall func([]byte)) (bool, error) {
	//fun := "packSplit"
	buffer := make([]byte, 2048)
	packBuff := make([]byte, 0)
	var bufLen uint64 = 0

	for {
		conn.SetReadDeadline(time.Now().Add(time.Duration(readtimeout) * time.Second))
		bytesRead, error := conn.Read(buffer)
		if error != nil {
			//slog.Infof("%s client:%s conn error: %s", fun, self, error)
			return true, error
		}



		packBuff = append(packBuff, buffer[:bytesRead]...)
		bufLen += uint64(bytesRead)


	    //slog.Infof("%s client:%s Recv: %d %d %d", fun, self, bytesRead, packBuff, bufLen)

		for {
			if (bufLen > 0) {
			    pacLen, sz := binary.Uvarint(packBuff[:bufLen])
				if sz < 0 {
					//slog.Warnf("%s client:%s package head error:%s", fun, self, packBuff[:bufLen])
					return false, errors.New(fmt.Sprintf("package head error var:%v", packBuff[:bufLen]))
				} else if sz == 0 {
				    break
				}

				//slog.Debugf("%s client:%s pacLen %d", fun, self, pacLen)
				// must < 5K
				if pacLen > 1024 * 5 {
					//slog.Warnf("%s client:%s package too long error:%s", fun, self, packBuff[:bufLen])
					return false, errors.New(fmt.Sprintf("package too long var:%v", packBuff[:bufLen]))
				} else if pacLen == 0 {
					return false, errors.New(fmt.Sprintf("package len 0 var:%v", packBuff[:bufLen]))

				}

				apacLen := uint64(sz)+pacLen+1
				if bufLen >= apacLen {
				    pad := packBuff[apacLen-1]
					if pad != 0 {
						//slog.Warnf("%s client:%s package pad error:%s", fun, self, packBuff[:bufLen])
						return false, errors.New(fmt.Sprintf("package pad error var:%v", packBuff[:bufLen]))
					}
				    //self.proto(packBuff[sz:apacLen-1])
					readCall(packBuff[sz:apacLen-1])
					packBuff = packBuff[apacLen:]
					bufLen -= apacLen
				} else {
					break
				}

			} else {
				break

			}

		}

	}

	return false, errors.New("unknown err")


}
