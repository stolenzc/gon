package gon

import (
	"net/http"
	"regexp"
)

// regEnLetter 用于匹配请求方式是否都是大写字母，使用 regexp.MustCompile 编译一次后可以复用，以此提高性能
var regEnLetter = regexp.MustCompile(`^[A-Z]*$`)

// anyMethods 定义了 Engine 默认支持的9颗路由树的路由方法名
var anyMethods = []string{
	http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete,
	http.MethodHead, http.MethodPatch, http.MethodOptions, http.MethodTrace, http.MethodConnect,
}

// IRouter 继承 IRoutes 并扩展了路由组管理（Group() 方法），用于创建层级路由
// 默认情况下，Engine 和 RouterGroup 都实现了 IRouter 接口 (基于 Gin 1.10+ 版本)
type IRouter interface {
	IRoutes
	Group(string, ...HandlerFunc) *RouterGroup // 添加一个路由组并指定路由组的中间件
}

// IRoutes 定义了所有的路由处理函数的注册接口
// 默认情况下，Engine 和 RouterGroup 都实现了 IRoutes 接口
type IRoutes interface {
	Use(...HandlerFunc) IRoutes // 注册路由中间件（处理函数链）

	Handler(string, string, ...HandlerFunc) IRoutes // 传入请求方式和路径进行路由注册
	Any(string, ...HandlerFunc) IRoutes             // 注册任意请求方式的路由处理函数
	GET(string, ...HandlerFunc) IRoutes             // 注册 GET 请求的路由处理函数
	POST(string, ...HandlerFunc) IRoutes            // 注册 POST 请求的路由处理函数
	PUT(string, ...HandlerFunc) IRoutes             // 注册 PUT 请求的路由处理函数
	DELETE(string, ...HandlerFunc) IRoutes          // 注册 DELETE 请求的路由处理函数
	HEAD(string, ...HandlerFunc) IRoutes            // 注册 HEAD 请求的路由处理函数
	PATCH(string, ...HandlerFunc) IRoutes           // 注册 PATCH 请求的路由处理函数
	OPTIONS(string, ...HandlerFunc) IRoutes         // 注册 OPTIONS 请求的路由处理函数
	MATCH([]string, string, ...HandlerFunc) IRoutes // 注册多种请求方式的路由处理函数

	// TODO 后续实现静态文件解析方法
	// StaticFile(string, string) IRoutes                    // 注册静态文件路由
	// StaticFileFS(string, string, http.FileSystem) IRoutes // 注册静态文件路由，使用指定的文件系统
	// Static(string, string) IRoutes                        // 注册静态文件目录路由
	// StaticFS(string, http.FileSystem) IRoutes             // 注册静态文件目录路由，使用指定的文件系统

}

// RouterGroup 用于组织路由组，便于管理和分组路由
type RouterGroup struct {
	Handlers HandlerChain // 该路由组的处理链，如果该路由组是根路由组，则Handlers存储的就是全局中间件
	basePath string       // 路由组的基础路径，根路由组的 basePath 是 "/"
	engine   *Engine      // 指向引擎实例
	root     bool         // 是否是根路由组
}

// 确保 RouterGroup 实现了 IRouter 接口，防止后期改错，如果不满足，编译器会报错
var _ IRouter = (*RouterGroup)(nil)

// Use 用于给路由组添加中间件（处理函数链）
func (group *RouterGroup) Use(middlewares ...HandlerFunc) IRoutes {
	group.Handlers = append(group.Handlers, middlewares...)
	return group.returnObj()
}

// Group 用来创建一个新的 RouterGroup ，该 RouterGroup 继承当前的 RouterGroup 的处理链和基础路径
func (group *RouterGroup) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	return &RouterGroup{
		Handlers: group.combineHandlers(handlers),
		basePath: group.calculateAbsolutePath(relativePath),
		engine:   group.engine,
		root:     false, // 新创建的路由组不是根路由组
	}

}

// handler 真实实现路由注册的逻辑，后续 GET、POST 等方法会调用该方法进行注册
func (group *RouterGroup) handler(httpMethod, relativePath string, handlers HandlerChain) IRoutes{
	absolutePath := group.calculateAbsolutePath(relativePath)
	handlers = group.combineHandlers(handlers)
	group.engine.addRoute(httpMethod, absolutePath, handlers)
	return group.returnObj()
}


func (group *RouterGroup) Handler(method string, path string, handlers ...HandlerFunc) IRoutes {
	if matched := regEnLetter.MatchString(method); !matched {
		panic("http method " + method + " is not valid")
	}
	return group.handler(method, path, handlers)
}

func (group *RouterGroup) GET(path string, handlers ...HandlerFunc) IRoutes {
	return group.handler(http.MethodGet, path, handlers)
}

func (group *RouterGroup) POST(path string, handlers ...HandlerFunc) IRoutes {
	return group.handler(http.MethodPost, path, handlers)
}

func (group *RouterGroup) PUT(path string, handlers ...HandlerFunc) IRoutes {
	return group.handler(http.MethodPut, path, handlers)
}

func (group *RouterGroup) DELETE(path string, handlers ...HandlerFunc) IRoutes {
	return group.handler(http.MethodDelete, path, handlers)
}

func (group *RouterGroup) HEAD(path string, handlers ...HandlerFunc) IRoutes {
	return group.handler(http.MethodHead, path, handlers)
}

func (group *RouterGroup) PATCH(path string, handlers ...HandlerFunc) IRoutes {
	return group.handler(http.MethodPatch, path, handlers)
}

func (group *RouterGroup) OPTIONS(path string, handlers ...HandlerFunc) IRoutes {
	return group.handler(http.MethodOptions, path, handlers)
}

// Any 会将该方法和 path 注册到所有支持的 HTTP 方法上
// GET、POST、PUT、DELETE、HEAD、PATCH、OPTIONS、TRACE、CONNECT
func (group *RouterGroup) Any(path string, handlers ...HandlerFunc) IRoutes {
	for _, method := range anyMethods {
		group.handler(method, path, handlers)
	}
	return group.returnObj()
}

// MATCH 会将该方法和 path 注册到指定的 HTTP 方法上
func (group *RouterGroup) MATCH(methods []string, path string, handlers ...HandlerFunc) IRoutes {
	for _, method := range methods {
		group.handler(method, path, handlers)
	}
	return group.returnObj()
}

// combineHandlers 用于合并当前路由组的处理链和传入的处理函数链，使用深拷贝返回一个新的函数处理链
func (group *RouterGroup) combineHandlers(handlers HandlerChain) HandlerChain {
	finalSize := len(group.Handlers) + len(handlers)
	assert1(finalSize < int(abortIndex), "too many handlers")
	mergedHandlers := make(HandlerChain, 0, finalSize)

	// 深拷贝当前路由组的处理链和传入的处理函数链
	copy(mergedHandlers, group.Handlers)
	copy(mergedHandlers[len(group.Handlers):], handlers)
	return mergedHandlers
}

// calculateAbsolutePath 用于计算绝对路由路径
func (group *RouterGroup) calculateAbsolutePath(relativePath string) string {
	return joinPaths(group.basePath, relativePath)

}

// returnObj 返回当前路由组的对象，如果是根路由组则返回引擎实例
func (group *RouterGroup) returnObj() IRoutes {
	if group.root {
		return group.engine
	}
	return group
}
