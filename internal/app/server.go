package app

import (
	"context"
	"fmt"
	"github.com/apache/thrift/lib/go/thrift"
	"github.com/fpawel/oxygen73/internal/pkg/must"
	"github.com/fpawel/oxygen73/internal/thriftgen/mainsvc"
	"github.com/jmoiron/sqlx"
	"net"
	"os"
	"strconv"
	"time"
)

func runServer(db *sqlx.DB) func() {
	port, errPort := strconv.Atoi(os.Getenv("OXYGEN73_API_PORT"))
	if errPort != nil {
		log.Debug("finding free port to serve api")
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			panic(err)
		}
		port = ln.Addr().(*net.TCPAddr).Port
		must.PanicIf(os.Setenv("OXYGEN73_API_PORT", strconv.Itoa(port)))
		must.PanicIf(ln.Close())
	}
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	log.Debug("serve api: " + addr)

	transport, err := thrift.NewTServerSocket(addr)
	if err != nil {
		panic(err)
	}
	handler := &mainSvcHandler{db: db}

	// "разогрев" БД

	_, _ = handler.ListYearMonths(context.Background())
	_, _ = handler.ListBucketsOfYearMonth(context.Background(), int32(time.Now().Year()), int32(time.Now().Month()))
	{
		var measurements []measurement
		t := time.Now()
		if err := db.Select(&measurements, `SELECT * FROM measurement LIMIT 100000`); err != nil {
			panic(err)
		}
		log.Printf("db: %d measurements, %v time", len(measurements), time.Since(t))
		measurements = nil
	}

	processor := mainsvc.NewMainSvcProcessor(handler)
	server := thrift.NewTSimpleServer4(processor, transport,
		thrift.NewTTransportFactory(), thrift.NewTBinaryProtocolFactoryDefault())

	go log.ErrIfFail(server.Serve, "problem", "`failed to serve`")

	return func() {
		log.ErrIfFail(server.Stop, "problem", "`failed to stop server`")
	}
}
