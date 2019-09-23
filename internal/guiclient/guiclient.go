package guiclient

import (
	"context"
	"errors"
	"fmt"
	"github.com/apache/thrift/lib/go/thrift"
	"github.com/fpawel/oxygen73/internal/thriftgen/guisvc"
	"github.com/powerman/structlog"
	"golang.org/x/sys/windows/registry"
	"io"
	"sync"
)

var (
	ErrClosed = errors.New("client is closed")
)

func NotifyIfOpened(f func(c guisvc.GuiSvc) error) {
	go func() {
		mu.Lock()
		defer mu.Unlock()
		if client == nil {
			return
		}
		if err := f(client); err != nil {
			log.ErrIfFail(doClose)
		}
	}()
}

func ErrIfFail(f func(c guisvc.GuiSvc) error) {
	go log.ErrIfFail(func() error {
		return Notify(f)
	}, structlog.KeyStack, structlog.Auto)
}

func Notify(f func(c guisvc.GuiSvc) error) error {
	mu.Lock()
	defer mu.Unlock()
	if client == nil {
		return ErrClosed
	}
	err := f(client)
	if err != nil {
		log.ErrIfFail(doClose)
	}
	return err
}

func Close() error {
	mu.Lock()
	defer mu.Unlock()

	if client != nil {
		return doClose()
	}
	return nil
}

func Open() error {

	mu.Lock()
	defer mu.Unlock()

	if client != nil {
		log.ErrIfFail(doClose)
	}

	regKey, _, err := registry.CreateKey(registry.CURRENT_USER, `oxygen73\tcp`, registry.QUERY_VALUE)
	if err != nil {
		return err
	}
	guiServerIP, _, err := regKey.GetStringValue("gui_ip")
	if err != nil {
		return err
	}
	guiServerPort, _, err := regKey.GetIntegerValue("gui_port")
	if err != nil {
		return err
	}

	transport, err = thrift.NewTSocket(fmt.Sprintf("%s:%d", guiServerIP, guiServerPort))
	if err != nil {
		return err
	}

	transportFactory := thrift.NewTTransportFactory()

	transport, err = transportFactory.GetTransport(transport)
	if err != nil {
		return err
	}

	if err := transport.Open(); err != nil {
		return err
	}
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()

	iprot := protocolFactory.GetProtocol(transport)
	oprot := protocolFactory.GetProtocol(transport)

	client = guisvc.NewGuiSvcClient(thrift.NewTStandardClient(iprot, oprot))
	w.ctx, w.cancel = context.WithCancel(context.Background())
	return nil
}

func WriterNotifyConsole() io.Writer {
	return w
}

func doClose() error {
	w.cancel()
	client = nil
	return transport.Close()
}

var (
	log       = structlog.New()
	mu        sync.Mutex
	transport thrift.TTransport
	client    *guisvc.GuiSvcClient

	w = new(writer)
)

type writer struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func (x *writer) Write(p []byte) (int, error) {
	mu.Lock()
	defer mu.Unlock()
	ctx := x.ctx
	go NotifyIfOpened(func(c guisvc.GuiSvc) error {
		return c.NotifyWriteConsole(ctx, string(p))
	})
	return len(p), nil
}
