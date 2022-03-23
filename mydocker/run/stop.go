package run

import (
	"dockerDemo/mydocker/container"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"syscall"
)

func StopContainer(containerName string) error {
	pid, err := getContainerPidByName(containerName)
	if err != nil {
		return err
	}

	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		return fmt.Errorf("convert pid %s to int err: %v", pid, err)
	}

	if err := syscall.Kill(pidInt, syscall.SIGTERM); err != nil {
		return fmt.Errorf("send sigterm %d, err: %v", pid, err)
	}

	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		return fmt.Errorf("get container info err: %v", err)
	}

	containerInfo.Status = container.STOP
	containerInfo.Pid = ""
	newContainerInfo, err := json.Marshal(containerInfo)
	if err != nil {
		return fmt.Errorf("json marshal %v,err: %v", containerInfo, err)
	}

	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configPath := dirUrl + container.ConfigName
	if err := ioutil.WriteFile(configPath, newContainerInfo, 0622); err != nil {
		return fmt.Errorf("write file %s err: %v", configPath, err)
	}
	return nil
}
