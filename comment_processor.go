package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"time"
	"syscall"
	"io/ioutil"
)

func createKeyValuePairs(m map[string]string) string {
	b := new(bytes.Buffer)
	for key, value := range m {
		_, err := fmt.Fprintf(b, " %s=\"%s\"", key, value)
		if err != nil {
			panic(err)
		}
	}
	return b.String()
}

func CreateDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}

func DoProcess(comment *Comment) (error, *CbsdTask) {
	dt := time.Now()

	CreateDirIfNotExist("log")

	filePath := fmt.Sprintf("log/%s_%s_%d.txt", dt.Format(time.RFC3339), comment.Command, comment.JobID)
	commentFile, err := os.Create(filePath)
	if err != nil {
		return err, nil
	}

	defer func() {
		_ = commentFile.Close()
	}()

	fmt.Printf("JobID %d\n", comment.JobID)

	cbsdArgs := createKeyValuePairs(comment.CommandArgs)

	cmdstr := fmt.Sprintf("/usr/local/bin/cbsd %s %s", comment.Command, cbsdArgs)
	_, err = fmt.Fprintf(commentFile, "%s\n", cmdstr)
	if err != nil {
		return err, nil
	}

	cmd := exec.Command("/bin/sh", filePath)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	filePath = fmt.Sprintf("log/%d.txt", comment.JobID)

	stdoutFile, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}

	defer func() {
		_ = stdoutFile.Close()
	}()

	cmd.Stdout = stdoutFile

	if err := cmd.Start(); err != nil {
		fmt.Printf("\ncmd.Start: %v\n")
	}

	if err != nil {
		fmt.Printf("\n%v\n", err)
	}

	cmdStatus := 0

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				cmdStatus = status.ExitStatus()
			}
		} else {
			fmt.Printf("\ncmd.Wait error: %v\n", err)
			cmdStatus = 0
		}
	} else {
		cmdStatus = 0
	}

	cbsdTask := CbsdTask{}

	cbsdTask.ErrCode = cmdStatus
	// progress always 100 for completed/failed command
	cbsdTask.Progress = 100

	fileLogPath := fmt.Sprintf("log/%d.txt", comment.JobID)

	b, err := ioutil.ReadFile(fileLogPath) // just pass the file name
	if err != nil {
		fmt.Print(err)
	}

	cbsdTask.Message = string(b) // convert content to a 'string'

	return nil, &cbsdTask
}
