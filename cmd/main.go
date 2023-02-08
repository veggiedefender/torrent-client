package main

import (
	"context"
	"fmt"
	"github.com/edelars/console-torrent-client/internal/config"
	"github.com/edelars/console-torrent-client/pkg/console_print"
	"github.com/edelars/console-torrent-client/pkg/pool"
	"github.com/edelars/console-torrent-client/version"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/jessevdk/go-flags"
)

const (
	app_name = "Console-torrent-client"
)

func main() {

	var env config.Environment

	p := flags.NewParser(&env, flags.Default)
	if _, err := p.Parse(); err != nil {
		fmt.Print("%w", err)
	}
	if err := env.Init(); err != nil {
		fmt.Print("%w", err)
	}

	cp := console_print.NewConsolePrint([]string{app_name}, []string{"-----------"})

	var wg sync.WaitGroup
	errs := make(chan error, 4)
	go waitInterruptSignal(errs)

	ctx := context.Background()

	cp.Log("Starting pool")
	mainPool := pool.NewTorrentPool(cp, env.TFDir, env.DFDir)
	wg.Add(1)
	go func() {
		defer wg.Done()
		errs <- mainPool.Start(ctx)
	}()

	cp.Log(fmt.Sprintf("Start release: %s, commit: %s, build time: %s",
		version.Release, version.Commit, version.BuildTime))

	if arg := os.Args[len(os.Args)-2]; arg == "add " {
		file := os.Args[len(os.Args)-1]

		cp.Log("Add file: " + file)

		if err := mainPool.AddFileToPool(file); err != nil {
			cp.Log(err.Error())
		}
	}

	err := <-errs
	if err != nil {
		cp.Log(err.Error())
	}
	cp.Log("trying to shutdown gracefully")

	wg.Add(1)
	go func() {
		defer wg.Done()
		errs <- mainPool.Stop()
	}()
	cp.Log("Pool stopped")
}

func waitInterruptSignal(errs chan<- error) {
	c := make(chan os.Signal, 3)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	errs <- fmt.Errorf("%s", <-c)
	signal.Stop(c)
}
