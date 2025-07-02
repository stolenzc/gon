package gon

import (
	"math"
	"net/http"
	"sync"
)

// abortIndex 表示中止函数中使用的典型值，其值为 127
// 会使用该值限制一个请求中的处理链函数的数量
const abortIndex int8 = math.MaxInt8 >> 1

// Context 用来存储请求上下文信息， 每个请求都会有一个独立的 Context 实例
type Context struct {
	Request  *http.Request       // HTTP 请求对象
	Writer   http.ResponseWriter // HTTP 可写入的响应
	handlers HandlerChain        // 当前请求的处理链
	index    int8                // 当前处理的中间件索引
	engine   *Engine             // 指回向入口 Engine
	params   *Params             // URL 参数列表，存储请求的 URL 参数

	mu       sync.RWMutex        // 读写锁，用于保护 Context 的并发访问

}
