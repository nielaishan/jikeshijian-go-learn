package goroutinepool

import (
	"git.zuoyebang.cc/pkg/golib/v2/zlog"
	"github.com/gin-gonic/gin"
	"github.com/panjf2000/ants"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultTimeOut = 5 * time.Second

	goPool_Uninitialized int32 = iota
	goPool_Start
	goPool_Running
	goPool_Stoping
	goPool_Stop

	goroutineMaxNum = 100
)

type execUnit struct {
	job  *Job
	pool *pool
	ctx  *gin.Context
}

type pool struct {
	mutex     sync.Mutex
	jobNum    int
	execs     []*execUnit
	pool      *ants.PoolWithFunc
	chRet     chan bool
	waitGroup *sync.WaitGroup
	status    int32
}

func TemplateFunc(i interface{}) {
	exe, ok := i.(*execUnit)
	if exe == nil || exe.pool == nil || exe.job == nil || !ok {
		panic("线程池断言失败")
	}
	defer exe.pool.waitGroup.Done()
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 2048)
			n := runtime.Stack(buf, false)
			zlog.Errorf(exe.ctx, "run job with panic: %v, %s", err, buf[:n])
		}
	}()
	exe.pool.chRet <- exe.job.Func(exe.ctx, exe.job.Input)
}

/**
goroutineNum 协程数量
jobNum：任务数,，jobNum >= 实际任务数，一定要
*/
func NewJobPool(ctx *gin.Context, goroutineNum int, jobNum int) (*pool, error) {
	jobPool := &pool{}
	atomic.SwapInt32(&(jobPool.status), goPool_Start)

	jobPool.execs = make([]*execUnit, 0)
	jobPool.chRet = make(chan bool, jobNum)
	jobPool.jobNum = jobNum
	jobPool.waitGroup = &sync.WaitGroup{}

	if goroutineNum > goroutineMaxNum {
		goroutineNum = goroutineMaxNum
	}

	pool, err := ants.NewPoolWithFunc(goroutineNum, func(i interface{}) {
		TemplateFunc(i)
	})
	if err != nil {
		zlog.Warnf(ctx, "jobPool init fail error: %v", err)
		return nil, err
	}
	jobPool.pool = pool
	atomic.SwapInt32(&(jobPool.status), goPool_Running)
	return jobPool, nil
}

//[]*Job
func (p *pool) AddJob(ctx *gin.Context, jobs ...*Job) {
	if atomic.LoadInt32(&(p.status)) != goPool_Running {
		zlog.Warn(ctx, "this goroutine pool is not a running stat! the stat is ", p.status)
		return
	}
	for _, job := range jobs {
		exe := &execUnit{
			pool: p,
			job:  job,
			ctx:  ctx,
		}
		p.waitGroup.Add(1)
		err := p.pool.Invoke(exe)
		if err != nil {
			p.chRet <- false
			p.waitGroup.Done()
			zlog.Warnf(ctx, "job submit fail err: %v", err)
		}
		p.execs = append(p.execs, exe)
	}
}

func (p *pool) Stop(ctx *gin.Context) {

	if atomic.LoadInt32(&(p.status)) != goPool_Running {
		zlog.Warn(ctx, "this goroutine pool is not a running stat,can't stop ! the stat is ", p.status)
		return
	}

	atomic.SwapInt32(&(p.status), goPool_Stoping)

	p.waitGroup.Wait()
	p.pool.Release()

	p.mutex.Lock()
	close(p.chRet)
	p.mutex.Unlock()
	atomic.SwapInt32(&(p.status), goPool_Stop)
}

func (p *pool) Wait(ctx *gin.Context, timeout time.Duration) {

	if atomic.LoadInt32(&(p.status)) != goPool_Running {
		zlog.Warn(ctx, "this goroutine pool is not a running stat,can't stop ! the stat is ", p.status)
		return
	}
	if timeout == 0 {
		timeout = defaultTimeOut
	}
	timer := time.NewTimer(timeout)

	done := make(chan bool)
	//等待线程运行完毕
	go func() {
		finishCnt := 0
		for {
			select {
			case _, ok := <-p.chRet:
				if !ok {
					zlog.Error(ctx, "p.chRet is close")
					return
				}
				finishCnt++
			case <-timer.C:
				return
			}
			if finishCnt >= len(p.execs) {
				done <- true
				return
			}
		}

	}()

	select {
	case <-time.After(timeout):
		zlog.Warnf(ctx, "wait timeout when it waits tasks that running , time: %s", timeout.String())
		return
	case <-done:
		zlog.Info(ctx, "all jobs run success")
		return
	}
}
