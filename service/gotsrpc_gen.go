// Code generated by gotsrpc https://github.com/foomo/gotsrpc/v2  - DO NOT EDIT.

package service

import (
	io "io"
	http "net/http"
	time "time"

	gotsrpc "github.com/foomo/gotsrpc/v2"
)

const (
	SiteContextServiceGoTSRPCProxyGetContext = "GetContext"
)

type SiteContextServiceGoTSRPCProxy struct {
	EndPoint string
	service  SiteContextService
}

func NewDefaultSiteContextServiceGoTSRPCProxy(service SiteContextService) *SiteContextServiceGoTSRPCProxy {
	return NewSiteContextServiceGoTSRPCProxy(service, "/service/sitecontextprovider")
}

func NewSiteContextServiceGoTSRPCProxy(service SiteContextService, endpoint string) *SiteContextServiceGoTSRPCProxy {
	return &SiteContextServiceGoTSRPCProxy{
		EndPoint: endpoint,
		service:  service,
	}
}

// ServeHTTP exposes your service
func (p *SiteContextServiceGoTSRPCProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	} else if r.Method != http.MethodPost {
		gotsrpc.ErrorMethodNotAllowed(w)
		return
	}
	defer io.Copy(io.Discard, r.Body) // Drain Request Body

	funcName := gotsrpc.GetCalledFunc(r, p.EndPoint)
	callStats, _ := gotsrpc.GetStatsForRequest(r)
	callStats.Func = funcName
	callStats.Package = "github.com/foomo/contentserver-mcp/service"
	callStats.Service = "SiteContextService"
	switch funcName {
	case SiteContextServiceGoTSRPCProxyGetContext:
		var (
			args []interface{}
			rets []interface{}
		)
		var (
			arg_path string
		)
		args = []interface{}{&arg_path}
		if err := gotsrpc.LoadArgs(&args, callStats, r); err != nil {
			gotsrpc.ErrorCouldNotLoadArgs(w)
			return
		}
		executionStart := time.Now()
		getContextRet, getContextRet_1 := p.service.GetContext(arg_path)
		callStats.Execution = time.Since(executionStart)
		rets = []interface{}{getContextRet, getContextRet_1}
		if err := gotsrpc.Reply(rets, callStats, r, w); err != nil {
			gotsrpc.ErrorCouldNotReply(w)
			return
		}
		gotsrpc.Monitor(w, r, args, rets, callStats)
		return
	default:
		gotsrpc.ClearStats(r)
		gotsrpc.ErrorFuncNotFound(w)
	}
}

const (
	ServiceGoTSRPCProxyGetDocument = "GetDocument"
)

type ServiceGoTSRPCProxy struct {
	EndPoint string
	service  Service
}

func NewDefaultServiceGoTSRPCProxy(service Service) *ServiceGoTSRPCProxy {
	return NewServiceGoTSRPCProxy(service, "/services/content")
}

func NewServiceGoTSRPCProxy(service Service, endpoint string) *ServiceGoTSRPCProxy {
	return &ServiceGoTSRPCProxy{
		EndPoint: endpoint,
		service:  service,
	}
}

// ServeHTTP exposes your service
func (p *ServiceGoTSRPCProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	} else if r.Method != http.MethodPost {
		gotsrpc.ErrorMethodNotAllowed(w)
		return
	}
	defer io.Copy(io.Discard, r.Body) // Drain Request Body

	funcName := gotsrpc.GetCalledFunc(r, p.EndPoint)
	callStats, _ := gotsrpc.GetStatsForRequest(r)
	callStats.Func = funcName
	callStats.Package = "github.com/foomo/contentserver-mcp/service"
	callStats.Service = "Service"
	switch funcName {
	case ServiceGoTSRPCProxyGetDocument:
		var (
			args []interface{}
			rets []interface{}
		)
		var (
			arg_path string
		)
		args = []interface{}{&arg_path}
		if err := gotsrpc.LoadArgs(&args, callStats, r); err != nil {
			gotsrpc.ErrorCouldNotLoadArgs(w)
			return
		}
		executionStart := time.Now()
		rw := gotsrpc.ResponseWriter{ResponseWriter: w}
		getDocumentRet, getDocumentRet_1 := p.service.GetDocument(&rw, r, arg_path)
		callStats.Execution = time.Since(executionStart)
		if rw.Status() == http.StatusOK {
			rets = []interface{}{getDocumentRet, getDocumentRet_1}
			if err := gotsrpc.Reply(rets, callStats, r, w); err != nil {
				gotsrpc.ErrorCouldNotReply(w)
				return
			}
		}
		gotsrpc.Monitor(w, r, args, rets, callStats)
		return
	default:
		gotsrpc.ClearStats(r)
		gotsrpc.ErrorFuncNotFound(w)
	}
}
