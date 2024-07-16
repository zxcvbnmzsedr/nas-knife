package alist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
	} `json:"data"`
}
type TaskInfo struct {
	Id       string  `json:"id"`
	Name     string  `json:"name"`
	State    int     `json:"state"`
	Status   string  `json:"status"`
	Progress float32 `json:"progress"`
	Error    string  `json:"error"`
}
type PutFileResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Task TaskInfo `json:"task"`
	} `json:"data"`
}
type TaskInfoResp struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    TaskInfo `json:"data"`
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
	fmt.Println("文件详情", string(body))
	err = json.Unmarshal(body, &resp)
	if err != nil {
		fmt.Println(err)
		return resp, err
	}
	if resp.Code != 200 {
		return GetFileDetailResp{}, fmt.Errorf("获取文件详情失败: %s", resp.Message)
	}
	return resp, nil
}

func PutFile(host string, token string, path string, file []byte) (GetFileDetailResp, error) {
	fmt.Println("上传文件", path)
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
	req.Header.Add("As-Task", "true")
	req.Header.Add("Content-Type", "text/plain")

	res, err := client.Do(req)
	if err != nil {
		return GetFileDetailResp{}, err
	}
	defer res.Body.Close()
	bar.Finish()
	resp := PutFileResp{}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return GetFileDetailResp{}, err
	}
	err = json.Unmarshal(body, &resp)
	fmt.Println(resp)
	if resp.Code != 200 {
		fmt.Println(resp.Message)
		return GetFileDetailResp{}, fmt.Errorf("上传失败: %s", resp.Message)
	}
	// 获取上传任务
	taskBar := pb.StartNew(100)
	for {
		taskInfoResp := GetTaskProcess(host, token, resp.Data.Task.Id)
		process := int64(taskInfoResp.Data.Progress)
		taskBar.SetCurrent(process)
		if taskInfoResp.Data.State == 2 {
			process = 100
			taskBar.SetCurrent(process)
			taskBar.Finish()
			break
		}
		time.Sleep(time.Millisecond)
	}
	refresh(host, token, path)
	return GetFileDetail(host, token, path)
}

func GetTaskProcess(host string, token string, taskId string) TaskInfoResp {

	url := host + "/api/admin/task/upload/info?tid=" + taskId

	client := &http.Client{}

	req, err := http.NewRequest("POST", url, nil)

	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Authorization", token)
	req.Header.Add("User-Agent", "NasKnife/1.0.0")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()
	resp := TaskInfoResp{}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(body, &resp)

	return resp
}

func refresh(host string, token string, path string) {
	url := host + "/api/fs/get"
	dir, _ := filepath.Split(path)
	a := `{"path":"` + dir + `,"password":"","page":1,"per_page":0,"refresh":true}`
	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, bytes.NewReader([]byte(a)))
	req.Header.Add("Authorization", token)
	req.Header.Add("User-Agent", "NasKnife/1.0.0")
	_, _ = client.Do(req)
	fmt.Println("刷新成功")
}
