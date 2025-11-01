package main

import (
	"bufio"
	"fmt"
	"os"
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
					fmt.Printf("%s not found\n", args[0])
				}
			}
		default:
			fmt.Printf("%s: command not found\n", line)
		}
	}
}
