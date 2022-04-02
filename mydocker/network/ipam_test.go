package network

import (
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestIPAM_Allocate(t *testing.T) {
	_, ipNet, _ := net.ParseCIDR("192.168.0.0/24")
	ip1, err := ipAllocator.Allocate(ipNet)
	assert.Equal(t, nil, err)
	assert.Equal(t, "192.168.0.1", ip1.String())

	ip2, err := ipAllocator.Allocate(ipNet)
	assert.Equal(t, nil, err)
	assert.Equal(t, "192.168.0.2", ip2.String())

	assert.Equal(t, nil, ipAllocator.Release(ipNet, &ip1))
	ip3, err := ipAllocator.Allocate(ipNet)
	assert.Equal(t, nil, err)
	assert.Equal(t, "192.168.0.1", ip3.String())
}
