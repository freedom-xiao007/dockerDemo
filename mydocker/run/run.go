package run

import (
	"dockerDemo/mydocker/container"
	log "github.com/sirupsen/logrus"
	"os"
)

// Run
/*
这里的Start方法是真正开始前面创建好的 command 的调用，
它首先会clone出来一个namespace隔离的进程，然后在子进程中，调用/proc/self/exe,也就是自己调用自己
发送 init 参数，调用我们写的 init 方法，去初始化容器的一些资源
*/
func Run(tty bool, cmdArray []string) {
	parent := container.NewParentProcess(tty, cmdArray)
	if err := parent.Start(); err != nil {
		log.Error(err)
		return
	}
	log.Infof("parent process run")
	_ = parent.Wait()
	os.Exit(-1)
}
