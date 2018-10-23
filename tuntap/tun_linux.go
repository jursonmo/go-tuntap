package tuntap

import (
	//"bytes"
	"errors"
	"os"
	"syscall"
	"unsafe"
)

func openDevice(ifPattern string, useMytun bool) (*os.File, error) {
	var file *os.File
	var err error
	if useMytun {
		file, err = os.OpenFile("/dev/net/mytun", os.O_RDWR, 0)
	} else {
		file, err = os.OpenFile("/dev/net/tun", os.O_RDWR, 0)
	}
	return file, err
}

func createInterface(file *os.File, ifPattern string, kind DevKind, meta bool) (string, error) {
	var req ifReq
	//req.Flags = iffOneQueue
	/*about  IFF_MULTI_QUEUE,

	tun_set_iff():
	int queues = ifr->ifr_flags & IFF_MULTI_QUEUE ?
				 MAX_TAP_QUEUES : 1;

	dev = alloc_netdev_mqs(sizeof(struct tun_struct), name,
		NET_NAME_UNKNOWN, tun_setup, queues,
		queues);
	*/
	req.Flags = 0
	req.Flags |= 0x0100 //add by mo,multi
	if len(ifPattern) > 15 {
		return "", errors.New("tun/tap name too long")
	}
	copy(req.Name[:15], ifPattern)
	switch kind {
	case DevTun:
		req.Flags |= iffTun
	case DevTap:
		req.Flags |= iffTap
	default:
		panic("Unknown interface type")
	}
	if !meta {
		req.Flags |= iffnopi
	}
	/*
		如果IFF_NO_PI标志没有被设置，每一帧格式如下：
		Flags [2 bytes]
		Proto [2 bytes]
		Raw protocol(IP, IPv6, etc) frame.

		// Protocol info prepended to the packets (when IFF_NO_PI is not set)
		#define TUN_PKT_STRIP	0x0001  //#include <linux/if_tun.h>// include/uapi/linux/if_tun.h
		struct tun_pi {
			__u16  flags;
			__be16 proto;
		};
	*/
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), uintptr(syscall.TUNSETIFF), uintptr(unsafe.Pointer(&req)))
	if err != 0 {
		return "", err
	}
	//idxNull := bytes.IndexByte(req.Name[:], 0)
	//if idxNull < 0 {
	//	idxNull = len(req.Name)
	//}
	//return string(req.Name[:idxNull]), nil
	return string(req.Name[:len(ifPattern)]), nil
}
