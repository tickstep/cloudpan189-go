package apiutil

import "github.com/tickstep/cloudpan189-go/library/requester"

type HttpClient interface {
	GetHttpClient() *requester.HTTPClient
}
