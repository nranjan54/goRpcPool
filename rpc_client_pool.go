package rpcPool

import (
	"net"
	"net/rpc"
	"time"
	"strings"
	"errors"
	"fmt"
	"log"
)

type Pool struct {
	clients chan *rpc.Client
	conf    *RpcClientPoolConfig
}

func NewPool(conf *RpcClientPoolConfig) *Pool {
	p := new(Pool)
	p.conf = conf
	p.clients = make(chan *rpc.Client, p.conf.RpcMaxConns)
	client, err := p.dialNew()
	if err != nil {
		log.Printf("Failed to create rpc_client_conn_pool: %v", err)
		return nil
	}
	p.clients <- client
	return p
}

func (p *Pool) dialNew() (*rpc.Client, error) {
	var conn net.Conn
	var err error

	// This loop will retry before giving up.
	for i := 0; i < p.conf.RpcDialRetry; i++ {
		conn, err = net.DialTimeout(p.conf.RpcProtocol, p.conf.RpcServerAddress, time.Second * time.Duration(p.conf.RpcDialTimeout))

		if err == nil {
			break
		}
		if !strings.Contains(err.Error(), "refused") {
			break
		}

		log.Printf("Failed to connect to rpc_server %s", p.conf.RpcServerAddress)
		time.Sleep(time.Second * 10)
	}
	if err != nil {
		return nil, err
	}

	return rpc.NewClient(conn), nil
}

func (p *Pool) Call(serviceMethod string, args interface{}, reply interface{}) error {
	client, err := p.get()
	if err != nil {
		return err
	}

	done := make(chan error, 1)

	go func() {
		err := client.Call(serviceMethod, args, reply)
		done <- err
	}()

	// Waiting for reply util timeout.
	select {
	case <-time.After(time.Second * time.Duration(p.conf.RpcCallTimeout)):
		log.Printf("rpc call timeout %s => %s", serviceMethod, p.conf.RpcServerAddress)
		client.Close()
		return errors.New(fmt.Sprintf("rpc call timeout %s => %s", serviceMethod, p.conf.RpcServerAddress))
	case err := <-done:
		if err != nil {
			client.Close()
			return err
		}
	}

	// Put client back to conn_pool
	select {
	case p.clients <- client:
		return nil
	default:
		return client.Close()
	}

	return nil
}

func (p *Pool) get() (*rpc.Client, error) {
	select {
	case client := <-p.clients:
		return client, nil
	default:
		return p.dialNew()
	}
}
