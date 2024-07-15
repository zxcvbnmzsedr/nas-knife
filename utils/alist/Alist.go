package alist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"io"
	"net/http"
	"os"
	"time"
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
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func PutFile(host string, token string, path string, file []byte) (GetFileDetailResp, error) {
	url := host + "/api/fs/put"

	client := &http.Client{}
	reader := bytes.NewReader(file)

	// create bar
	bar := pb.New(reader.Len()).SetRefreshRate(time.Second).SetWriter(os.Stdout).Set(pb.Bytes, true).Set(pb.SIBytesPrefix, true).Start()
	r := bar.NewProxyReader(reader)

	req, err := http.NewRequest("PUT", url, r)

	if err != nil {
		return GetFileDetailResp{}, err
	}
	req.Header.Add("Authorization", token)
	req.Header.Add("File-Path", path)
	req.Header.Add("User-Agent", "NasKnife/1.0.0")
	req.Header.Add("As-Task", "false")
	req.Header.Add("Content-Type", "text/plain")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return GetFileDetailResp{}, err
	}
	defer res.Body.Close()
	bar.Finish()

	_, err = io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return GetFileDetailResp{}, err
	}
	return GetFileDetail(host, token, path)
}

type ProgressReader struct {
	io.Reader
	Reporter func(r int64)
}

func (pr *ProgressReader) Read(p []byte) (n int, err error) {
	n, err = pr.Reader.Read(p)
	pr.Reporter(int64(n))
	return
}
