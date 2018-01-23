// Copyright (c) 2014 Ashley Jeffs
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package tunny

import (
	"errors"
	"sync"
)

var (
	ErrPoolNotRunning = errors.New("the pool is not running")
	ErrWorkerClosed   = errors.New("worker was closed")
)

type Worker interface {
	Process(interface{})
}

// Pool is a struct that manages a collection of workers, each with their own
// goroutine. The Pool can initialize, expand, compress and close the workers,
// as well as processing jobs with the workers synchronously.
type Pool struct {
	ctor    func() Worker
	workers []*workerWrapper
	reqChan chan workRequest

	workerMut sync.Mutex
}

// New creates a new Pool of workers that starts with n workers. You must
// provide a constructor function that creates new Worker types and when you
// change the size of the pool the constructor will be called to create each new
// Worker.
func New(workers []Worker) *Pool {
	p := &Pool{
		workers: make([]*workerWrapper, 0, len(workers)),
		//change queue length
		reqChan: make(chan workRequest, len(workers)),
	}
	// Add extra workers if N > len(workers)
	for i := 0; i < len(workers); i++ {
		p.workers = append(p.workers, newWorkerWrapper(p.reqChan, workers[i]))
	}

	return p
}

func (p *Pool) Process(payload interface{}) {
	request, open := <-p.reqChan
	if !open {
		panic(ErrPoolNotRunning)
	}

	request.jobChan <- payload
}

// Close will terminate all workers and close the job channel of this Pool.
func (p *Pool) Close() {
	for i := 0; i < len(p.workers); i++ {
		p.workers[i].stop()
	}
	for i := 0; i < len(p.workers); i++ {
		p.workers[i].join()
	}
	close(p.reqChan)
}
