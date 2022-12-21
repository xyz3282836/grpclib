package main

import (
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
)

// 定义 Scheme 名称
const discoveryScheme = "discovery"

type discoveryResolverBuilder struct {
	addrsStore map[string][]string
}

func NewDiscoveryResolverBuilder(addrsStore map[string][]string) *discoveryResolverBuilder {
	return &discoveryResolverBuilder{addrsStore: addrsStore}
}

func (drb *discoveryResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	// 初始化 resolver, 将 addrsStore 传递进去
	r := &discoveryResolver{
		target:     target,
		cc:         cc,
		addrsStore: drb.addrsStore,
	}
	// 调用 start 初始化地址
	r.start()
	return r, nil
}
func (e *discoveryResolverBuilder) Scheme() string { return discoveryScheme }

type discoveryResolver struct {
	target     resolver.Target
	cc         resolver.ClientConn
	addrsStore map[string][]string
}

func (r *discoveryResolver) start() {
	// 在静态路由表中查询此 Endpoint 对应 addrs
	addrStrs := r.addrsStore[r.target.Endpoint]
	addrs := make([]resolver.Address, len(addrStrs))
	for i, s := range addrStrs {
		addrs[i] = resolver.Address{Addr: s}
	}
	// addrs 列表转化为 state, 调用 cc.UpdateState 更新地址
	r.cc.UpdateState(resolver.State{Addresses: addrs})
}
func (*discoveryResolver) ResolveNow(o resolver.ResolveNowOptions) {}
func (*discoveryResolver) Close()                                  {}

func main() {

	// 注册我们的 resolver
	resolver.Register(NewDiscoveryResolverBuilder(map[string][]string{
		"xroom": []string{"127.0.0.1:8080", "127.0.0.1:8081"},
	}))

	// 建立对应 scheme 的连接, 并且配置负载均衡
	conn, err := grpc.Dial("discovery:///test", grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`))
	if err != nil {
		fmt.Println(err, conn)
	}
}
