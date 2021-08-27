package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/guptarohit/asciigraph"

	pkgErr "github.com/pkg/errors"
)

func main() {
	err := realMain()
	if err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}
}

func realMain() error {
	if len(os.Args) < 2 {
		return errors.New("command must be given")
	}
	sh := os.Getenv("SHELL")
	if sh == "" {
		sh = "sh"
	}
	cmdStr := strings.Join(os.Args[1:], " ")
	ctx := context.Background()
	executor := cmdExecutor{cmd: cmdStr, sh: sh}
	render := graphRenderer{max: 100}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	for {
		select {
		case t := <-ticker.C:
			p, err := executor.Run(ctx)
			if err != nil {
				return err
			}
			asciigraph.Clear()
			fmt.Println(t)
			fmt.Println(cmdStr)
			fmt.Print(render.Next(p))
		case s := <-sig:
			fmt.Printf("stopped by %s\n", s)
			return nil
		}
	}
}

type cmdExecutor struct {
	cmd string
	sh  string
}

func (e *cmdExecutor) Run(ctx context.Context) (float64, error) {
	cmd := exec.CommandContext(ctx, e.sh, "-c", e.cmd)
	out, err := cmd.Output()
	if err != nil {
		return 0, pkgErr.Wrapf(err, "failed to run: %s", e.cmd)
	}
	s := strings.TrimSpace(string(out))
	p, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, pkgErr.Wrapf(err, "failed to parse: %s", s)
	}
	return p, nil
}

type graphRenderer struct {
	data []float64
	max  int
}

func (r *graphRenderer) Next(d float64) string {
	if r.max > 0 && len(r.data) == r.max {
		r.data = append(r.data[1:], d)
	} else {
		r.data = append(r.data, d)
	}
	return asciigraph.Plot(r.data, asciigraph.Height(10))
}
