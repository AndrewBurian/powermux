package powermux

import "sync"

type executionPool struct {
	p *sync.Pool
}

func (ep *executionPool) Get() *routeExecution {
	return ep.p.Get().(*routeExecution)
}

func (ep *executionPool) Put(ex *routeExecution) {
	ex.middleware = ex.middleware[0:0]
	for key := range ex.params {
		delete(ex.params, key)
	}
	ex.handler = nil
	ex.notFound = nil
	ep.p.Put(ex)
}

func createExecution() interface{} {
	return &routeExecution{
		middleware: make([]Middleware, 0),
		params:     make(map[string]string),
	}
}

func newExecutionPool() *executionPool {
	return &executionPool{
		p: &sync.Pool{
			New: createExecution,
		},
	}
}
