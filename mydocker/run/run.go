package run

import (
	"dockerDemo/mydocker/cgroup"
	"dockerDemo/mydocker/cgroup/subsystem"
	"dockerDemo/mydocker/container"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
)

// Run
/*
这里的Start方法是真正开始前面创建好的 command 的调用，
它首先会clone出来一个namespace隔离的进程，然后在子进程中，调用/proc/self/exe,也就是自己调用自己
发送 init 参数，调用我们写的 init 方法，去初始化容器的一些资源
*/
func Run(tty bool, cmdArray []string, config *subsystem.ResourceConfig) {
	pwd, err := os.Getwd()
	if err != nil {
		log.Errorf("Run get pwd err: %v", err)
		return
	}
	mntUrl := pwd + "/mnt/"
	rootUrl := pwd + "/"
	parent, writePipe := container.NewParentProcess(tty, rootUrl, mntUrl)
	if err := parent.Start(); err != nil {
		log.Error(err)
		return
	}

	cgroupManager := cgroup.NewCgroupManager("mydocker-cgroup")
	defer cgroupManager.Destroy()
	if err := cgroupManager.Apply(parent.Process.Pid); err != nil {
		log.Errorf("cgroup apply err: %v", err)
		return
	}
	if err := cgroupManager.Set(config); err != nil {
		log.Errorf("cgoup set err: %v", err)
		return
	}

	sendInitCommand(cmdArray, writePipe)

	log.Infof("parent process run")
	_ = parent.Wait()
	deleteWorkSpace(rootUrl, mntUrl)
	os.Exit(-1)
}

// 将运行参数写入管道
func sendInitCommand(array []string, writePipe *os.File) {
	command := strings.Join(array, " ")
	log.Infof("all command is : %s", command)
	if _, err := writePipe.WriteString(command); err != nil {
		log.Errorf("write pipe write string err: %v", err)
		return
	}
	if err := writePipe.Close(); err != nil {
		log.Errorf("write pipe close err: %v", err)
	}
}

func deleteWorkSpace(rootUrl, mntUrl string) {
	deleteMountPoint(mntUrl)
	deleteWriteLayer(rootUrl)
}

func deleteMountPoint(mntUrl string) {
	cmd := exec.Command("umount", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("deleteMountPoint umount %s err : %v", mntUrl, err)
	}
	if err := os.RemoveAll(mntUrl); err != nil {
		log.Errorf("deleteMountPoint remove %s err : %v", mntUrl, err)
	}
}

func deleteWriteLayer(rootUrl string) {
	writeUrl := rootUrl + "writeLayer/"
	if err := os.RemoveAll(writeUrl); err != nil {
		log.Errorf("deleteMountPoint remove %s err : %v", writeUrl, err)
	}
}
