package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"syscall"
)

// NewParentProcess
/*
这里是父进程，也就是当前进程执行的内容
1.这里的 /proc/self/exe 调用中， /proc/self/ 指定是当前运行进程自己的环境，exec是自己调用自己，使用这种方式对创造出来的进程进行初始化
2.后面的args是参数，其中init是传递给本进程的第一个参数，在本例中，其实就是会去调用initCommand去初始化进程的一些环境和资源
3.下面的clone参数就是去fork出来一个新进程，并且使用了namespace隔离新创建的进程和外部环境
4.如果用户指定了-ti参数，就需要把当前进程的输入输出导入到标准输入输出上
*/
func NewParentProcess(tty bool, rootUrl, mntUrl string) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		log.Errorf("create pipe error: %v", err)
		return nil, nil
	}

	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	// 将管道的一端传入fork的进程中
	cmd.ExtraFiles = []*os.File{readPipe}
	if err := newWorkSpace(rootUrl, mntUrl); err != nil {
		log.Errorf("new work space err: %v", err)
		return nil, nil
	}
	cmd.Dir = mntUrl
	return cmd, writePipe
}

func newWorkSpace(rootUrl string, mntUrl string) error {
	if err := createReadOnlyLayer(rootUrl); err != nil {
		return err
	}
	if err := createWriteLayer(rootUrl); err != nil {
		return err
	}
	if err := createMountPoint(rootUrl, mntUrl); err != nil {
		return err
	}
	return nil
}

// 我们直接把busybox放到了工程目录下，直接作为容器的只读层
func createReadOnlyLayer(rootUrl string) error {
	busyboxUrl := rootUrl + "busybox/"
	exist, err := pathExist(busyboxUrl)
	if err != nil {
		return err
	}
	if !exist {
		return fmt.Errorf("busybox dir don't exist: %s", busyboxUrl)
	}
	return nil
}

// 创建一个名为writeLayer的文件夹作为容器的唯一可写层
func createWriteLayer(rootUrl string) error {
	writeUrl := rootUrl + "writeLayer/"
	if err := os.Mkdir(writeUrl, 0777); err != nil {
		return fmt.Errorf("create write layer failed: %v", err)
	}
	return nil
}

func createMountPoint(rootUrl string, mntUrl string) error {
	// 创建mnt文件夹作为挂载点
	if err := os.Mkdir(mntUrl, 0777); err != nil {
		return fmt.Errorf("mkdir faild: %v", err)
	}
	// 把writeLayer和busybox目录mount到mnt目录下
	dirs := "dirs=" + rootUrl + "writeLayer:" + rootUrl + "busybox"
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mmt dir err: %v", err)
	}
	return nil
}

func pathExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, err
	}
	return false, err
}
