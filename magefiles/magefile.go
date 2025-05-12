package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/magefile/mage/mg"
)

const (
	TargetDir = "target"
)

func ensureDirectoryExists(name string) error {
	switch i, err := os.Stat(name); {
	case err == nil && !i.IsDir():
		return fmt.Errorf(`"%v" exists and is not a directory`, name)
	case os.IsNotExist(err):
		return os.Mkdir(name, 0o755)
	case err != nil:
		return err
	}
	return nil
}

type Build mg.Namespace

func (Build) CLI() error {
	if err := ensureDirectoryExists(TargetDir); err != nil {
		return err
	}
	cmd := exec.Command("go", "build", "-o", fmt.Sprintf("./%s/autoklept", TargetDir), "./cmd/cli")
	fmt.Println(cmd.String())
	return cmd.Run()
}

func (Build) Batch() error {
	if err := ensureDirectoryExists(TargetDir); err != nil {
		return err
	}
	cmd := exec.Command("go", "build", "-o", fmt.Sprintf("./%s/autoklept-batch", TargetDir), "./cmd/batch")
	fmt.Println(cmd.String())
	return cmd.Run()
}

func (b Build) All() error {
	if err := b.CLI(); err != nil {
		return fmt.Errorf("error build CLI: %v", err)
	}
	if err := b.Batch(); err != nil {
		return fmt.Errorf("error build batcher: %v", err)
	}
	return nil
}

func Clean() error {
	fmt.Printf("removing path %v/\n", TargetDir)
	return os.RemoveAll(TargetDir)
}
