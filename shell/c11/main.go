package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Fprint(os.Stdout, "$ ")

		line, err := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if err != nil {
			panic(err)
		}

		args := strings.Split(line, " ")
		cmd, args := args[0], args[1:]
		content := strings.Join(args, " ")

		switch cmd {
		case "exit":
			if len(args) == 1 {
				code, err := strconv.Atoi(args[0])
				if err != nil {
					panic(err)
				}
				os.Exit(code)
			} else {
				fmt.Println("exit: too many arguments")
			}
		case "echo":
			fmt.Printf("%s\n", strings.Join(args, " "))
		case "type":
			for _, arg := range args {
				switch arg {
				case "exit", "echo", "type", "pwd", "cd":
					fmt.Printf("%s is a shell builtin\n", args[0])
				default:
					path, err := exec.LookPath(content)
					if err != nil {
						fmt.Printf("%s: not found\n", content)
						break
					}
					fmt.Printf("%s is %s\n", content, path)
				}
			}
		case "pwd":
			pwd, err := os.Getwd()
			if err != nil {
				panic(err)
			}
			fmt.Printf("%s\n", pwd)
		case "cd":
			p := path.Clean(args[0])
			if !path.IsAbs(p) {
				dir, _ := os.Getwd()
				p = path.Join(dir, p)
			}

			if err := os.Chdir(p); err != nil {
				fmt.Printf("cd: %s: No such file or directory\n", args[0])
			}
		default:
			command := exec.Command(cmd, content)
			command.Stderr = os.Stderr
			command.Stdout = os.Stdout
			err := command.Run()
			if err != nil {
				fmt.Printf("%s: command not found\n", cmd)
			}
		}
	}
}
