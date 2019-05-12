package shsched

import "os/exec"

type Recipe struct {
	ExecFile   string `json:"execFile"`
	Script     string `json:"script"`
	RetAddress string `json:"retAddress"`
}

type Task struct {
	CMD        exec.Cmd
	Dir        string
	RetAddress string
}

type Output struct {
	RetAddress string
	Output     string
}

type Input struct {
	RetAddress     string
	RecipeFilePath string
}
