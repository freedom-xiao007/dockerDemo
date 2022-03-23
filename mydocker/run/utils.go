package run

import (
	"dockerDemo/mydocker/container"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func getContainerPidByName(containerName string) (string, error) {
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirUrl + container.ConfigName
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return "", fmt.Errorf("read file %s err: %v", configFilePath, err)
	}

	var containerInfo container.ContainerInfo
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return "", fmt.Errorf("json ummarshal err: %v", err)
	}
	return containerInfo.Pid, nil
}

func getContainerInfoByName(containerName string) (*container.ContainerInfo, error) {
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirUrl + container.ConfigName
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("read file %s err: %v", configFilePath, err)
	}

	var containerInfo container.ContainerInfo
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return nil, fmt.Errorf("json ummarshal err: %v", err)
	}
	return &containerInfo, nil
}
