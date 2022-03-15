package container

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// RunContainerInitProcess
/*
之类的init函数是在容器内部执行的，也就是说，代码执行到这里后，容器所在的进程其实就已经创建出来了，这是本容器执行的第一个进程。
使用mount先去挂载proc文件系统，以便于后面通过ps命等系统命令去查看当前进程资源的情况
*/
func RunContainerInitProcess() error {
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

	cmdArray := readUserCommand()
	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		log.Errorf("can't find exec path: %s %v", cmdArray[0], err)
		return err
	}
	log.Infof("find path: %s", path)
	if err := syscall.Exec(path, cmdArray, os.Environ()); err != nil {
		log.Errorf("syscall exec err: %v", err.Error())
	}
	return nil
}

// 读取程序传入参数
func readUserCommand() []string {
	// 进程默认三个管道，从fork那边传过来的就是第四个（从0开始计数）
	readPipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(readPipe)
	if err != nil {
		log.Errorf("read init argv pipe err: %v", err)
		return nil
	}
	return strings.Split(string(msg), " ")
}
