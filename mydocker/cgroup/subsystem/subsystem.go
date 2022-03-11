package subsystem

type ResourceConfig struct {
	MemoryLimit string
	CpuShare    string
	CpuSet      string
}

type Subsystem interface {
	Name() string
	Set(cgroupName string, res *ResourceConfig) error
	Apply(cgroupName string, pid int) error
	Remove(cgroupName string) error
}

var SubsystemIns = []Subsystem{
	&CpuSetSubsystem{},
	&CpuShareSubsystem{},
	&MemorySubsystem{},
}
