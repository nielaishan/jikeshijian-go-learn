package goroutinepool

import (
	"git.zuoyebang.cc/pkg/golib/v2/zlog"
	"github.com/gin-gonic/gin"
	"testing"
	"time"
)

var ctx *gin.Context

type Rsp struct {
	Name string
}

type Req struct {
	Name string
}

type HellowordDTO struct {
	Params *Req
	Err    error
	Data   *Rsp
}
/**
注意：业务逻辑没有错误，ERR务必置为nil

*/
func Helloword(ctx *gin.Context, input interface{}) bool {
	//zlog.Infof(ctx, "jobname start: %s", input)
	i, ok := input.(*HellowordDTO)
	if !ok {
		zlog.Warnf(ctx, "断言错误")
		return false
	}
	//zlog.Infof(ctx, "jobname: %v", i.Params)
	rsp := Rsp{
		Name: "job result",
	}
	i.Data = &rsp
	i.Err = nil
	return true
}

func TestPoolDemo(t *testing.T) {
	req :=  new(Req)
	req.Name = "job"
	dto := &HellowordDTO{
		Params: req,
	}
	pool, _:= NewJobPool(ctx, 1, 1)
	pool.AddJob(ctx, &Job{
		Input:   dto,
		Name:    "helloword",
		Func:    Helloword,
	})
	pool.Wait(ctx, 5 * time.Second)

}



