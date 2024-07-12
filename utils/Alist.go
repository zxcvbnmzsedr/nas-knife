package utils

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func GetFileDetail(host string, token string, path string) {

	url := host + "/api/fs/get"
	method := "POST"

	payload := strings.NewReader(`{
		"path": "",
		"password": "",
		"page": 1,
		"per_page": 0,
		"refresh": false
	}`)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Authorization", token)
	req.Header.Add("User-Agent", "NasKnife/1.0.0")
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
