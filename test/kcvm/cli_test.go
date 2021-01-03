package test

import (
	"io/ioutil"
	"testing"

	// "fmt"
	cmds "github.com/raqqun/kcvm/pkg/kcvm"
)

var kcvm = cmds.InitCLI()

func TestInitCmd(t *testing.T) {
	kcvm.Writer = ioutil.Discard

	err := kcvm.Run([]string{"kcvm", "--dir", "/tmp/kcvm", "init"})
	if err != nil {
		t.Error(err)
	}
}

func TestInstallCmd(t *testing.T) {
	kcvm.Writer = ioutil.Discard

	err := kcvm.Run([]string{"kcvm", "--dir", "/tmp/kcvm", "install", "v1.20.0"})
	if err != nil {
		t.Error(err)
	}
}

func TestUseCmd(t *testing.T) {
	kcvm.Writer = ioutil.Discard

	err := kcvm.Run([]string{"kcvm", "--dir", "/tmp/kcvm", "use", "v1.20.0"})
	if err != nil {
		t.Error(err)
	}
}

func TestListRemoteCmd(t *testing.T) {
	kcvm.Writer = ioutil.Discard

	err := kcvm.Run([]string{"kcvm", "--dir", "/tmp/kcvm", "list", "remote"})
	if err != nil {
		t.Error(err)
	}
}

func TestListLocalCmd(t *testing.T) {
	kcvm.Writer = ioutil.Discard

	err := kcvm.Run([]string{"kcvm", "--dir", "/tmp/kcvm", "list", "local"})
	if err != nil {
		t.Error(err)
	}
}
