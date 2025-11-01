package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
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
			os.Exit(1)
		}

		args := strings.Split(line, " ")
		cmd, args := args[0], args[1:]

		switch cmd {
		case "exit":
			if len(args) == 1 {
				code, err := strconv.Atoi(args[0])
				if err != nil {
					os.Exit(1)
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
				case "exit", "echo", "type":
					fmt.Printf("%s is a shell builtin\n", args[0])
				default:
					content := strings.Join(args, " ")
					path, err := exec.LookPath(content)
					if err != nil {
						fmt.Printf("%s: not found\n", content)
						break
					}
					fmt.Printf("%s is %s\n", content, path)
				}
			}
		default:
			fmt.Printf("%s: command not found\n", line)
		}
	}
}
