package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"os"
	"os/signal"
	"strconv"
	"time"
)

var (
	version = "snapshot"
)

func main() {
	app := cli.NewApp()
	app.Name = "UUID Generator"
	app.Version = version
	app.Usage = "Generates UUIDs"

	ctx, cancel := handleSignals()
	defer cancel()

	uuid.EnableRandPool()

	defaultCommand := &cli.Command{
		Name:    "infinite",
		Aliases: []string{"i"},
		Usage:   "Generates UUIDs until the app is stopped",
		Action: func(c *cli.Context) error {
			return ticker(ctx, c.Duration("refresh"))
		},
		Flags: []cli.Flag{
			&cli.DurationFlag{
				Name:    "refresh",
				Usage:   "create a uuid every n amount of time",
				Value:   time.Second,
				EnvVars: []string{"REFRESH"},
			},
		},
	}
	generateCommand := &cli.Command{
		Name:    "generate",
		Aliases: []string{"g"},
		Usage:   "Generates one UUID or if specified the amount desired",
		Action: func(c *cli.Context) error {
			countArg := c.Args().Get(0)
			count, err := strconv.Atoi(countArg)
			if err != nil || count == 0 {
				count = 1
			}
			return amount(ctx, count)
		},
	}

	app.DefaultCommand = defaultCommand.Name

	app.Commands = []*cli.Command{
		defaultCommand,
		generateCommand,
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("error on run: %s", err)
	}
}

func amount(ctx context.Context, count int) error {
	i := 0
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if i >= count {
				return nil
			}
			if err := printUUID(); err != nil {
				return err
			}
		}
		i++
	}
}

func ticker(ctx context.Context, refresh time.Duration) error {
	tick := time.NewTicker(refresh)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-tick.C:
			if err := printUUID(); err != nil {
				return err
			}
		}
	}
}

func printUUID() error {
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	fmt.Println(id)
	return nil
}

func handleSignals() (context.Context, context.CancelFunc) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	cancel := func() {
		signal.Stop(signalChan)
		ctxCancel()
	}

	go func() {
		select {
		case <-signalChan: // first signal, cancel context
			fmt.Println("received signal, start shutdown")
			ctxCancel()
		case <-ctx.Done():
		}

		<-signalChan
		fmt.Println("received second signal, hard exit")
		os.Exit(1)
	}()

	return ctx, cancel
}
