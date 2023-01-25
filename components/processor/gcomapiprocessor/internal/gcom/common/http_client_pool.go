package common

import "net/http"

type input struct {
	req      *http.Request
	outputCh chan<- output
}

type output struct {
	resp *http.Response
	err  error
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type httpPool struct {
	client *http.Client

	poolSize int
	workerCh chan<- input
}

func NewHTTPPool(client *http.Client, poolSize int) HTTPClient {
	workerCh := make(chan input)

	for i := 0; i < poolSize; i++ {
		go requestHandler(client, workerCh)
	}

	return httpPool{
		client:   client,
		poolSize: poolSize,
		workerCh: workerCh,
	}
}

func requestHandler(client *http.Client, inpCh <-chan input) {
	for inp := range inpCh {
		resp, err := client.Do(inp.req)
		inp.outputCh <- output{resp: resp, err: err}
	}
}

func (hp httpPool) Do(req *http.Request) (*http.Response, error) {
	outputCh := make(chan output)
	defer close(outputCh)

	hp.workerCh <- input{req: req, outputCh: outputCh}
	out := <-outputCh
	return out.resp, out.err
}
