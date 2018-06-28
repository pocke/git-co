package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
)

func main() {
	if err := Main(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}

type History map[string][]string

func Main(args []string) error {
	if len(args) >= 2 && args[1] == "--list" {
		return List()
	}
	return CheckoutAndRecord(args)
}

func List() error {
	h, err := loadHistoryFile()
	if err != nil {
		return err
	}
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	branches := h[pwd]
	if branches == nil {
		branches = []string{}
	}
	for i := 1; i <= len(branches); i++ {
		fmt.Println(branches[len(branches)-i])
	}
	return nil
}

func CheckoutAndRecord(args []string) error {
	a := []string{"checkout"}
	a = append(a, args[1:]...)
	cmd := exec.Command("git", a...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	return RecordCommits(args[1:])
}

func RecordCommits(args []string) error {
	for _, v := range args {
		if strings.HasPrefix(v, "-") ||
			strings.HasPrefix(v, "@") ||
			FileExists(v) ||
			!CommitExist(v) {
			continue
		}

		err := RecordCommit(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func RecordCommit(ref string) error {
	fmt.Printf("recording %s\n", ref)
	history, err := loadHistoryFile()
	if err != nil {
		return err
	}
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if history[pwd] == nil {
		history[pwd] = []string{}
	}
	history[pwd] = AppendUniq(history[pwd], ref)
	return saveHistoryFile(history)
}

func CommitExist(ref string) bool {
	cmd := exec.Command("git", "rev-parse", ref)
	return cmd.Run() == nil
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func loadHistoryFile() (History, error) {
	path, err := historyFilePath()
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if len(content) == 0 {
		content = []byte("{}")
	}

	var res History
	err = json.Unmarshal(content, &res)
	return res, err
}

func saveHistoryFile(h History) error {
	path, err := historyFilePath()
	if err != nil {
		return err
	}

	b, err := json.Marshal(h)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, b, 0644)
}

func historyFilePath() (string, error) {
	return homedir.Expand("~/.cache/git-co-history.json")
}

func AppendUniq(slice []string, s string) []string {
	slice = append(slice, s)
	for idx, v := range slice {
		if idx == len(slice)-1 {
			break
		}
		if v == s {
			slice = append(slice[:idx], slice[idx+1:]...)
			break
		}
	}
	return slice
}
