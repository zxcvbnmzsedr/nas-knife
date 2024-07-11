package utils

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"
)

type Options struct {
	AlistHost        string
	TsFilePath       string
	KeyPath          string
	SourceFile       string
	TargetFolderName string
}

func NewVideoSlice() *cobra.Command {
	var opts Options
	cmd := &cobra.Command{
		Use:     "video_slice",
		Aliases: []string{"vl"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(opts.AlistHost) == 0 {
				return fmt.Errorf("host别空啊，用于替换m3u8的实际路径，302跳转用得到！")
			}
			if len(opts.TsFilePath) == 0 {
				return fmt.Errorf("TsFilePath别空啊， 用于存放ts文件")
			}
			if len(opts.KeyPath) == 0 {
				opts.KeyPath = opts.TsFilePath
			}
			if len(opts.SourceFile) == 0 {
				return fmt.Errorf("SourceFile别空啊，要切片的视频文件")
			}
			if len(opts.TargetFolderName) == 0 {
				return fmt.Errorf("TargetFolderName别空啊，不然我咋知道你的番号是啥？？")
			}
			return slice(opts.AlistHost, opts.TsFilePath, opts.KeyPath, opts.SourceFile, opts.TargetFolderName)
		},
	}
	cmd.Flags().StringVar(&opts.AlistHost, "alist", "alist", "alist路径")
	cmd.Flags().StringVarP(&opts.TsFilePath, "target", "t", "", "目标存储路径")
	cmd.Flags().StringVarP(&opts.KeyPath, "keyPath", "k", "", "目标Key存储路径")
	cmd.Flags().StringVarP(&opts.SourceFile, "source", "s", "", "源文件")
	cmd.Flags().StringVarP(&opts.TargetFolderName, "folder", "f", "", "目录名")
	return cmd

}
func slice(alistHost string, tsFilePath string, keyPath string, sourceFile string, targetFolderName string) error {
	if keyPath == "" {
		keyPath = tsFilePath
	}
	// 先创建秘钥
	cmd := exec.Command("sh", "-c", "openssl rand 16 > ./encipher.key")
	err := cmd.Run()
	if err != nil {
		return err
	}
	// 获取16位随机字符串
	cmd = exec.Command("sh", "-c", "openssl rand -hex 16")
	// 16位字符串字符串
	iv, _ := cmd.CombinedOutput()
	// 将这些信息写入到key.keyinfo文件中，第一行为alist的key路径，第二行是秘钥路径，第三行是iv
	if err = os.WriteFile("./key.keyinfo", []byte(alistHost+keyPath+targetFolderName+"/encipher.key\n"+"./encipher.key\n"+string(iv)), 0666); err != nil {
		log.Fatal(err)
	}
	// 调用ffmpeg进行切片
	cmd = exec.Command("ffmpeg", "-y", "-hwaccel", "videotoolbox", "-i", sourceFile,
		"-vcodec", "copy", "-acodec", "copy",
		"-f", "hls", "-hls_time", "15", "-hls_list_size", "0", "-hls_key_info_file", "./key.keyinfo", "-hls_playlist_type", "vod", "-hls_flags", "single_file",
		"-hls_base_url", alistHost+tsFilePath+targetFolderName+"/", "out.m3u8")
	// 命令的错误输出和标准输出都连接到同一个管道
	stdout, err := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout

	if err != nil {
		return err
	}
	if err = cmd.Start(); err != nil {
		return err
	}
	// 从管道中实时获取输出并打印到终端
	for {
		tmp := make([]byte, 1024)
		_, err := stdout.Read(tmp)
		fmt.Print(string(tmp))
		if err != nil {
			break
		}
	}
	if err = cmd.Wait(); err != nil {
		return err
	}

	cmd = exec.Command("rclone", "-P", "copy", "out.m3u8", "webdav:"+keyPath+targetFolderName)
	stdout, err = cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout

	if err != nil {
		return err
	}
	if err = cmd.Start(); err != nil {
		return err
	}
	// 从管道中实时获取输出并打印到终端
	for {
		tmp := make([]byte, 1024)
		_, err := stdout.Read(tmp)
		fmt.Print(string(tmp))
		if err != nil {
			break
		}
	}
	if err = cmd.Wait(); err != nil {
		return err
	}
	cmd = exec.Command("rclone", "-P", "copy", "./encipher.key", "webdav:"+keyPath+targetFolderName)
	stdout, err = cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout

	if err != nil {
		return err
	}
	if err = cmd.Start(); err != nil {
		return err
	}
	// 从管道中实时获取输出并打印到终端
	for {
		tmp := make([]byte, 1024)
		_, err := stdout.Read(tmp)
		fmt.Print(string(tmp))
		if err != nil {
			break
		}
	}
	if err = cmd.Wait(); err != nil {
		return err
	}

	// 数据拷贝
	cmd = exec.Command("rclone", "-P", "copy", "out.ts", "webdav:"+tsFilePath+targetFolderName)
	stdout, err = cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout

	if err != nil {
		return err
	}
	if err = cmd.Start(); err != nil {
		return err
	}
	// 从管道中实时获取输出并打印到终端
	for {
		tmp := make([]byte, 1024)
		_, err := stdout.Read(tmp)
		fmt.Print(string(tmp))
		if err != nil {
			break
		}
	}
	if err = cmd.Wait(); err != nil {
		return err
	}
	return nil

}
