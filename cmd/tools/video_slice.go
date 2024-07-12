package tools

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"nas-knif/utils/alist"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

type Options struct {
	AlistHost        string
	TsFilePath       string
	KeyPath          string
	SourceFile       string
	TargetFolderName string
	AuthKey          string
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
			if len(opts.AuthKey) == 0 {
				return fmt.Errorf("AuthKey别空啊，我要提取签名文件")
			}
			if len(opts.TargetFolderName) == 0 {
				_, fileName := filepath.Split(opts.SourceFile)
				opts.TargetFolderName = strings.TrimSuffix(fileName, path.Ext(fileName))
			}
			return slice(opts.AlistHost, opts.AuthKey, opts.TsFilePath, opts.KeyPath, opts.SourceFile, opts.TargetFolderName)
		},
	}
	cmd.Flags().StringVar(&opts.AlistHost, "alist", "alist", "alist路径")
	cmd.Flags().StringVarP(&opts.TsFilePath, "target", "t", "", "目标存储路径")
	cmd.Flags().StringVarP(&opts.KeyPath, "keyPath", "k", "", "目标Key存储路径")
	cmd.Flags().StringVarP(&opts.SourceFile, "source", "s", "", "源文件")
	cmd.Flags().StringVarP(&opts.TargetFolderName, "folder", "f", "", "目录名")
	cmd.Flags().StringVarP(&opts.AuthKey, "auth", "a", "", "Alist令牌")
	return cmd
}
func slice(alistHost string, alistToken string, tsFilePath string, keyPath string, sourceFile string, targetFolderName string) error {
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
	err = ExecCmd(cmd)
	if err != nil {
		return err
	}

	cmd = exec.Command("rclone", "-P", "copy", "out.m3u8", "webdav:"+keyPath+targetFolderName)
	err = ExecCmd(cmd)
	if err != nil {
		return err
	}
	cmd = exec.Command("rclone", "-P", "copy", "./encipher.key", "webdav:"+keyPath+targetFolderName)
	err = ExecCmd(cmd)
	if err != nil {
		return err
	}

	sign, err := alist.GetFileDetail(alistHost, alistToken, keyPath+targetFolderName+"/out.m3u8")
	if err = os.WriteFile("./movie.strm", []byte(alistHost+keyPath+targetFolderName+"/out.m3u8?sign="+sign.Data.Sign), 0666); err != nil {
		log.Fatal(err)
	}
	// 数据拷贝
	cmd = exec.Command("rclone", "-P", "copy", "movie.strm", "webdav:"+keyPath+targetFolderName)
	err = ExecCmd(cmd)
	if err != nil {
		return err
	}

	// 数据拷贝
	cmd = exec.Command("rclone", "-P", "copy", "out.ts", "webdav:"+tsFilePath+targetFolderName)
	err = ExecCmd(cmd)
	if err != nil {
		return err
	}
	return nil

}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
