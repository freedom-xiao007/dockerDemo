package network

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"net"
	"os/exec"
	"strings"
	"time"
)

type BridgeNetworkDriver struct {
}

func (b *BridgeNetworkDriver) Name() string {
	return "bridge"
}

func (b *BridgeNetworkDriver) Create(subnet string, name string) (*NetWork, error) {
	_, ipRange, _ := net.ParseCIDR(subnet)
	ip, err := ipAllocator.Allocate(ipRange)
	if err != nil {
		return nil, err
	}
	ipRange.IP = ip
	n := &NetWork{
		Name:      name,
		IpRange:   ipRange,
		Driver:    b.Name(),
		GatewayIP: ip,
		Subnet:    subnet,
	}
	log.Infof("BridgeNetworkDriver creat network subnet: %s, gateway ip: %s", ipRange.String(), ip.String())
	return n, b.initBridge(n)
}

func (b *BridgeNetworkDriver) Delete(network NetWork) error {
	bridgeName := network.Name
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}
	return netlink.LinkDel(br)
}

// Connect 连接一个网络和网络端点
func (b *BridgeNetworkDriver) Connect(network *NetWork, endpoint *Endpoint) error {
	// 获取网络名，linux bridge 的接口名
	bridgeName := network.Name
	// 通过接口名获取到linux bridge接口的对象和接口属性
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}

	// 创建veth接口的配置
	la := netlink.NewLinkAttrs()
	// 由于linux接口名的限制，名字取endpoint id的前5位
	la.Name = endpoint.ID[:5]
	// 通过设置veth接口的master属性，设置这个veth的一端挂载到网络对应的linux bridge上
	la.MasterIndex = br.Attrs().Index

	// 创建Veth对象，通过peerName配置veth另外一端的接口名
	// 配置veth另外一端的名字cif-{endpoint id的前5位}
	endpoint.Device = netlink.Veth{
		LinkAttrs: la,
		PeerName:  "cif-" + endpoint.ID[:5],
	}

	// 调用netlink的linkadd方法创建出这个veth接口
	// 因为上面指定了link的masterIndex是网络对应的linux bridge
	// 所以Veth的一端就已经挂载到了网络对应的linux bridge上了
	if err = netlink.LinkAdd(&endpoint.Device); err != nil {
		return fmt.Errorf("add endpoint device err: %w", err)
	}

	// 调用netlink的linkSetup方法，设置veth启动
	// 相当于ip link set xxx up 命令
	if err = netlink.LinkSetUp(&endpoint.Device); err != nil {
		return fmt.Errorf("add endpoint device setup err: %w", err)
	}
	return nil
}

func (b *BridgeNetworkDriver) Disconnect(network NetWork, endpoint *Endpoint) error {
	return nil
}

func (b *BridgeNetworkDriver) initBridge(n *NetWork) error {
	// try to get bridge by name, if it already exists then just exit
	bridgeName := n.Name
	if err := createBridgeInterface(bridgeName); err != nil {
		return err
	}
	log.Infof("createBridgeInterface success")

	// set bridge ip
	gatewayIp := *n.IpRange
	gatewayIp.IP = n.IpRange.IP

	if err := setInterfaceIp(bridgeName, gatewayIp.String()); err != nil {
		return err
	}
	log.Infof("setInterfaceIp success")

	if err := setInterfaceUp(bridgeName); err != nil {
		return err
	}
	log.Infof("crsetInterfaceUp success")

	if err := setupIpTables(bridgeName, n.IpRange); err != nil {
		return err
	}
	log.Infof("setInterfaceUp success")
	return nil
}

func (b *BridgeNetworkDriver) deleteBridge(n *NetWork) error {
	bridgeName := n.Name
	// get the link
	l, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return fmt.Errorf("get link with name %s failed: %w", bridgeName, err)
	}
	// delete the link
	if err := netlink.LinkDel(l); err != nil {
		return fmt.Errorf("remove bridge interface %s err: %w", bridgeName, err)
	}
	return nil
}

func createBridgeInterface(bridgeName string) error {
	_, err := net.InterfaceByName(bridgeName)
	if err == nil || !strings.Contains(err.Error(), "no such network interface") {
		return err
	}

	// create *netlink.Bridge object
	la := netlink.NewLinkAttrs()
	la.Name = bridgeName

	br := &netlink.Bridge{LinkAttrs: la}
	if err := netlink.LinkAdd(br); err != nil {
		return fmt.Errorf("bridge creating failed for bridge %s, err: %w", bridgeName, err)
	}
	return nil
}

func setInterfaceUp(interfaceName string) error {
	iface, err := netlink.LinkByName(interfaceName)
	if err != nil {
		return fmt.Errorf("error retrieving a link named [ %s ], err: %w", iface.Attrs().Name, err)
	}

	if err := netlink.LinkSetUp(iface); err != nil {
		return fmt.Errorf("error enabling interface for %s, err: %w", interfaceName, err)
	}
	log.Infof("ip link set up: %s -- %d", interfaceName, iface.Attrs().Index)
	return nil
}

func setInterfaceIp(name string, rawIp string) error {
	retries := 2
	var iface netlink.Link
	var err error
	for i := 0; i < retries; i++ {
		iface, err = netlink.LinkByName(name)
		if err == nil {
			break
		}
		log.Debugf("error retrieving new bridge netlink link [ %s ]... retrying", name)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("abandoning retrieving the new bridge ink from netlink, run [ip link] to troubleshoot the err: %w", err)
	}

	ipNet, err := netlink.ParseIPNet(rawIp)
	if err != nil {
		return fmt.Errorf("netlink.ParseIPNet err: %w", err)
	}

	addr := &netlink.Addr{
		IPNet:     ipNet,
		Peer:      ipNet,
		Label:     "",
		Flags:     0,
		Scope:     0,
		Broadcast: nil,
	}
	log.Infof("add addr, iface: %d -- %s, ip: %s", iface.Attrs().Index, iface.Attrs().Name, addr.String())
	return netlink.AddrAdd(iface, addr)
}

func setupIpTables(bridgeName string, subnet *net.IPNet) error {
	iptablesCmd := fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE", subnet.String(), bridgeName)
	cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("iptablse err %v, %w", output, err)
	}
	return nil
}
