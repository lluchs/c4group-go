package c4group

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
	"regexp"
	"testing"
)

type c4groupProcess struct {
	Cmd            *exec.Cmd
	Stdout, Stderr bytes.Buffer
	Stdin          io.WriteCloser
}

func startC4Group() (*c4groupProcess, error) {
	p := &c4groupProcess{
		Cmd: exec.Command("c4group", "/dev/stdin", "-l"),
	}
	p.Cmd.Stdout = &p.Stdout
	p.Cmd.Stderr = &p.Stderr
	var err error
	p.Stdin, err = p.Cmd.StdinPipe()
	err = p.Cmd.Start()
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

var statusRegexp = regexp.MustCompile(`^Status:\s*`)

func (p *c4groupProcess) VerifyStatus() error {
	err := p.Stdin.Close()
	if err != nil {
		return err
	}
	err = p.Cmd.Wait()
	if err != nil {
		return err
	}
	//println("stdout: ", string(p.Stdout.Bytes()))
	//println("stderr: ", string(p.Stderr.Bytes()))
	status := statusRegexp.ReplaceAllString(string(p.Stderr.Bytes()), "")
	if status != "" {
		return errors.New(status)
	}
	return nil
}

func TestEmpty(t *testing.T) {
	p, err := startC4Group()
	if err != nil {
		t.Fatal(err)
	}
	cw := NewWriter(p.Stdin)
	err = cw.WriteHeader(&Header{
		Entries: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = cw.Close(); err != nil {
		t.Fatal(err)
	}
	if err = p.VerifyStatus(); err != nil {
		t.Error(err)
	}
}

func TestEmptyEntry(t *testing.T) {
	p, err := startC4Group()
	if err != nil {
		t.Fatal(err)
	}
	cw := NewWriter(p.Stdin)
	err = cw.WriteHeader(&Header{
		Entries: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = cw.WriteEntry(&Entry{
		Filename: "foobar.txt",
		Size:     0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = cw.Close(); err != nil {
		t.Fatal(err)
	}
	if err = p.VerifyStatus(); err != nil {
		t.Error(err)
	}
}

func TestSingleFile(t *testing.T) {
	p, err := startC4Group()
	if err != nil {
		t.Fatal(err)
	}
	cw := NewWriter(p.Stdin)
	err = cw.WriteHeader(&Header{
		Entries: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	str := "Hello World!"
	err = cw.WriteEntry(&Entry{
		Filename: "foobar.txt",
		Size:     len(str),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = io.WriteString(cw, str); err != nil {
		t.Fatal(err)
	}
	if err = cw.Close(); err != nil {
		t.Fatal(err)
	}
	if err = p.VerifyStatus(); err != nil {
		t.Error(err)
	}
}
