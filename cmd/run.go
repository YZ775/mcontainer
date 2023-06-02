/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run container",
	Long:  "run container",
	RunE:  runMain,
}

var initFlag string

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVar(&initFlag, "init", "", "fork new process")
	// rootCmd.Flags().MarkHidden("init")
}

var namespaceTable = map[string]uintptr{
	"pid":     syscall.CLONE_NEWPID,
	"network": syscall.CLONE_NEWNET,
	"ipc":     syscall.CLONE_NEWIPC,
	"uts":     syscall.CLONE_NEWUTS,
	"mount":   syscall.CLONE_NEWNS,
	"cgroup":  syscall.CLONE_NEWCGROUP,
}

func runMain(cmd *cobra.Command, args []string) error {
	if initFlag == "true" {
		err := init_container()
		return err
	}
	fmt.Printf("main process pid %d\n", os.Getpid())
	forkCmd := exec.Command("/proc/self/exe", append(os.Args[1:], "--init=true")...)
	forkCmd.Stdin = os.Stdin
	forkCmd.Stdout = os.Stdout
	forkCmd.Stderr = os.Stderr
	ns := []string{"pid", "network", "ipc", "uts", "mount", "cgroup"}

	var nsFlag uintptr
	for _, n := range ns {
		val, ok := namespaceTable[n]
		if ok {
			nsFlag |= val
		}
	}

	forkCmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: nsFlag,
	}
	err := forkCmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func init_container() error {
	fmt.Printf("container process pid %d\n", os.Getpid())
	newcmd := exec.Command("/bin/bash")
	newcmd.Stdin = os.Stdin
	newcmd.Stdout = os.Stdout
	newcmd.Stderr = os.Stderr
	err := newcmd.Run()
	if err != nil {
		return err
	}
	return nil
}
