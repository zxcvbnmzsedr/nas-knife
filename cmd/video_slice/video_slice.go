package video_slice

import "os/exec"

var alistHost = ""
var TsFilePath = ""
var KeyPath = ""

func slice(sourceFile string, targetFolderName string) {
	// 先创建秘钥
	cmd := exec.Command(" openssl rand 16 > ~/.nas-knife/video_slice/encipher.key")
	err := cmd.Run()
	if err != nil {
		return
	}
	// 获取16位随机字符串
	cmd = exec.Command("openssl rand -hex 16")
	// 16位字符串字符串
	iv, _ := cmd.CombinedOutput()
	// 将这些信息写入到key.keyinfo文件中，第一行为alist的key路径，第二行是秘钥路径，第三行是iv
	cmd = exec.Command("echo", alistHost+KeyPath+targetFolderName+"/encipher.key", "~/.nas-knife/video_slice/encipher.key", string(iv), ">", "~/.nas-knife/video_slice/key.keyinfo")
	err = cmd.Run()
	// 调用ffmpeg进行切片
	//	ffmpeg -y -hwaccel videotoolbox -i ./SSIS-878.mp4 \
	// -vcodec copy -acodec copy \
	//-f hls -hls_time 15 -hls_list_size 0 -hls_key_info_file ./key.keyinfo -hls_playlist_type vod -hls_flags single_file  -hls_base_url https://pan.shiyitopo.tech:7334/d/aliyun/TS/SSIS-878/ out.m3u8
	cmd = exec.Command("ffmpeg", "-y", "-hwaccel", "videotoolbox", "-i", sourceFile,
		"-vcodec", "copy", "-acodec", "copy",
		"-f", "hls", "-hls_time", "15", "-hls_list_size", "0", "-hls_key_info_file", "./key.keyinfo", "-hls_playlist_type", "vod", "-hls_flags", "single_file",
		"-hls_base_url", alistHost+TsFilePath+targetFolderName+"/", targetFolderName+".m3u8")
	if err = cmd.Start(); err != nil {
		return
	}
	err = cmd.Wait()
	if err != nil {
		return
	}
}
