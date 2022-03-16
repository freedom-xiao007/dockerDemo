# 自己动手写Docker系列 -- 3.3使用命令管道优化参数传递
***

## 简介
目前docker demo中还是使用的系统原有proc，不怎么纯净，本篇中使用busybox来更换docker demo的系统挂载点

## 效果查看
下面是在没有修改代码前的我们的docker的挂载情况，可以看到非常的庞大，不纯净

```shell
➜  dockerDemo git:(main) ✗ ./main run -ti /bin/sh
{"level":"info","msg":"conmand all is /bin/sh","time":"2022-03-17T05:44:58+08:00"}
{"level":"info","msg":"memory cgroup path: /sys/fs/cgroup/memory/mydocker-cgroup","time":"2022-03-17T05:44:58+08:00"}
{"level":"info","msg":"memory cgroup path: /sys/fs/cgroup/memory/mydocker-cgroup","time":"2022-03-17T05:44:58+08:00"}
{"level":"info","msg":"parent process run","time":"2022-03-17T05:44:58+08:00"}
{"level":"info","msg":"init come on","time":"2022-03-17T05:44:58+08:00"}
{"level":"info","msg":"command: /bin/sh, args: [/bin/sh]","time":"2022-03-17T05:44:58+08:00"}
{"level":"info","msg":"RunContainerInitProcess command /bin/sh, args [/bin/sh]","time":"2022-03-17T05:44:58+08:00"}
{"level":"info","msg":"find path: /bin/sh","time":"2022-03-17T05:44:58+08:00"}
# pwd
/home/lw/code/go/dockerDemo
# ls -l
总用量 4668
drwxrwxr-x 2 lw lw    4096 3月  17 05:44 docs
drwxrwxr-x 3 lw lw    4096 3月   7 04:55 example
-rw-rw-r-- 1 lw lw     382 3月  12 10:18 go.mod
-rw-rw-r-- 1 lw lw    1965 3月  12 10:18 go.sum
-rw-rw-r-- 1 lw lw   11558 3月  12 10:18 LICENSE
-rwxrwxr-x 1 lw lw 4741951 3月  14 20:58 main
drwxrwxr-x 6 lw lw    4096 3月  12 10:20 mydocker
-rw-rw-r-- 1 lw lw     473 3月  12 10:18 README.md
# mount
/dev/sda2 on / type ext4 (rw,relatime,errors=remount-ro)
udev on /dev type devtmpfs (rw,nosuid,noexec,relatime,size=16103592k,nr_inodes=4025898,mode=755,inode64)
devpts on /dev/pts type devpts (rw,nosuid,noexec,relatime,gid=5,mode=620,ptmxmode=000)
tmpfs on /dev/shm type tmpfs (rw,nosuid,nodev,inode64)
hugetlbfs on /dev/hugepages type hugetlbfs (rw,relatime,pagesize=2M)
mqueue on /dev/mqueue type mqueue (rw,nosuid,nodev,noexec,relatime)
tmpfs on /run type tmpfs (rw,nosuid,nodev,noexec,relatime,size=3227448k,mode=755,inode64)
tmpfs on /run/lock type tmpfs (rw,nosuid,nodev,noexec,relatime,size=5120k,inode64)
tmpfs on /run/user/125 type tmpfs (rw,nosuid,nodev,relatime,size=3227444k,mode=700,uid=125,gid=130,inode64)
gvfsd-fuse on /run/user/125/gvfs type fuse.gvfsd-fuse (rw,nosuid,nodev,relatime,user_id=125,group_id=130)
tmpfs on /run/user/0 type tmpfs (rw,nosuid,nodev,relatime,size=3227444k,mode=700,inode64)
gvfsd-fuse on /run/user/0/gvfs type fuse.gvfsd-fuse (rw,nosuid,nodev,relatime,user_id=0,group_id=0)
nsfs on /run/docker/netns/d503d5c5cda8 type nsfs (rw)
sysfs on /sys type sysfs (rw,nosuid,nodev,noexec,relatime)
securityfs on /sys/kernel/security type securityfs (rw,nosuid,nodev,noexec,relatime)
tmpfs on /sys/fs/cgroup type tmpfs (ro,nosuid,nodev,noexec,mode=755,inode64)
cgroup2 on /sys/fs/cgroup/unified type cgroup2 (rw,nosuid,nodev,noexec,relatime,nsdelegate)
cgroup on /sys/fs/cgroup/systemd type cgroup (rw,nosuid,nodev,noexec,relatime,xattr,name=systemd)
cgroup on /sys/fs/cgroup/net_cls,net_prio type cgroup (rw,nosuid,nodev,noexec,relatime,net_cls,net_prio)
cgroup on /sys/fs/cgroup/pids type cgroup (rw,nosuid,nodev,noexec,relatime,pids)
cgroup on /sys/fs/cgroup/cpuset type cgroup (rw,nosuid,nodev,noexec,relatime,cpuset)
cgroup on /sys/fs/cgroup/cpu,cpuacct type cgroup (rw,nosuid,nodev,noexec,relatime,cpu,cpuacct)
cgroup on /sys/fs/cgroup/rdma type cgroup (rw,nosuid,nodev,noexec,relatime,rdma)
cgroup on /sys/fs/cgroup/misc type cgroup (rw,nosuid,nodev,noexec,relatime,misc)
cgroup on /sys/fs/cgroup/perf_event type cgroup (rw,nosuid,nodev,noexec,relatime,perf_event)
cgroup on /sys/fs/cgroup/blkio type cgroup (rw,nosuid,nodev,noexec,relatime,blkio)
cgroup on /sys/fs/cgroup/hugetlb type cgroup (rw,nosuid,nodev,noexec,relatime,hugetlb)
cgroup on /sys/fs/cgroup/memory type cgroup (rw,nosuid,nodev,noexec,relatime,memory)
cgroup on /sys/fs/cgroup/devices type cgroup (rw,nosuid,nodev,noexec,relatime,devices)
cgroup on /sys/fs/cgroup/freezer type cgroup (rw,nosuid,nodev,noexec,relatime,freezer)
pstore on /sys/fs/pstore type pstore (rw,nosuid,nodev,noexec,relatime)
efivarfs on /sys/firmware/efi/efivars type efivarfs (rw,nosuid,nodev,noexec,relatime)
none on /sys/fs/bpf type bpf (rw,nosuid,nodev,noexec,relatime,mode=700)
debugfs on /sys/kernel/debug type debugfs (rw,nosuid,nodev,noexec,relatime)
tracefs on /sys/kernel/tracing type tracefs (rw,nosuid,nodev,noexec,relatime)
fusectl on /sys/fs/fuse/connections type fusectl (rw,nosuid,nodev,noexec,relatime)
configfs on /sys/kernel/config type configfs (rw,nosuid,nodev,noexec,relatime)
proc on /proc type proc (rw,nosuid,nodev,noexec,relatime)
systemd-1 on /proc/sys/fs/binfmt_misc type autofs (rw,relatime,fd=28,pgrp=0,timeout=0,minproto=5,maxproto=5,direct,pipe_ino=26120)
/var/lib/snapd/snaps/core20_1361.snap on /snap/core20/1361 type squashfs (ro,nodev,relatime,x-gdu.hide)
/var/lib/snapd/snaps/bare_5.snap on /snap/bare/5 type squashfs (ro,nodev,relatime,x-gdu.hide)
/var/lib/snapd/snaps/gnome-3-38-2004_99.snap on /snap/gnome-3-38-2004/99 type squashfs (ro,nodev,relatime,x-gdu.hide)
/var/lib/snapd/snaps/gtk-common-themes_1519.snap on /snap/gtk-common-themes/1519 type squashfs (ro,nodev,relatime,x-gdu.hide)
/var/lib/snapd/snaps/snapd_14978.snap on /snap/snapd/14978 type squashfs (ro,nodev,relatime,x-gdu.hide)
/var/lib/snapd/snaps/snap-store_558.snap on /snap/snap-store/558 type squashfs (ro,nodev,relatime,x-gdu.hide)
/var/lib/snapd/snaps/core20_1376.snap on /snap/core20/1376 type squashfs (ro,nodev,relatime,x-gdu.hide)
/dev/nvme0n1p1 on /boot/efi type vfat (rw,relatime,fmask=0077,dmask=0077,codepage=437,iocharset=iso8859-1,shortname=mixed,errors=remount-ro)
/var/lib/snapd/snaps/snapd_15177.snap on /snap/snapd/15177 type squashfs (ro,nodev,relatime,x-gdu.hide)
overlay on /var/lib/docker/overlay2/d50aa7883be1a9bfc50a2519093575713c356885fd916d5973793492bb8d2d07/merged type overlay (rw,relatime,lowerdir=/var/lib/docker/overlay2/l/6DRQTGKE2LPQHTVVITKND2RW3J:/var/lib/docker/overlay2/l/YESCAXPAOBR6LE7TQHTNL3OF4Y,upperdir=/var/lib/docker/overlay2/d50aa7883be1a9bfc50a2519093575713c356885fd916d5973793492bb8d2d07/diff,workdir=/var/lib/docker/overlay2/d50aa7883be1a9bfc50a2519093575713c356885fd916d5973793492bb8d2d07/work)
proc on /proc type proc (rw,nosuid,nodev,noexec,relatime)
```

下面是跟着书中写的，使用busybox作为挂载点，创建容器，可以看到比较纯净了

```shell
➜  busybox git:(main) ✗ ./main run -ti /bin/sh
{"level":"info","msg":"memory cgroup path: /sys/fs/cgroup/memory/mydocker-cgroup","time":"2022-03-17T06:17:48+08:00"}
{"level":"info","msg":"memory cgroup path: /sys/fs/cgroup/memory/mydocker-cgroup","time":"2022-03-17T06:17:48+08:00"}
{"level":"info","msg":"all command is : /bin/sh","time":"2022-03-17T06:17:48+08:00"}
{"level":"info","msg":"parent process run","time":"2022-03-17T06:17:48+08:00"}
{"level":"info","msg":"init come on","time":"2022-03-17T06:17:48+08:00"}
{"level":"info","msg":"current location: /home/lw/code/go/dockerDemo/busybox","time":"2022-03-17T06:17:48+08:00"}
{"level":"info","msg":"find path: /bin/sh","time":"2022-03-17T06:17:48+08:00"}
/ # ls
bin   dev   etc   home  main  proc  root  sys   tmp   usr   var
/ # mount
/dev/sda2 on / type ext4 (rw,relatime,errors=remount-ro)
proc on /proc type proc (rw,nosuid,nodev,noexec,relatime)
/ #
```

## 环境依赖准备
使用busybox需要做一些准备工作：

1. 安装docker
2. 需要拉取busybox的docker 镜像，打包后解压以备使用

### 安装docker
需要在机器上安装docker，Ubuntu20可以参考：[https://docs.docker.com/engine/install/ubuntu/](https://docs.docker.com/engine/install/ubuntu/)

咔咔咔一顿敲命令就行了，这里就不再赘述了

### 拉取busybox的docker 镜像，打包后解压以备使用
操作过程的命令如下，照着运行即可

```shell
docker pull busybox

// 可以看到busybox才2M不到，真小
➜  dockerDemo git:(main) ✗ docker images
REPOSITORY    TAG       IMAGE ID       CREATED        SIZE
busybox       latest    2fb6fc2d97e1   5 days ago     1.24MB
hello-world   latest    feb5d9fea6a5   5 months ago   13.3kB

// 后台运行起来
➜  dockerDemo git:(main) ✗ docker run -d busybox top -b
29148027fac463baa870831bcbc4dcaea48d5f2253ebd95118b8c65136017021
➜  dockerDemo git:(main) ✗ docker ps
CONTAINER ID   IMAGE     COMMAND    CREATED          STATUS          PORTS     NAMES
29148027fac4   busybox   "top -b"   15 seconds ago   Up 15 seconds             loving_edison

// 打包运行的镜像，创建目录，解压到指定目录中
➜  dockerDemo git:(main) ✗ docker export -o /home/lw/Downloads/busybox.tar 29148027fac4
➜  dockerDemo git:(main) ✗ mkdir /opt/busybox
➜  dockerDemo git:(main) ✗ tar -xvf /home/lw/Downloads/busybox.tar -C /opt/busybox/

// 可以看到和系统根目录还挺像
➜  dockerDemo git:(main) ✗ ls /opt/busybox
bin  dev  etc  home  proc  root  sys  tmp  usr  var
```

这样，准备工作就完成了

## 代码编写
本篇中代码比较少，就修改一个init文件即可

pivot_root是一个重要的概念，可以参考书中描述和后面参考链接的一篇博客

照抄书中的代码跑不动，需要结合博客一起

> pivot_root是一个系统调用，主要功能是去改变当前的root文件系统。
> pivot_root可以将当前进程的root文件系统移动到put_old文件夹中，然后使new_root成为新的root 文件系统。
> new_root 和put_old必须不能同时存在当前root 的同一个文件系统中。
> pivot_root和chroot的主要区别是，pivot_root是把整个系统切换到一个新的root目录，而移除对之前root文件系统的依赖，这样你就能够umount原先的root文件系统。
> 而chroot是针对某个进程，系统的其他部分依旧运行于老的root目录中。

实现代码如下：

```go
func RunContainerInitProcess() error {
	// 用setupMount代替原来的相关挂载操作
	if err := setUpMount(); err != nil {
		return err
	}

	cmdArray := readUserCommand()
	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		log.Errorf("can't find exec path: %s %v", cmdArray[0], err)
		return err
	}
	log.Infof("find path: %s", path)
	if err := syscall.Exec(path, cmdArray, os.Environ()); err != nil {
		log.Errorf("syscall exec err: %v", err.Error())
	}
	return nil
}

// 初始化挂载点
func setUpMount() error {
	// 首先设置根目录为私有模式，防止影响pivot_root，这一步是书中没有的，但没有运行不起来
	if err := syscall.Mount("/", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, ""); err != nil {
		return fmt.Errorf("setUpMount Mount proc err: %v", err)
	}

	// 获取当前路径
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current location err: %v", err)
	}
	log.Infof("current location: %s", pwd)

	err = privotRoot(pwd)
	if err != nil {
		return err
	}

	// mount proc
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	err = syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	if err != nil {
		log.Errorf("proc挂载 failed: %v", err)
		return err
	}
	syscall.Mount("tmpfs", "/dev", "tempfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
	return nil
}

func privotRoot(root string) error {
	// 为了使当前root的老root和新root不在同一个文件系统下，我们把root重新mount一次
	// bind mount 是把相同的内容换了一个挂载点的挂载方法
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("mount rootfs to itself error: %v", err)
	}

	// 创建 rootfs、.pivot_root 存储 old_root
	pivotDir := filepath.Join(root, ".pivot_root")
	// 判断当前目录是否已有该文件夹
	if _, err := os.Stat(pivotDir); err == nil {
		// 存在则删除
		if err := os.Remove(pivotDir); err != nil {
			return err
		}
	}
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return fmt.Errorf("mkdir of pivot_root err: %v", err)
	}

	// pivot_root 到新的rootfs，老的old_root现在挂载在rootfs/.pivot_root上
	// 挂载点目前依然可以在mount命令中看到
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot_root err: %v", err)
	}

	// 修改当前工作目录到跟目录
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir root err: %v", err)
	}

	// 取消临时文件.pivot_root的挂载并删除它
	// 注意当前已经在根目录下，所以临时文件的目录也改变了
	pivotDir = filepath.Join("/", ".pivot_root")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root dir err: %v", err)
	}
	return os.Remove(pivotDir)
}
```

编写完成后，还有重要的一步：记得运行程序的当前目录要在busybox下

因为我们代码中需要挂载 /proc，如果在其他目录下，那是挂载不了这个的，但 busybox里面有这个目录

## 参考链接
- [动手实现一个docker引擎-3-实现文件系统隔离、Volume与镜像打包](https://blog.csdn.net/weixin_43988498/article/details/121307202)
