package adguard

import (
	tea "github.com/charmbracelet/bubbletea"
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
	return Model{}
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
				return m, tea.Quit
			}
		}
	}
	return m, nil
}
func (m Model) View() string {
	s := "AdGuard\n"
	if m.installed {
		s += "已安装\n"
		s += m.versionInfo
		if m.autostart {
			s += "AdGuard已开机启动\n"
		} else {
			s += "AdGuard未开机启动\n"
		}
	} else {
		s += "AdGuard未安装\n"
		s += "是否安装AdGuard (y/n ?)\n"
	}
	return s
}
