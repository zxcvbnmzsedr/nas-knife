package zerotier

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Model struct {
	installed      bool
	installing     bool
	zerotierStatus string
	joinedNetworks []struct {
		nwid string
		name string
	}
	selectedCursor int
}

func InitialModel() Model {
	installed, _ := hasInstalled()
	zerotierStatus := ""
	var joinedNetworks []struct {
		nwid string
		name string
	}
	if installed {
		zerotierStatus, joinedNetworks = getZerotierStatus()
	}
	return Model{
		installed:      installed,
		installing:     false,
		zerotierStatus: zerotierStatus,
		joinedNetworks: joinedNetworks,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.installed {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "n":
				return m, tea.Quit
			case "y":
				m.installing = true
				installZerotier()
				return m, tea.Quit
			}
		}
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.selectedCursor > 0 {
				m.selectedCursor--
			}
		case "down", "j":
			if m.selectedCursor < len(m.joinedNetworks)-1 {
				m.selectedCursor++
			}
		}
	}
	return m, nil
}
func (m Model) View() string {
	s := "Zerotier: \n"
	if !m.installed {
		s += "尚未安装, 是否进行安装(y/n)?"
	} else {
		s += "已加入的网络: \n" + m.zerotierStatus
	}

	return s
}

func hasInstalled() (bool, error) {
	cmd := exec.Command("zerotier-cli")
	err := cmd.Run()
	if err != nil {
		return false, err
	}
	return true, err
}

func installZerotier() {
	cmd := exec.Command("bash", "-c", `curl -s https://install.zerotier.com | bash`)
	// 创建用于存储标准输出和标准错误的缓冲区
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	cmd.Wait()
}

func getZerotierStatus() (string, []struct {
	nwid string
	name string
}) {
	cmd := exec.Command("zerotier-cli", "listnetworks")
	stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		fmt.Println("Execute failed when Start:" + err.Error())
	}

	out_bytes, _ := io.ReadAll(stdout)
	stdout.Close()

	if err := cmd.Wait(); err != nil {
		fmt.Println("Execute failed when Wait:" + err.Error())
	}

	result := string(out_bytes)
	lines := strings.Split(result, "\n")
	var joinedNetworks []struct {
		nwid string
		name string
	}
	for i, line := range lines {
		if i == 0 {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) > 3 {
			newNetWork := struct {
				nwid string
				name string
			}{
				nwid: parts[2],
				name: parts[3],
			}
			joinedNetworks = append(joinedNetworks, newNetWork)
		}
	}

	return result, joinedNetworks

}
