package crypt

import (
	`encoding/hex`
	`github.com/alistanis/goenc`
	`go-micro.dev/v4/config/reader`
)

type config struct {
	Secret string `json:"secret"`
}

var params config

func Configure(r reader.Value) error {
	err := r.Scan(&params)
	if nil != err {
		return err
	}
	return nil
}

func Encode(value string) (res string, err error) {
	
	c, err := goenc.NewCipher(goenc.CBC, goenc.InteractiveComplexity)
	if nil != err {
		return
	}
	item, err := c.Encrypt([]byte(params.Secret), []byte(value))
	if err != nil {
		return
	}
	res = hex.EncodeToString(item)
	return
	
}
func Decode(data string) (res string, err error) {
	c, err := goenc.NewCipher(goenc.CBC, goenc.InteractiveComplexity)
	
	dec, err := hex.DecodeString(data)
	if nil != err {
		return
	}
	item, err := c.Decrypt([]byte(params.Secret), dec)
	if nil != err {
		return
	}
	res = string(item)
	return
}
