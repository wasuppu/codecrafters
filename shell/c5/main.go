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

		cmds := strings.Split(line, " ")

		switch cmds[0] {
		case "exit":
			code, err := strconv.Atoi(cmds[1])
			if err != nil {
				os.Exit(1)
			}
			os.Exit(code)
		case "echo":
			fmt.Fprintf(os.Stdout, "%s\n", strings.Join(cmds[1:], " "))
		default:
			fmt.Fprintf(os.Stdout, "%s: command not found\n", line)
		}
	}
}
