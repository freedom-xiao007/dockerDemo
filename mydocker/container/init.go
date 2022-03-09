package container

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"syscall"
)

// RunContainerInitProcess
/*
之类的init函数是在容器内部执行的，也就是说，代码执行到这里后，容器所在的进程其实就已经创建出来了，这是本容器执行的第一个进程。
使用mount先去挂载proc文件系统，以便于后面通过ps命等系统命令去查看当前进程资源的情况
*/
func RunContainerInitProcess(command string, args []string) error {
	log.Infof("RunContainerInitProcess command %s, args %s", command, args)

	// private 方式挂载，不影响宿主机的挂载
	err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	if err != nil {
		log.Errorf("private 方式挂载 failed: %v", err)
		return err
	}

	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	err = syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	if err != nil {
		log.Errorf("proc挂载 failed: %v", err)
		return err
	}

	// 试验容器内的第一个进程非我们传入的运行命令时，可放开下面的注释，关闭后面的Exec
	//cmd := exec.Command(command)
	//cmd.Stdin = os.Stdin
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr
	//if err := cmd.Run(); err != nil {
	//	log.Fatal(err)
	//}
	//os.Exit(-1)

	path, err := exec.LookPath(command)
	if err != nil {
		log.Errorf("can't find exec path: %s %v", command, err)
		return err
	}
	log.Infof("find path: %s", path)
	if err := syscall.Exec(path, args, os.Environ()); err != nil {
		log.Errorf("syscall exec err: %v", err.Error())
	}
	return nil
}
