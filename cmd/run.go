/*
Copyright © 2023 Yuzuki Mimura
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/YZ775/mcontainer/pkg/device"
	"github.com/YZ775/mcontainer/pkg/status"
	"github.com/opencontainers/runtime-spec/specs-go"
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
	rootCmd.Flags().MarkHidden("init")
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
		err := initContainer(cmd, args)
		if err != nil {
			return err
		}
		return nil
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
	err := forkCmd.Start()
	if err != nil {
		return err
	}

	statusFile := "/var/run/mcontainer/" + strconv.Itoa(forkCmd.Process.Pid)
	err = os.MkdirAll("/var/run/mcontainer/", 0755)
	if err != nil {
		return err
	}
	f, err := os.Create(statusFile)
	defer f.Close()
	if err != nil {
		return err
	}

	spec, err := readSpec(filepath.Clean(args[0]))
	status := &status.Status{}
	status.OciVersion = spec.Version
	status.Pid = forkCmd.Process.Pid
	status.Status = "created"
	status.Bundle = filepath.Clean(args[0])
	status.ID = strconv.Itoa(rand.Int())
	fmt.Printf("%v\n", status)
	_, err = f.Write([]byte(fmt.Sprintf("%v\n", status)))
	if err != nil {
		return err
	}

	forkCmd.Wait()

	status.Status = "stopped"
	f, err = os.Create(statusFile)
	defer f.Close()
	_, err = f.Write([]byte(fmt.Sprintf("%v\n", status)))
	if err != nil {
		return err
	}

	return nil
}

func initContainer(cmd *cobra.Command, args []string) error {
	fmt.Printf("container process pid %d\n", os.Getpid())
	bundlePath := filepath.Clean(args[0])
	spec, err := readSpec(bundlePath)
	if err != nil {
		return err
	}
	root := filepath.Join(bundlePath, spec.Root.Path)

	err = mount(spec, root)
	if err != nil {
		return err
	}
	err = defaultDevMount(root)

	if err != nil {
		return err
	}
	err = syscall.Chroot(root)
	if err != nil {
		return err
	}
	err = os.Chdir("/")
	if err != nil {
		return err
	}

	newcmd := exec.Command("/bin/bash")
	newcmd.Stdin = os.Stdin
	newcmd.Stdout = os.Stdout
	newcmd.Stderr = os.Stderr
	err = newcmd.Run()
	if err != nil {
		return err
	}
	defer func() {
		err = unmountAll(spec)
		if err != nil {
			panic(err)
		}
	}()
	return nil
}

func readSpec(path string) (*specs.Spec, error) {
	jsonFile, err := ioutil.ReadFile(filepath.Join(path, "config.json"))
	if err != nil {
		return nil, err
	}
	spec := &specs.Spec{}
	err = json.Unmarshal(jsonFile, &spec)
	if err != nil {
		return nil, err
	}
	return spec, nil
}

func defaultDevMount(roofsPath string) error {
	for _, d := range device.DefaultDevices {
		devicePathInRootfs := filepath.Join(roofsPath, d)
		f, err := os.Create(devicePathInRootfs)
		defer f.Close()
		err = syscall.Mount(d, devicePathInRootfs, "bind", syscall.MS_BIND, "")
		if err != nil {
			return err
		}
	}
	return nil
}

func mount(spec *specs.Spec, roofsPath string) error {
	for _, m := range spec.Mounts {
		destinationPathInRootfs := filepath.Join(roofsPath, m.Destination)
		err := os.MkdirAll(destinationPathInRootfs, 0775)
		if err != nil {
			return err
		}
		err = syscall.Mount(m.Source, destinationPathInRootfs, m.Type, 0, "")
		if err != nil {
			fmt.Println(err)
			continue
			// return err
		}
	}
	return nil
}

func unmountAll(spec *specs.Spec) error {
	for _, d := range device.DefaultDevices {
		err := syscall.Unmount(d, 0)
		if err != nil {
			return err
		}
	}
	for i := len(spec.Mounts) - 1; i >= 0; i-- {
		err := syscall.Unmount(spec.Mounts[i].Destination, 0)
		if err != nil {
			fmt.Println(err)
			continue
			// return err
		}
	}
	return nil
}
