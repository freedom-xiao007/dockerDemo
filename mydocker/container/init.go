package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// RunContainerInitProcess
/*
之类的init函数是在容器内部执行的，也就是说，代码执行到这里后，容器所在的进程其实就已经创建出来了，这是本容器执行的第一个进程。
使用mount先去挂载proc文件系统，以便于后面通过ps命等系统命令去查看当前进程资源的情况
*/
func RunContainerInitProcess() error {
	if err := setUpMount(); err != nil {
		return err
	}

	cmdArray := readUserCommand()
	log.Infof("cmd Array []: %s", cmdArray[0])
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

// 初始化挂载点
func setUpMount() error {
	// 首先设置根目录为私有模式，防止影响pivot_root
	if err := syscall.Mount("/", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, ""); err != nil {
		return fmt.Errorf("setUpMount Mount proc err: %v", err)
	}

	// 进入Busybox,固定路径，busybox提前解压好，放到指定的配置路径
	//err := privotRoot(RootUrl + "/mnt/busybox")
	err := privotRoot(BusyboxPath)
	if err != nil {
		return err
	}

	// mount proc
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	err = syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	if err != nil {
		log.Errorf("proc挂载 failed: %v", err)
		return err
	}
	syscall.Mount("tmpfs", "/dev", "tempfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
	return nil
}

func privotRoot(root string) error {
	pwd, err := os.Getwd()
	if err != nil {
		log.Errorf("pwd err: %v", err)
		return err
	}
	log.Infof("current pwd: %s", pwd)

	// 为了使当前root的老root和新root不在同一个文件系统下，我们把root重新mount一次
	// bind mount 是把相同的内容换了一个挂载点的挂载方法
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("mount rootfs to itself error: %v", err)
	}

	// 创建 rootfs、.pivot_root 存储 old_root
	pivotDir := filepath.Join(root, ".pivot_root")
	// 判断当前目录是否已有该文件夹
	if _, err := os.Stat(pivotDir); err == nil {
		// 存在则删除
		if err := os.Remove(pivotDir); err != nil {
			return err
		}
	}
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return fmt.Errorf("mkdir of pivot_root err: %v", err)
	}

	// pivot_root 到新的rootfs，老的old_root现在挂载在rootfs/.pivot_root上
	// 挂载点目前依然可以在mount命令中看到
	log.Infof("root: %s， pivotDir: %s", root, pivotDir)
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot_root err: %v", err)
	}

	// 修改当前工作目录到跟目录
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir root err: %v", err)
	}

	// 取消临时文件.pivot_root的挂载并删除它
	// 注意当前已经在根目录下，所以临时文件的目录也改变了
	pivotDir = filepath.Join("/", ".pivot_root")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root dir err: %v", err)
	}
	return os.Remove(pivotDir)
}
