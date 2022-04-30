package run

import (
	"dockerDemo/mydocker/container"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os/exec"
)

func CommitContainer(containerName string) error {
	mntUrl := container.RootUrl + "/mnt/" + containerName
	imageTar := container.RootUrl + "/" + containerName + ".tar"
	log.Infof("commit file path: %s", imageTar)
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntUrl, ".").CombinedOutput(); err != nil {
		return fmt.Errorf("tar folder err: %s, %v", mntUrl, err)
	}
	log.Infof("end commit file: %s", imageTar)
	return nil
}
