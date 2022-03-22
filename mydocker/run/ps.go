package run

import (
	"dockerDemo/mydocker/container"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"text/tabwriter"
)

func ListContainers() error {
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, "")
	dirUrl = dirUrl[:len(dirUrl)-1]
	files, err := ioutil.ReadDir(dirUrl)
	if err != nil {
		return fmt.Errorf("read dir %s err: %v", dirUrl, err)
	}

	var containers []*container.ContainerInfo
	for _, file := range files {
		tmpContainer, err := getContainerInfo(file)
		if err != nil {
			return err
		}
		containers = append(containers, tmpContainer)
	}

	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	_, _ = fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	for _, item := range containers {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", item.ID, item.Name, item.Pid, item.Status, item.Command, item.CreateTime)
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("flush ps write err: %v", err)
	}
	return nil
}

func getContainerInfo(file fs.FileInfo) (*container.ContainerInfo, error) {
	containerName := file.Name()
	configFileDir := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := configFileDir + container.ConfigName
	content, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("read file %s err: %v", configFilePath, err)
	}
	var containerInfo container.ContainerInfo
	if err := json.Unmarshal(content, &containerInfo); err != nil {
		return nil, fmt.Errorf("json unmarshal err: %v", err)
	}
	return &containerInfo, nil
}
