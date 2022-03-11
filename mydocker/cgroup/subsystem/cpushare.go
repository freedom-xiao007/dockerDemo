package subsystem

type CpuShareSubsystem struct {
}

func (c CpuShareSubsystem) Name() string {
	return "cpu"
}

func (c CpuShareSubsystem) Set(cgroupName string, res *ResourceConfig) error {
	return nil
}

func (c CpuShareSubsystem) Apply(cgroupName string, pid int) error {
	return nil
}

func (c CpuShareSubsystem) Remove(cgroupName string) error {
	return nil
}
