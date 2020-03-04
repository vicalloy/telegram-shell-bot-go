package main

import (
	"bufio"
	"fmt"
	tb "gopkg.in/tucnak/telebot.v2"
	"os"
	"os/exec"
	"strings"
	"time"
)

func handleHelp(m *tb.Message) {
	message := `
Any input text will call as shell commend.
Support command:
/tasks show all running tasks
`
	_, _ = Bot.Send(m.Chat, message)
}

func handleTasks(m *tb.Message) {
	tasksText := []string{}

	for i := range Tasks {
		tasksText = append(tasksText, Tasks[i].String())
	}
	msg := strings.Join(tasksText, "\r\n")
	if len(msg) == 0 {
		msg = "Task list is empty"
	}
	_, _ = Bot.Send(m.Chat, msg)
}

func doCd(m *tb.Message) bool {
	cmd := m.Text
	if !strings.HasPrefix(strings.ToLower(cmd), "cd ") {
		return false
	}
	err := os.Chdir(cmd[3:])
	if err != nil {
		_, _ = Bot.Send(m.Chat, err.Error())
		return false
	}
	pwd, _ := os.Getwd()
	fmt.Println(pwd)
	msg := fmt.Sprintf("pwd: %s", pwd)
	_, _ = Bot.Send(m.Chat, msg)
	return true
}

func handleExecCommand(m *tb.Message) {
	// todo check cmd
	if doCd(m) {
		return
	}
	commandText := m.Text
	doExecCommand(commandText, m)
}

func replyCmdOut(out string, m *tb.Message) {
	if len(out) == 0 {
		return
	}
	if len(out) > 500 {
		out = out[:500]
	}
	_, _ = Bot.Send(m.Chat, out)
}

func doExecCommand(commandText string, m *tb.Message) {
	cmd := exec.Command("/bin/bash", "-c", commandText)

	stdout, err := cmd.StdoutPipe()

	if err != nil {
		fmt.Println(err)
		return
	}

	err = cmd.Start()
	if err != nil {
		fmt.Println(err)
		return
	}

	task := Task{cmd.Process.Pid, commandText, cmd}
	Tasks = append(Tasks, task)

	reader := bufio.NewReader(stdout)
	out := ""
	idx := 0

	startTime := time.Now()

	leapTime := time.Duration(1) * time.Second
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil { // io.EOF
			break
		}
		out += line
		if time.Since(startTime) > leapTime {
			replyCmdOut(out, m)
			idx += 1
			out = ""
			startTime = time.Now()
		}
		if idx > 3 {
			msg := fmt.Sprintf("Command not finished, you can kill by send kill %d", task.Pid)
			_, _ = Bot.Send(m.Chat, msg)
			break
		}
	}

	cmd.Wait()

	// Tasks.remove(task)
	for i := 0; i < len(Tasks); i++ {
		if Tasks[i] == task {
			Tasks = append(Tasks[:i], Tasks[i+1:]...)
			i-- // maintain the correct index
		}
	}

	replyCmdOut(out, m)

	if idx > 3 {
		msg := fmt.Sprintf("Task finished: %s", task.CmdText)
		_, _ = Bot.Send(m.Chat, msg)
	}
}
