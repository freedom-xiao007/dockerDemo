package subsystem

type CpuSetSubsystem struct {
}

func (c CpuSetSubsystem) Name() string {
	return "cpuset"
}

func (c CpuSetSubsystem) Set(cgroupName string, res *ResourceConfig) error {
	return nil
}

func (c CpuSetSubsystem) Apply(cgroupName string, pid int) error {
	return nil
}

func (c CpuSetSubsystem) Remove(cgroupName string) error {
	return nil
}
