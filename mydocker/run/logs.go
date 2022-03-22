package run

import (
	"dockerDemo/mydocker/container"
	"fmt"
	"io/ioutil"
	"os"
)

func LogContainer(containerName string) error {
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	logFilePath := dirUrl + container.ContainerLogFile
	file, err := os.Open(logFilePath)
	defer file.Close()
	if err != nil {
		return fmt.Errorf("open file %s, err: %v", logFilePath, err)
	}

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("read file %s, err: %v", logFilePath, err)
	}
	fmt.Fprint(os.Stdout, string(content))
	return nil
}
