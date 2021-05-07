package goroutinepool

import (
	"github.com/gin-gonic/gin"
)

type jobFunc func(ctx *gin.Context, input interface{}) bool

type Job struct {
	Input interface{}
	Name  string
	Func  jobFunc
}
