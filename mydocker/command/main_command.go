package command

import (
	"dockerDemo/mydocker/cgroup/subsystem"
	"dockerDemo/mydocker/container"
	"dockerDemo/mydocker/run"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
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
		return container.RunContainerInitProcess()
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
		cli.StringFlag{
			Name:  "mem",
			Usage: "memory limit",
		},
		cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit",
		},
		cli.StringFlag{
			Name:  "cpushare",
			Usage: "cpushare limit",
		},
		// 添加-v标签
		cli.StringFlag{
			Name:  "v",
			Usage: "volume",
		},
		// 添加-d标签
		cli.BoolFlag{
			Name:  "d",
			Usage: "detach container",
		},
		// 提供run后面的-name指定容器名字参数
		cli.StringFlag{
			Name:  "name",
			Usage: "container name",
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
		detach := context.Bool("d")
		resConfig := &subsystem.ResourceConfig{
			MemoryLimit: context.String("mem"),
			CpuShare:    context.String("cpuShare"),
			CpuSet:      context.String("cpuSet"),
		}
		volume := context.String("v")
		// 将取到的容器名称传递下去，如果没有则为空
		containerName := context.String("name")
		run.Run(tty, detach, cmdArray, resConfig, volume, containerName)
		return nil
	},
}

var CommitCommand = cli.Command{
	Name:  "commit",
	Usage: "commit a container into image",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("Missing container name")
		}
		imageName := context.Args().Get(0)
		return run.CommitContainer(imageName)
	},
}

var ListCommand = cli.Command{
	Name:  "ps",
	Usage: "list all the container",
	Action: func(context *cli.Context) error {
		return run.ListContainers()
	},
}
