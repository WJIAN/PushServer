package util

import (
	"os"
	"io/ioutil"
	"encoding/binary"

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
