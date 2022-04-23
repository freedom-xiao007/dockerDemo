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

// 配置容器网络端点的地址和路由
func configEndpointIpAddressAndRoute(ep *Endpoint, cinfo *container.ContainerInfo) error {
	// 通过网络端点中Veth的另一端
	peerLink, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		return fmt.Errorf("fail config endpoint err: %w", err)
	}
	log.Infof("peerLink index: %d -- %s", peerLink.Attrs().Index, peerLink.Attrs().Name)

	// 将容器的网络端点加入到容器的网络空间中
	// 并使这个函数下面的操作都在这个网络空间中进行
	// 执行完函数后，恢复为默认的网络空间，具体实现参考具体函数
	defer enterContainerNetns(&peerLink, cinfo, ep)

	return nil
}

// 将容器的网络端点加入到容器的网络空间中
// 并锁定当前程序所执行的线程，使当前线程进入到容器的网络空间
// 返回值是一个函数指针，执行这个返回函数才会退出容器的网络空间，回归到宿主机的网络空间
// 这个函数中引用了之前介绍的github.com/vishvananda/netns类库来做namespace操作
func enterContainerNetns(enLink *netlink.Link, cinfo *container.ContainerInfo, ep *Endpoint) func() {
	log.Infof("enterContainerNetns: %s", cinfo.Pid)

	// 找到容器的net namespa
	// /proc/{pid}/ns/net打开这个文件的文件描述符就可以来操作net namespace
	// 而containInfo中的PID，即容器在宿主机上映射的进程ID
	// 它对应的 /proc/{pid}/ns/net 就是容器内部的net namespace
	f, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", cinfo.Pid), os.O_RDONLY, 0)
	if err != nil {
		log.Errorf("error get container net namespace, %v", err)
	}

	// 取到文件的文件描述符
	nsFD := f.Fd()

	// 锁定当前程序所执行的线程，如果不锁定操作系统线程的话
	// go 语言的groutine可能会调度到别的线程上去
	// 就不能保证一致在所需要的网络空间中
	// 所以调用runtime。lockOSThread时要先锁定当前程序锁定的线程
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

	// 获取到容器的ip地址及网段，用于配置容器内部接口地址
	// 比如容器IP是192.168.1.2，而网络的网段是192.168.1.0/24
	// 那么这里产出的ip字符串就是192.168.1.2/24，用于容器内veth的配置
	interfaceIp := *ep.Network.IpRange
	interfaceIp.IP = ep.IpAddress
	log.Infof("容器的ip地址和网段：%s, %s", interfaceIp.String(), interfaceIp.IP.String())

	// 启动容器内的veth端点
	if err = setInterfaceUp(ep.Device.PeerName); err != nil {
		//return fmt.Errorf("setInterfaceUp ip %s, err: %w", ep.Device.PeerName, err)
		log.Errorf("setInterfaceUp ip %s, err: %w", ep.Device.PeerName, err)
	}
	// 调用函数设置容器内的Veth端点的IP
	if err = setInterfaceIp(ep.Device.PeerName, interfaceIp.String()); err != nil {
		//return fmt.Errorf("setinterface ip %v, err: %w", ep.Network, err)
		log.Errorf("setinterface ip %v, err: %w", ep.Network, err)
	}
	// net namespace中默认的本地地址是127.0.0.1 的lo网卡关闭状态
	// 启动它以保证容器访问自己的请求
	if err = setInterfaceUp("lo"); err != nil {
		//return fmt.Errorf("setInterfaceUp ip lo, err: %w", err)
		log.Errorf("setInterfaceUp ip lo, err: %w", err)
	}

	// 设置容器内的外部请求都通过容器内的veth端点网络
	// 0.0.0.0/0 的网段，标识所有的ip地址段
	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")

	// 构建要添加的路由数据，包括网络设备、网关IP及目的网段
	// 相当于route add -net 0.0.0.0/0 gw {bridge 网桥地址} dev {容器内的veth端点设备}
	defaultRoute := &netlink.Route{
		LinkIndex: (*enLink).Attrs().Index,
		Gw:        ep.Network.GatewayIP,
		Dst:       cidr,
	}
	log.Infof("default route: %s", defaultRoute.String())

	// 调用netlink的routeAdd，添加路由到容器内的网络空间
	// routeadd函数相当于route add命令
	if err = netlink.RouteAdd(defaultRoute); err != nil {
		//return fmt.Errorf("add route err: %w", err)
		log.Errorf("add route err: %w", err)
	}

	// 返回之前的net namespace
	// 在容器的网络空间中，执行完容器配置之后，调用此函数就可以将程序恢复到原生的net namespace
	return func() {
		// 恢复到上面获取到的之前的net namespace
		_ = netns.Set(originNet)
		// 关闭namespace
		_ = originNet.Close()
		// 取消对当前程序的线程锁定
		runtime.UnlockOSThread()
		// 关闭namespace文件
		_ = f.Close()
		log.Infof("退出容器网络空间")
	}
}

// 配置端口映射
func configPortMapping(ep *Endpoint, cinfo *container.ContainerInfo) error {
	log.Infof("端口映射：%s, %s", ep.IpAddress, ep.PortMapping)
	// 遍历容器端口映射列表
	for _, pm := range ep.PortMapping {
		// 分割宿主机的端口和容器的端口
		portMapping := strings.Split(pm, ":")
		if len(portMapping) != 2 {
			log.Errorf("port mapping format err: %v", pm)
			continue
		}

		// 由于iptable没有go语言版本的实现，所以采用exec.command的方式直接调用命令配置
		// 在iptables的PREROUTING中添加DNAT规则
		// 将宿主机的端口请求转发到容器的地址和端口上
		iptableCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s",
			portMapping[0], ep.IpAddress.String(), portMapping[1])
		log.Infof(iptableCmd)
		// 执行iptables命令，添加端口映射和转发规则
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
	_, ipNet, _ := net.ParseCIDR(network.Subnet)
	ip, err := ipAllocator.Allocate(ipNet)
	if err != nil {
		return err
	}
	log.Infof("分配的容器ip地址为: %s", ip)

	// 创建网络端点
	ep := &Endpoint{
		ID:          fmt.Sprintf("%s-%s", cinfo.ID, networkName),
		IpAddress:   ip,
		Network:     network,
		PortMapping: cinfo.PortMapping,
	}

	// 调用网络驱动挂载和配置网络端点
	if err = drivers[network.Driver].Connect(network, ep); err != nil {
		log.Errorf("connet err: %v", err)
		return err
	}

	// 配置端口映射
	if err = configPortMapping(ep, cinfo); err != nil {
		log.Errorf("config port mapping: %v", err)
		return err
	}

	// 到容器的namespace配置容器的网络设备IP地址
	if err = configEndpointIpAddressAndRoute(ep, cinfo); err != nil {
		log.Errorf("configEndpointIpAddressAndRoute err: %v", err)
		return err
	}

	return nil
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
