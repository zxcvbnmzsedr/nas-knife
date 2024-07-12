package alist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type GetFileDetailReq struct {
	Path string `json:"path"`
}
type GetFileDetailResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Sign string `json:"sign"`
	}
}

func GetFileDetail(host string, token string, path string) (GetFileDetailResp, error) {

	url := host + "/api/fs/get"
	method := "POST"

	data, _ := json.Marshal(GetFileDetailReq{Path: path})
	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewReader(data))
	resp := GetFileDetailResp{}

	if err != nil {
		return resp, err
	}
	req.Header.Add("Authorization", token)
	req.Header.Add("User-Agent", "NasKnife/1.0.0")
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return resp, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return resp, err
	}
	fmt.Println(string(body))

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return resp, err
	}
	return resp, nil
}
