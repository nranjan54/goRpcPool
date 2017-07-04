package rpcPool

import (
	"net"
	"testing"
	"errors"
	"net/rpc"
	"time"

	"github.com/stretchr/testify/assert"
)

type Args struct {
	A, B int
}

type Quotient struct {
	Quo, Rem int
}

type Arith int

func (t *Arith) Multiply(args *Args, reply *int) error {
	*reply = args.A * args.B
	return nil
}

func (t *Arith) Divide(args *Args, quo *Quotient) error {
	if args.B == 0 {
		return errors.New("divide by zero")
	}
	quo.Quo = args.A / args.B
	quo.Rem = args.A % args.B
	return nil
}

var (
	testMulArgs *Args = &Args{
		A : 10,
		B : 10,
	}
	testMulReply int = 100

	testDivArgs *Args = &Args{
		A : 93,
		B : 10,
	}
	testDivQuo *Quotient = &Quotient{
		Quo : 9,
		Rem : 3,
	}

	testAddress string = "127.0.0.1:2345"
	testConf	*RpcClientPoolConfig = &RpcClientPoolConfig{
		RpcProtocol		: "tcp",
		RpcServerAddress	: "127.0.0.1:2345",
		RpcDialTimeout		: 10,
		RpcDialRetry		: 3,
		RpcCallTimeout		: 10,
		RpcMaxConns		: 10000,
	}
)

var (
	pool	*Pool
	server	*rpc.Server
)

func runRpcServer(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			time.Sleep(time.Duration(100) * time.Millisecond)
			continue
		}
		go server.ServeConn(conn)
	}
}

func Test_RpcConnPool(t *testing.T) {
	// Create rpc server
	server = rpc.NewServer()
	server.Register(new(Arith))

	lis, err := net.Listen("tcp", testAddress)
	assert.Nil(t, err)

	go runRpcServer(lis)

	//time.Sleep(time.Second * 1)

	pool = NewPool(testConf)
	assert.NotNil(t, pool)

	var reply int
	pool.Call("Arith.Multiply", testMulArgs, &reply)
	assert.Equal(t, reply, testMulReply)

	quo := new(Quotient)
	pool.Call("Arith.Divide", testDivArgs, quo)
	assert.Equal(t, quo, testDivQuo)
}
