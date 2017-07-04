package rpcPool

type RpcClientPoolConfig struct {
	RpcProtocol		string		`yaml:"rpc_protocol"`
	RpcServerAddress	string		`yaml:"rpc_server_addrerss"`
	RpcDialTimeout		int		`yaml:"rpc_dial_timeout"`
	RpcDialRetry		int		`yaml:"rpc_dial_retry"`
	RpcCallTimeout		int		`yaml:"rpc_call_timeout"`
	RpcMaxConns		int		`yaml:"rpc_max_conns"`
}
