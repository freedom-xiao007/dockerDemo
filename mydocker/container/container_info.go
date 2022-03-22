package container

type ContainerInfo struct {
	Pid        string `json:"pid"` // 容器的init进程在宿主机上的PID
	ID         string `json:"id"`
	Name       string `json:"name"`
	Command    string `json:"command"`
	CreateTime string `json:"createTime"`
	Status     string `json:"status"`
}

var (
	RUNNING             = "running"
	STOP                = "stop"
	EXIT                = "exited"
	DefaultInfoLocation = "/var/run/mydocker/%s/"
	ConfigName          = "config.json"
	ContainerLogFile    = "container.log"
)
