package utils

import (
	"context"
	"runtime"
)

type task struct {
	ctx   context.Context
	input interface{}
	f     TaskFunc
}

type TaskFunc func(context.Context, interface{}) (interface{}, error)

func NewTask(ctx context.Context, input interface{}, f TaskFunc) task {
	return task{ctx, input, f}
}

type Result struct {
	Value interface{}
	Error error
}

type WorkPool interface {
	Run()
	Done() chan Result
	AddTask(task)
}

type workPool struct {
	size      int
	taskQueue chan task
	resultCh  chan Result
	done      chan interface{}
}

var _ WorkPool = &workPool{}

func NewWorkPool(tasksize int) *workPool {
	return &workPool{
		size:      tasksize,
		taskQueue: make(chan task, tasksize),
		resultCh:  make(chan Result, tasksize),
		done:      make(chan interface{}),
	}
}

func (wp *workPool) Run() {
	ncpu := runtime.NumCPU()
	closeChs := make([]chan interface{}, 0, ncpu)
	for i := 0; i < ncpu; i++ {
		closeChs = append(closeChs, make(chan interface{}))
		go func(i int) {
			for t := range wp.taskQueue {
				select {
				case <-t.ctx.Done():
				default:
					v, err := t.f(t.ctx, t.input)
					wp.resultCh <- Result{v, err}
				}
			}
			close(closeChs[i])
		}(i)
	}

	go func() {
		for _, ch := range closeChs {
			<-ch
		}
		wp.done <- struct{}{}
	}()
}

func (wp *workPool) AddTask(t task) {
	wp.taskQueue <- t
}

func (wp *workPool) Done() chan Result {
	close(wp.taskQueue)
	<-wp.done
	close(wp.resultCh)
	return wp.resultCh
}
