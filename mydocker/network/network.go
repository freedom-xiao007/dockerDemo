package network

import (
	"dockerDemo/mydocker/container"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"
)

var (
	defaultNetworkPath = "/var/run/mydocker/network/network/"
	drivers            = map[string]NetworkDriver{}
	networks           = map[string]*NetWork{}
)

type Endpoint struct {
	ID          string           `json:"id"`
	Device      netlink.Veth     `json:"device"`
	IpAddress   net.IP           `json:"ipAddress"`
	MacAddress  net.HardwareAddr `json:"macAddress"`
	Network     *NetWork         `json:"network"`
	PortMapping []string
}

type NetWork struct {
	Name      string
	IpRange   *net.IPNet
	Driver    string
	GatewayIP net.IP
	Subnet    string
}

type NetworkDriver interface {
	Name() string
	Create(subnet string, name string) (*NetWork, error)
	Delete(network NetWork) error
	Connect(network *NetWork, endpoint *Endpoint) error
	Disconnect(network NetWork, endpoint *Endpoint) error
}

func (nw *NetWork) dump(dumpPath string) error {
	if _, err := os.Stat(dumpPath); err != nil {
		if os.IsNotExist(err) {
			_ = os.MkdirAll(dumpPath, 0644)
		} else {
			return fmt.Errorf("dump path err: %w", err)
		}
	}

	nwPath := path.Join(dumpPath, nw.Name)
	nwFile, err := os.OpenFile(nwPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("open file %s, err: %w", nwPath, err)
	}
	defer nwFile.Close()

	nwJson, err := json.Marshal(nw)
	if err != nil {
		return fmt.Errorf("%s file json marshal err: %w", nwPath, err)
	}

	_, err = nwFile.Write(nwJson)
	if err != nil {
		return fmt.Errorf("save network config json err: %w", err)
	}
	return nil
}

func (nw *NetWork) remove(dumpPath string) error {
	if _, err := os.Stat(path.Join(dumpPath, nw.Name)); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return fmt.Errorf("remvove path err: %w", err)
		}
	} else {
		return os.Remove(path.Join(dumpPath, nw.Name))
	}
}

func (nw *NetWork) load(dumpPath string) error {
	nwConfigFile, err := os.Open(dumpPath)
	defer nwConfigFile.Close()
	if err != nil {
		return fmt.Errorf("open file err: %w", err)
	}

	nwJson := make([]byte, 2000)
	n, err := nwConfigFile.Read(nwJson)
	if err != nil {
		return fmt.Errorf("read file err： %w", err)
	}

	err = json.Unmarshal(nwJson[:n], nw)
	if err != nil {
		return fmt.Errorf("load nw info err: %w", err)
	}
	return nil
}

func Init() error {
	var bridgeDriver = BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = &bridgeDriver

	if _, err := os.Stat(defaultNetworkPath); err != nil {
		if os.IsNotExist(err) {
			_ = os.MkdirAll(defaultNetworkPath, 0644)
		} else {
			return err
		}
	}

	_ = filepath.Walk(defaultNetworkPath, func(nwPath string, info os.FileInfo, err error) error {
		if strings.HasSuffix(nwPath, "/") {
			return nil
		}
		_, nwName := path.Split(nwPath)
		nw := &NetWork{
			Name: nwName,
		}

		if err := nw.load(nwPath); err != nil {
			log.Errorf("error load network: %v", err)
		}

		networks[nwName] = nw
		return nil
	})
	return nil
}

func enterContainerNetns(enLink *netlink.Link, cinfo *container.ContainerInfo) func() {
	f, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", cinfo.Pid), os.O_RDONLY, 0)
	if err != nil {
		log.Errorf("error get container net namespace, %v", err)
	}

	nsFD := f.Fd()
	runtime.LockOSThread()

	// 修改veth peer另外一端移动容器的namespace中
	if err = netlink.LinkSetNsFd(*enLink, int(nsFD)); err != nil {
		log.Errorf("error set link netns, %v", err)
	}

	// 获取当前网络的namespace
	originNet, err := netns.Get()
	if err != nil {
		log.Errorf("get current netns err: %v", err)
	}

	// 设置当前进程到新的网络namespace，并咋函数执行完成后，再恢复到之前的namespace
	if err = netns.Set(netns.NsHandle(nsFD)); err != nil {
		log.Errorf("error set netns, %v", err)
	}
	return func() {
		netns.Set(originNet)
		originNet.Close()
		runtime.UnlockOSThread()
		f.Close()
	}
}

func configEndpointIpAddressAndRoute(ep *Endpoint, cinfo *container.ContainerInfo) error {
	peerLink, err := netlink.LinkByName(ep.Device.Name)
	if err != nil {
		return fmt.Errorf("fail config endpoint err: %w", err)
	}
	defer enterContainerNetns(&peerLink, cinfo)

	interfaceIp := *ep.Network.IpRange
	interfaceIp.IP = ep.IpAddress

	if err = setInterfaceIp(ep.Device.PeerName, interfaceIp.String()); err != nil {
		return fmt.Errorf("setinterface ip %v, err: %w", ep.Network, err)
	}
	if err = setInterfaceUp(ep.Device.PeerName); err != nil {
		return fmt.Errorf("setInterfaceUp ip %s, err: %w", ep.Device.PeerName, err)
	}
	if err = setInterfaceUp("lo"); err != nil {
		return fmt.Errorf("setInterfaceUp ip lo, err: %w", err)
	}

	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")
	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw:        ep.Network.IpRange.IP,
		Dst:       cidr,
	}
	if err = netlink.RouteAdd(defaultRoute); err != nil {
		return fmt.Errorf("add route err: %w", err)
	}
	return nil
}

func configPortMapping(ep *Endpoint, cinfo *container.ContainerInfo) error {
	for _, pm := range ep.PortMapping {
		portMapping := strings.Split(pm, ":")
		if len(portMapping) != 2 {
			log.Errorf("port mapping format err: %v", pm)
			continue
		}

		iptableCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destintion %s:%s",
			portMapping[0], ep.IpAddress.String(), portMapping[1])
		cmd := exec.Command("iptables", strings.Split(iptableCmd, " ")...)
		output, err := cmd.Output()
		if err != nil {
			log.Errorf("iptables output err:%s -- %v", output, err)
			continue
		}
	}
	return nil
}

func Connect(networkName string, cinfo *container.ContainerInfo) error {
	network, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("no such network: %s", networkName)
	}

	// 分配容器IP地址
	ip, err := ipAllocator.Allocate(network.IpRange)
	if err != nil {
		return err
	}

	// 创建网络端点
	ep := &Endpoint{
		ID:          fmt.Sprintf("%s-%s", cinfo.ID, networkName),
		IpAddress:   ip,
		Network:     network,
		PortMapping: cinfo.PortMapping,
	}

	// 调用网络驱动挂载和配置网络端点
	if err = drivers[network.Driver].Connect(network, ep); err != nil {
		return err
	}
	// 到容器的namespace配置容器的网络设备IP地址
	if err = configEndpointIpAddressAndRoute(ep, cinfo); err != nil {
		return err
	}

	return configPortMapping(ep, cinfo)
}

func CreateNetwork(driver, subnet, name string) error {
	nw, err := drivers[driver].Create(subnet, name)
	if err != nil {
		return err
	}
	log.Infof("create network success")
	return nw.dump(defaultNetworkPath)
}

func ListNetwork() error {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	_, _ = fmt.Fprint(w, "NAME\tIpRange\tDriver\n")
	for _, nw := range networks {
		fmt.Fprintf(w, "%s\t%s\t\t%s\n",
			nw.Name,
			nw.IpRange.String(),
			nw.Driver,
		)
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("flush error: %w", err)
	}
	return nil
}

func DeleteNetwork(networkName string) error {
	nw, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("no such network: %s", networkName)
	}

	_, ipNet, _ := net.ParseCIDR(nw.Subnet)
	if err := ipAllocator.Release(ipNet, &nw.GatewayIP); err != nil {
		return fmt.Errorf("remove network gateway ip err: %w", err)
	}

	if err := drivers[nw.Driver].Delete(*nw); err != nil {
		return fmt.Errorf("remove network driver err: %w", err)
	}

	return nw.remove(defaultNetworkPath)
}
