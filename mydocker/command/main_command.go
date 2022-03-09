package command

import (
	"dockerDemo/mydocker/container"
	"dockerDemo/mydocker/run"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"strings"
)

var InitCommand = cli.Command{
	Name:  "init",
	Usage: "Init container process run user's process in container. Do not call it outside",
	/*
		1.获取传递过来的command参数
		2.执行容器初始化操作
	*/
	Action: func(context *cli.Context) error {
		log.Infof("init come on")
		cmd := context.Args().Get(0)
		args := strings.Split(context.Args().Get(1), " ")
		log.Infof("command: %s, args: %s", cmd, args)
		return container.RunContainerInitProcess(cmd, args)
	},
}

var RunCommand = cli.Command{
	Name:  "run",
	Usage: `Create a container with namespace and cgroups limit mydocker run -ti [command]`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
		},
	},
	/*
		这里是run命令执行的真正函数
		1.判断参数是否包含command
		2.获取用户指定的command
		3.调用Run function 去准备启动容器
	*/
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container command")
		}
		var cmdArray []string
		for _, arg := range context.Args() {
			cmdArray = append(cmdArray, arg)
		}
		tty := context.Bool("ti")
		run.Run(tty, cmdArray)
		return nil
	},
}
