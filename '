package utils

import (
	"fmt"
	"net/http"
	"io/ioutil"
)

func Check(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func Fetch(url string) []byte {
	res, err := http.Get(url)
	Check(err)
	defer res.Body.Close()
	bodyData, parseBodyErr := ioutil.ReadAll(res.Body)
	Check(parseBodyErr)
	return bodyData
}
