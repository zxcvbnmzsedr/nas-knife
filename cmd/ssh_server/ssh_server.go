package ssh_server

import (
	"bufio"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"os/exec"
	"strings"
)

type Model struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
}

func InitialModel() Model {
	m := Model{
		choices: []string{
			"是否允许ssh登陆",
		},
		selected: make(map[int]struct{}),
	}
	permitRootLoginStatus := getPermitRootLoginStatus()
	if permitRootLoginStatus {
		m.selected[0] = struct{}{}
	}
	return m
}

func (m Model) Init() tea.Cmd {

	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "w":
			if _, ok := m.selected[0]; ok {
				writeRootLoginStatus("yes")
			} else {
				writeRootLoginStatus("no")
			}
			restartSSHServer()
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	s := "配置SSH登录方式\n\n"
	for i, choice := range m.choices {
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}
		checked := " " // not selected
		if _, ok := m.selected[i]; ok {
			checked = "x"
		}
		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}
	s += "\nPress w to save.\n"
	return s
}

func getPermitRootLoginStatus() bool {
	file, err := os.Open("/etc/ssh/sshd_config")
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line) // remove leading and trailing white space
		if strings.HasPrefix(line, "PermitRootLogin") {
			parts := strings.Split(line, " ")
			if len(parts) >= 2 {
				return parts[1] == "yes"
			}
			break
		}
	}
	return false
}

func writeRootLoginStatus(status string) {
	file, err := os.Open("/etc/ssh/sshd_config")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	found := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "PermitRootLogin") {
			line = "PermitRootLogin " + status
			found = true
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}

	// If PermitRootLogin was not found, add it
	if !found {
		lines = append(lines, "\nPermitRootLogin "+status)
	}

	output := strings.Join(lines, "\n")
	err = os.WriteFile("/etc/ssh/sshd_config.cp", []byte(output), 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func restartSSHServer() {
	cmd := exec.Command("mv", "/etc/ssh/sshd_config.cp", "/etc/ssh/sshd_config")
	err := cmd.Run()

	cmd = exec.Command("systemctl", "restart", "sshd")
	// 执行Cmd
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error restarting SSH service: ", err)
	} else {
		fmt.Println("SSH service restarted successfully.")
	}
}
