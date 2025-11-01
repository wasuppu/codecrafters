package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

const usage = `mydocker is a simple container runtime implementation.
Usage:
  mydocker run -ti [command]     # Run container with tty
  mydocker init [command]       # Init container process`

func main() {
	if len(os.Args) < 2 {
		fmt.Println(usage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		cmd := flag.NewFlagSet("run", flag.ExitOnError)
		tty := cmd.Bool("ti", false, "enable tty")
		cmd.Parse(os.Args[2:])

		if len(cmd.Args()) == 0 {
			fmt.Println("missing container command")
			os.Exit(1)
		}
		runContainer(*tty, cmd.Args()[0])

	case "init":
		if len(os.Args) < 3 {
			fmt.Println("missing command")
			os.Exit(1)
		}
		initContainer(os.Args[2])

	default:
		fmt.Println("Unknown command")
		os.Exit(1)
	}
}

func runContainer(tty bool, command string) {
	parent := exec.Command("/proc/self/exe", "init", command)
	parent.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWPID | syscall.CLONE_NEWNET,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	if tty {
		parent.Stdin = os.Stdin
		parent.Stdout = os.Stdout
		parent.Stderr = os.Stderr
	}

	if err := parent.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func initContainer(command string) {
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), ""); err != nil {
		fmt.Printf("Failed to mount proc: %v\n", err)
		os.Exit(1)
	}

	argv := []string{command}
	if err := syscall.Exec(command, argv, os.Environ()); err != nil {
		fmt.Printf("Failed to execute command: %v\n", err)
		os.Exit(1)
	}
}
