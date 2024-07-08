package zerotier

import (
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"os/exec"
)

type Model struct {
	installed  bool
	installing bool
}

func InitialModel() Model {
	installed, _ := hasInstalled()
	return Model{
		installed:  installed,
		installing: false,
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
				//installZerotier()
				return m, tea.Quit
			}
		}
	}
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}
func (m Model) View() string {
	s := "Zerotier: \n"
	if !m.installed {
		s += "尚未安装, 是否进行安装(y/n)?"
		if m.installing {
			s += "安装中....."
			info, _ := installZerotier()
			s += info
		}
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

func installZerotier() (string, error) {
	cmd := exec.Command("bash", "-c", `curl -s https://install.zerotier.com | bash`)
	// 创建用于存储标准输出和标准错误的缓冲区
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	err := cmd.Wait()
	return "", err
}
