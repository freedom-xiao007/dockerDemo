package network

import (
	"github.com/stretchr/testify/assert"
	"net"
	"os"
	"testing"
)

func TestIPAM_Allocate(t *testing.T) {
	// 首先把文件删除了，清理重置下环境
	_ = os.RemoveAll(ipamDefaultAllocatorPath)

	// 每次释放和分配ip时，都需要重新调用下面的函数进行IPNet的获取，因为函数调用后，IPNet的值会发生变化
	_, ipNet, _ := net.ParseCIDR("192.168.0.0/24")
	// 第一次分配
	ip1, err := ipAllocator.Allocate(ipNet)
	assert.Equal(t, nil, err)
	assert.Equal(t, "192.168.0.1", ip1.String())

	// 第二个ip分配
	_, ipNet, _ = net.ParseCIDR("192.168.0.0/24")
	ip2, err := ipAllocator.Allocate(ipNet)
	assert.Equal(t, nil, err)
	assert.Equal(t, "192.168.0.2", ip2.String())

	// 释放调第一个IP
	_, ipNet, _ = net.ParseCIDR("192.168.0.1/24")
	assert.Equal(t, nil, ipAllocator.Release(ipNet, &ip1))

	// 能分配得第一个IP
	_, ipNet, _ = net.ParseCIDR("192.168.0.0/24")
	ip3, err := ipAllocator.Allocate(ipNet)
	assert.Equal(t, nil, err)
	assert.Equal(t, "192.168.0.1", ip3.String())

	// 分配第三个IP
	_, ipNet, _ = net.ParseCIDR("192.168.0.0/24")
	ip4, err := ipAllocator.Allocate(ipNet)
	assert.Equal(t, nil, err)
	assert.Equal(t, "192.168.0.3", ip4.String())

	// 释放调第2个IP
	_, ipNet, _ = net.ParseCIDR("192.168.0.2/24")
	assert.Equal(t, nil, ipAllocator.Release(ipNet, &ip2))

	// 第二个ip分配
	_, ipNet, _ = net.ParseCIDR("192.168.0.0/24")
	ip2, err = ipAllocator.Allocate(ipNet)
	assert.Equal(t, nil, err)
	assert.Equal(t, "192.168.0.2", ip2.String())
}
