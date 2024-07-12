package utils

import (
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

func NewGif() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gif",
		Aliases: []string{"gif"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return toGif(args[0])
		},
	}
	return cmd
}

func toGif(filePath string) error {
	//ffmpeg -y -i t.mp4 -vf fps=10,scale=-1:-1:flags=lanczos,palettegen o.gif
	cmd := exec.Command("ffmpeg", "-y", "-i", filePath, "-vf", "fps=10,scale=-1:-1:flags=lanczos,palettegen", "o_tmp.gif")
	err := ExecCmd(cmd)
	if err != nil {
		return err
	}
	_, fileName := filepath.Split(filePath)
	gifFileName := strings.TrimSuffix(fileName, path.Ext(fileName))
	// ffmpeg -i t.mp4 -i o.gif -vf "fps=15,scale=256:-1,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse" xxx.gif
	err = ExecCmd(exec.Command("ffmpeg", "-i", filePath, "-i", "o_tmp.gif", "-vf", "fps=15,scale=256:-1,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse", gifFileName+".gif"))
	if err != nil {
		return err
	}
	err = os.Remove("o_tmp.gif")
	if err != nil {
		return err
	}
	return nil
}
