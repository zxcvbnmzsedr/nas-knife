package tools

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/grafov/m3u8"
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
	fmt.Println("生成秘钥成功")

	// 获取16位随机字符串
	cmd = exec.Command("sh", "-c", "openssl rand -hex 16")
	// 16位字符串字符串
	iv, _ := cmd.CombinedOutput()

	encipherFileByte, _ := os.ReadFile("encipher.key")
	encipherFile, err := alist.PutFile(alistHost, alistToken, keyPath+targetFolderName+"/encipher.key", encipherFileByte)

	// 将这些信息写入到key.keyinfo文件中，第一行为alist的key路径，第二行是秘钥路径，第三行是iv
	if err = os.WriteFile("./key.keyinfo", []byte(alistHost+"/d"+keyPath+targetFolderName+"/encipher.key?sign="+encipherFile.Data.Sign+"\n"+"./encipher.key\n"+string(iv)), 0666); err != nil {
		log.Fatal(err)
	}
	fmt.Println("生成KeyInfo成功")
	//加密, 不用AES加密了，每次都TM不一样老有重复文件
	encipherTargetFolderName := fmt.Sprintf("%x", md5.Sum([]byte(targetFolderName)))

	// 调用ffmpeg进行切片
	cmd = exec.Command("ffmpeg", "-y", "-hwaccel", "videotoolbox", "-i", sourceFile,
		"-vcodec", "copy", "-acodec", "copy",
		"-f", "hls", "-hls_time", "15", "-hls_list_size", "0", "-hls_key_info_file", "./key.keyinfo", "-hls_playlist_type", "vod", "-hls_flags", "single_file",
		"-hls_base_url", alistHost+"/d"+tsFilePath+"/",
		"out.m3u8")
	err = ExecCmd(cmd)
	if err != nil {
		return err
	}
	// 生成封面图
	cmd = exec.Command("ffmpeg", "-i", sourceFile, "-y", "-f", "image2", "-frames:", "1", "poster.jpg")
	err = ExecCmd(cmd)
	if err != nil {
		return err
	}

	posterFileByte, _ := os.ReadFile("poster.jpg")
	_, err = alist.PutFile(alistHost, alistToken, keyPath+targetFolderName+"/poster.jpg", posterFileByte)
	if err != nil {
		return err
	}

	// 上传ts文件
	tsFileByte, err := os.ReadFile("out.ts")
	if err != nil {
		return err
	}
	tsFile, err := alist.PutFile(alistHost, alistToken, tsFilePath+encipherTargetFolderName+".ts", tsFileByte)
	if err != nil {
		return err
	}

	m3u8FileByte, _ := os.ReadFile("out.m3u8")
	p, listType, _ := m3u8.DecodeFrom(bytes.NewReader(m3u8FileByte), true)
	switch listType {
	case m3u8.MEDIA:
		mediapl := p.(*m3u8.MediaPlaylist)
		// 替换生成的视频文件地址为实际地址
		for i := range mediapl.Segments {
			if mediapl.Segments[i] != nil {
				mediapl.Segments[i].URI = strings.Replace(mediapl.Segments[i].URI, "out.ts", encipherTargetFolderName+".ts", -1) + "?sign=" + tsFile.Data.Sign
			}
		}
		m3u8File, err := alist.PutFile(alistHost, alistToken, keyPath+targetFolderName+"/out.m3u8", mediapl.Encode().Bytes())
		if err != nil {
			return err
		}

		// 生成strm文件，并上传
		if err = os.WriteFile("./movie.strm", []byte(alistHost+"/d"+keyPath+targetFolderName+"/out.m3u8?sign="+m3u8File.Data.Sign), 0666); err != nil {
			log.Fatal(err)
		}
		movieStrmFile, _ := os.ReadFile("movie.strm")
		_, err = alist.PutFile(alistHost, alistToken, keyPath+targetFolderName+"/movie.strm", movieStrmFile)
		if err != nil {
			return err
		}
	default:
		panic("unhandled default case")
	}

	// 清理文件
	_ = os.Remove("encipher.key")
	_ = os.Remove("out.ts")
	_ = os.Remove("out.m3u8")
	_ = os.Remove("key.keyinfo")
	_ = os.Remove("movie.strm")
	_ = os.Remove("poster.jpg")
	return nil

}
