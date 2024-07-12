package docker

import (
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"os/exec"
	"strings"
)

type Model struct {
	// 是否安装
	installed bool
	// 是否启动
	running bool
	// 是否开机启动
	autostart bool
	// 当前版本信息
	versionInfo string
}

func InitialModel() Model {
	installed, versionInfo := getDockerVersion()
	autostart := false
	if installed {
		autostart = isAutoStart()
	}
	return Model{
		installed:   installed,
		versionInfo: versionInfo,
		autostart:   autostart,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	if !m.installed {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "n":
				return m, tea.Quit
			case "y":
				installDocker()
				return m, tea.Quit
			}
		}
	}
	return m, nil
}
func (m Model) View() string {
	s := "Docker安装管理\n"
	if m.installed {
		s += "Docker已安装\n"
		s += m.versionInfo
		if m.autostart {
			s += "Docker已开机启动\n"
		} else {
			s += "Docker未开机启动\n"
		}
	} else {
		s += "Docker未安装\n"
		s += "是否安装Docker (y/n ?)\n"
	}
	return s
}

func getDockerVersion() (bool, string) {
	//	docker version
	cmd := exec.Command("docker", "version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, ""
	}
	return true, string(out)
}
func installDocker() {
	//curl -fsSL https://get.docker.com | bash -s docker --mirror Aliyun
	cmd := exec.Command("bash", "-c", "curl -fsSL https://get.docker.com | bash -s docker --mirror Aliyun")
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func isAutoStart() bool {
	// systemctl is-enabled docker
	cmd := exec.Command("systemctl", "is-enabled", "docker")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "enabled"
}
