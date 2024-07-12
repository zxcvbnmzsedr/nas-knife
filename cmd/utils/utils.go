package utils

import (
	"fmt"
	"github.com/spf13/cobra"
	"os/exec"
)

func New() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.AddCommand(NewVideoSlice(), NewGif())
	return cmd
}

func ExecCmd(cmd *exec.Cmd) error {
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
	return nil
}
