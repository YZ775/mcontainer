/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"

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

func runMain(cmd *cobra.Command, args []string) error {
	if initFlag == "true" {
		init_container()
		return nil
	}
	fmt.Printf("main process pid %d\n", os.Getpid())
	forkCmd := exec.Command("/proc/self/exe", append(os.Args[1:], "--init=true")...)
	forkCmd.Stdin = os.Stdin
	forkCmd.Stdout = os.Stdout
	forkCmd.Stderr = os.Stderr
	// forkCmd.SysProcAttr = &syscall.SysProcAttr{
	// 	Cloneflags: syscall.CLONE_NEWIPC | syscall.CLONE_NEWNET | syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	// }
	err := forkCmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func init_container() {
	fmt.Printf("container process pid %d\n", os.Getpid())
}
