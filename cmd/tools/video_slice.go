package tools

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/grafov/m3u8"
	"github.com/spf13/cobra"
	"io"
	"io/fs"
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
	ClearSource      bool
	PosterImg        bool
}

// validateOptions 检查选项是否有效
func validateOptions(opts Options) error {
	if len(opts.AlistHost) == 0 {
		return fmt.Errorf("host 不能为空，用于替换 m3u8 的实际路径，302 跳转用得到！")
	}
	if len(opts.TsFilePath) == 0 {
		return fmt.Errorf("TsFilePath 不能为空， 用于存放 ts 文件")
	}
	if len(opts.SourceFile) == 0 {
		return fmt.Errorf("SourceFile 不能为空，要切片的视频文件")
	}
	if len(opts.AuthKey) == 0 {
		return fmt.Errorf("AuthKey 不能为空，我要提取签名文件")
	}
	if opts.KeyPath == "" {
		opts.KeyPath = opts.TsFilePath
	}
	return nil
}

// existsOnAlist 检查文件是否在 alist 中存在
func existsOnAlist(opts Options, encipherTargetFolderName string) bool {
	_, existError := alist.GetFileDetail(opts.AlistHost, opts.AuthKey, opts.TsFilePath+encipherTargetFolderName+".ts")
	return existError == nil
}

// isVideoFile 检查文件是否为常见的视频文件类型
func isVideoFile(file string) bool {
	kind := strings.ToLower(path.Ext(file))
	return kind == ".mp4" || kind == ".avi" || kind == ".mkv" || kind == ".flv" || kind == ".wmv"
}
func removeFromAlist(opts Options, encipherTargetFolderName string, targetFolderName string) error {
	if err := alist.RemoveFile(opts.AlistHost, opts.AuthKey, opts.TsFilePath+encipherTargetFolderName+".ts"); err != nil {
		return err
	}
	if err := alist.RemoveFile(opts.AlistHost, opts.AuthKey, opts.KeyPath+targetFolderName); err != nil {
		return err
	}
	return nil
}
func NewVideoSlice() *cobra.Command {
	var opts Options
	cmd := &cobra.Command{
		Use:     "video_slice",
		Aliases: []string{"vl"},
		RunE: func(cmd *cobra.Command, args []string) error {
			// 检查必要的选项是否为空
			if err := validateOptions(opts); err != nil {
				return err
			}
			fileInfo, err := os.Stat(opts.SourceFile)
			if err != nil {
				return fmt.Errorf("sourceFile有点不太对 %s", err.Error())
			}
			if fileInfo.IsDir() {
				files := GetFiles(opts.SourceFile)
				var needVlFiles []string
				for _, file := range files {
					// 检查是否为常见的视频文件类型
					if isVideoFile(file) {
						_, fileName := filepath.Split(file)
						targetFolderName := strings.TrimSuffix(fileName, path.Ext(fileName))
						//加密, 不用AES加密了，每次都TM不一样老有重复文件
						encipherTargetFolderName := fmt.Sprintf("%x", md5.Sum([]byte(targetFolderName)))
						if existsOnAlist(opts, encipherTargetFolderName) {
							var o string
							prompt := &survey.Input{
								Message: file + "文件已经存在，是否替换(y) default n ?",
							}
							err := survey.AskOne(prompt, &o)
							if err != nil {
								return err
							}
							if o == "y" {
								if err := removeFromAlist(opts, encipherTargetFolderName, targetFolderName); err != nil {
									return err
								}
								needVlFiles = append(needVlFiles, file)
							}
						} else {
							needVlFiles = append(needVlFiles, file)
						}
					}
				}
				fmt.Println("上传文件列表", needVlFiles)
				for _, file := range needVlFiles {
					opts.SourceFile = file
					if e := slice(opts); e != nil {
						return e
					}
				}
			} else {
				return slice(opts)
			}
			return err
		},
	}
	cmd.Flags().StringVar(&opts.AlistHost, "alist", "alist", "alist路径")
	cmd.Flags().StringVarP(&opts.TsFilePath, "target", "t", "", "目标存储路径")
	cmd.Flags().StringVarP(&opts.KeyPath, "keyPath", "k", "", "目标Key存储路径")
	cmd.Flags().StringVarP(&opts.SourceFile, "source", "s", "", "源文件")
	cmd.Flags().StringVarP(&opts.TargetFolderName, "folder", "f", "", "目录名")
	cmd.Flags().StringVarP(&opts.AuthKey, "auth", "a", "", "Alist令牌")
	cmd.Flags().BoolVarP(&opts.ClearSource, "clear", "c", false, "上传完是否删除源文件")
	cmd.Flags().BoolVarP(&opts.PosterImg, "posterImg", "p", true, "是否生成封面图")
	return cmd
}
func slice(opts Options) error {
	if len(opts.TargetFolderName) == 0 {
		_, fileName := filepath.Split(opts.SourceFile)
		opts.TargetFolderName = strings.TrimSuffix(fileName, path.Ext(fileName))
	}
	alistHost := opts.AlistHost
	alistToken := opts.AuthKey
	tsFilePath := opts.TsFilePath
	keyPath := opts.KeyPath
	sourceFile := opts.SourceFile
	targetFolderName := opts.TargetFolderName
	fmt.Println("开始处理", opts)
	//加密, 不用AES加密了，每次都TM不一样老有重复文件
	encipherTargetFolderName := fmt.Sprintf("%x", md5.Sum([]byte(opts.TargetFolderName)))

	// 生成秘钥信息
	iv, err := generateHexKey()
	if err != nil {
		return err
	}
	fmt.Println("生成秘钥成功")

	// 上传秘钥文件
	encipherFileByte, _ := os.ReadFile("encipher.key")
	encipherFile, err := alist.PutFileForByte(alistHost, alistToken, keyPath+targetFolderName+"/encipher.key", encipherFileByte)

	// 将这些信息写入到key.keyinfo文件中，第一行为alist的key路径，第二行是秘钥路径，第三行是iv
	if err = os.WriteFile("./key.keyinfo", []byte(alistHost+"/d"+keyPath+targetFolderName+"/encipher.key?sign="+encipherFile.Data.Sign+"\n"+"./encipher.key\n"+iv), 0666); err != nil {
		log.Fatal(err)
		return err
	}
	fmt.Println("生成KeyInfo成功")
	// 调用ffmpeg进行切片
	cmd := exec.Command("ffmpeg", "-y", "-i", sourceFile,
		"-vcodec", "copy", "-acodec", "copy",
		"-f", "hls", "-hls_time", "5", "-hls_list_size", "0", "-hls_key_info_file", "./key.keyinfo", "-hls_playlist_type", "vod", "-hls_flags", "single_file",
		"-hls_base_url", alistHost+"/d"+tsFilePath,
		"out.m3u8")
	fmt.Println("切片命令 ", cmd.String())
	if err = ExecCmd(cmd); err != nil {
		return err
	}
	// 生成封面图
	if opts.PosterImg {
		cmd = exec.Command("ffmpeg", "-i", sourceFile, "-y", "-f", "image2", "-frames:", "1", "poster.jpg")
		fmt.Println("生成封面图 ", cmd.String())
		if err = ExecCmd(cmd); err != nil {
			return err
		}
		posterFileByte, _ := os.ReadFile("poster.jpg")
		_, err = alist.PutFileForByte(alistHost, alistToken, keyPath+targetFolderName+"/poster.jpg", posterFileByte)
		if err != nil {
			return err
		}
	}
	// 上传ts文件
	tsFileByte, err := os.Open("out.ts")
	if err != nil {
		fmt.Println("读取ts文件失败", err.Error())
		return err
	}
	tsFile, err := alist.PutFileForFile(alistHost, alistToken, tsFilePath+encipherTargetFolderName+".ts", tsFileByte)
	if err != nil {
		return err
	}
	err = tsFileByte.Close()
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
		m3u8File, err := alist.PutFileForByte(alistHost, alistToken, keyPath+targetFolderName+"/out.m3u8", mediapl.Encode().Bytes())
		if err != nil {
			return err
		}

		// 生成strm文件，并上传
		if err = os.WriteFile("./movie.strm", []byte(alistHost+"/d"+keyPath+targetFolderName+"/out.m3u8?sign="+m3u8File.Data.Sign), 0666); err != nil {
			log.Fatal(err)
		}
		movieStrmFile, _ := os.ReadFile("movie.strm")
		_, err = alist.PutFileForByte(alistHost, alistToken, keyPath+targetFolderName+"/movie.strm", movieStrmFile)
		if err != nil {
			return err
		}
	default:
		panic("unhandled default case")
	}

	cleanFiles()
	if opts.ClearSource {
		_ = os.Remove(sourceFile)
	}
	return nil
}

// cleanFiles 清理临时文件
func cleanFiles() {
	_ = os.Remove("encipher.key")
	_ = os.Remove("out.ts")
	_ = os.Remove("out.m3u8")
	_ = os.Remove("key.keyinfo")
	_ = os.Remove("movie.strm")
	_ = os.Remove("poster.jpg")
}

func GetFiles(folder string) (filesList []string) {
	err := filepath.Walk(folder, func(path string, file fs.FileInfo, err error) error {
		if !file.IsDir() {
			filesList = append(filesList, filepath.ToSlash(path))
		}
		return nil
	})
	if err != nil {
		return nil
	}
	return filesList
}

func generateHexKey() (string, error) {
	key := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, key)
	if err != nil {
		return "", err
	}

	file, err := os.Create("./encipher.key")
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = file.Write(key)
	if err != nil {
		return "", err
	}
	key = make([]byte, 16)
	_, err = rand.Read(key)
	if err != nil {
		fmt.Println("Error generating random key:", err)
		return "", err
	}
	return fmt.Sprintf("%x", key), nil
}
