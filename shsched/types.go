package shsched

import "os/exec"

type Recipe struct {
	ExecFile string `json:"execFile"`
	Script   string `json:"script"`
}

type Task struct {
	CMD exec.Cmd
	Dir string
}
