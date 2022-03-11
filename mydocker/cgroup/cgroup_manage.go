package cgroup

import "dockerDemo/mydocker/cgroup/subsystem"

type CgroupManager struct {
	CgroupName string
	Resouce    *subsystem.ResourceConfig
}

func NewCgroupManager(cgroupName string) *CgroupManager {
	return &CgroupManager{
		CgroupName: cgroupName,
	}
}

// Apply 将PID加入Cgroup
func (c *CgroupManager) Apply(pid int) error {
	for _, ins := range subsystem.SubsystemIns {
		err := ins.Apply(c.CgroupName, pid)
		if err != nil {
			return err
		}
	}
	return nil
}

// Set 设置限制
func (c *CgroupManager) Set(res *subsystem.ResourceConfig) error {
	for _, ins := range subsystem.SubsystemIns {
		err := ins.Set(c.CgroupName, res)
		if err != nil {
			return err
		}
	}
	return nil
}

// Destroy 释放 Cgroup
func (c *CgroupManager) Destroy() error {
	for _, ins := range subsystem.SubsystemIns {
		err := ins.Remove(c.CgroupName)
		if err != nil {
			return err
		}
	}
	return nil
}
