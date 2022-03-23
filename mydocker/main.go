package main

import (
	"dockerDemo/mydocker/command"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
)

const usage = `mydocker is a simple container runtime implementation.
               The purpose of this projects is to learn how docker works and how to write a docker by ourselves
               Enjoy it, just for fun.`

func main() {
	app := cli.NewApp()
	app.Name = "mydocker"
	app.Usage = usage

	app.Commands = []cli.Command{
		command.InitCommand,
		command.RunCommand,
		command.CommitCommand,
		command.ListCommand,
		command.LogCommand,
		command.ExecCommand,
		command.StopCommand,
		command.RemoveCommand,
	}

	app.Before = func(context *cli.Context) error {
		log.SetFormatter(&log.JSONFormatter{})
		log.SetOutput(os.Stdout)
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
