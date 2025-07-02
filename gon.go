package gon

import (
	"sync"
)

// HandlerFunc 表示一个请求处理函数的类型
type HandlerFunc func(*Context)

// OptionFunc 用来定义修改默认配置的函数切片
type OptionFunc func(*Engine)

// HandlerChain 是一个handler函数的切片，用来存储一个请求的处理链
type HandlerChain []HandlerFunc

// Engine 是gon的核心引擎结构体，实现了 http.Handler 接口
type Engine struct {
	RouterGroup             // 路由组
	pool        sync.Pool   // 用于存储 Context 对象的池，减少内存分配和垃圾回收的开销
	trees       methodTrees // 存储不同 HTTP 方法的路由树
	maxParams   uint16      // TODO maxParams 数据含义
}


// New 创建一个新的 Engine 实例，返回指向 Engine 的指针
func New(opts ...OptionFunc) *Engine {
	engine := &Engine{
		RouterGroup: RouterGroup{
			Handlers: nil, // 根路由组的处理链为空
			basePath: "/",
			root:     true, // 根路由组
		},
		trees: make(methodTrees, 0, 9), // 初始化路由树切片，最多存储9种HTTP方法
	}

	engine.engine = engine // 设置根路由组 RouterGroup 引擎指针，指向自身

	engine.pool.New = func() any {
		return engine.allocateContext(engine.maxParams)
	}
	engine.With(opts...)
	return engine
}

// Default 返回已附加 Logger 和 Recovery 中间件的 Engine 实例。
func Default(opts ...OptionFunc) *Engine {
	// TODO debug 调试信息
	engine := New()
	// engine.Use(Logger(), Recovery()) // TODO 添加默认的 Logger 和 Recovery 中间件
	return engine.With(opts...)
}

// allocateContext
func (engine *Engine) allocateContext(maxParams uint16) *Context{
	v := make(Params, 0, maxParams)
	// TODO 补全这里的逻辑
	return &Context{engine: engine, params: &v}
}


// Use 是 Engine 的方法，用于添加中间件到根路由组的处理链中
func (engine *Engine) Use(middleware ...HandlerFunc) IRoutes {
	// 将传入的处理函数添加到根路由组的处理链中
	engine.RouterGroup.Use(middleware...)
	engine.Handlers = append(engine.Handlers, middleware...)
	return engine
}

// With 是 Engine 的方法，用于配置引擎的选项，可以传入一个修改配置的函数切片，然后在函数中修改engine的配置
func (engine *Engine) With(opts ...OptionFunc) *Engine {
	for _, opt := range opts {
		opt(engine)
	}
	return engine
}
