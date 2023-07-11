package main

import (
	"context"
	"github.com/dpull/my-sock5/socks5"
	"log"
	"strings"
)

type myRuleSet struct {
}

func (myRuleSet) Allow(ctx context.Context, req *socks5.Request) (context.Context, bool) {
	if strings.Contains(req.DestAddr.FQDN, "openai.com") {
		return ctx, true
	}
	if strings.Contains(req.DestAddr.FQDN, "cloudflare.com") {
		return ctx, true
	}
	if strings.Contains(req.DestAddr.FQDN, "auth0.com") {
		return ctx, true
	}
	log.Println("Block", req.DestAddr.FQDN)
	return ctx, false
}

func main() {
	conf := &socks5.Config{}
	conf.Rules = myRuleSet{}

	server, err := socks5.New(conf)
	if err != nil {
		log.Panicln(err)
	}

	if err := server.ListenAndServe("tcp", ":8241"); err != nil {
		log.Panicln(err)
	}
}
