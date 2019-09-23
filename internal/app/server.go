package app

import (
	"github.com/apache/thrift/lib/go/thrift"
	"github.com/fpawel/oxygen73/internal/thriftgen/mainsvc"
	"golang.org/x/sys/windows/registry"
	"net"
)

func newServer(handler mainsvc.MainSvc) thrift.TServer {
	serverAddr := determineServerAddr()
	transport, err := thrift.NewTServerSocket(serverAddr)
	if err != nil {
		panic(err)
	}
	processor := mainsvc.NewMainSvcProcessor(handler)
	server := thrift.NewTSimpleServer4(processor, transport,
		thrift.NewTTransportFactory(), thrift.NewTBinaryProtocolFactoryDefault())
	log.Println(serverAddr)
	return server
}

func determineServerAddr() string {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	addr := ln.Addr().(*net.TCPAddr)
	regKey, _, err := registry.CreateKey(registry.CURRENT_USER, `oxygen73\tcp`, registry.ALL_ACCESS)
	if err != nil {
		panic(err)
	}

	if err := regKey.SetStringValue("main_ip", addr.IP.String()); err != nil {
		panic(err)
	}
	if err := regKey.SetDWordValue("main_port", uint32(addr.Port)); err != nil {
		panic(err)
	}
	log.ErrIfFail(ln.Close)

	ln, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	guiAddr := ln.Addr().(*net.TCPAddr)
	if err := regKey.SetStringValue("gui_ip", guiAddr.IP.String()); err != nil {
		panic(err)
	}
	if err := regKey.SetDWordValue("gui_port", uint32(guiAddr.Port)); err != nil {
		panic(err)
	}
	log.ErrIfFail(ln.Close)

	return addr.String()
}
