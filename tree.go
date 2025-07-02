package gon

// Param 是单个URL参数的结构体，由一个键和一个值组成
type Param struct {
	Key   string
	Value string
}

// Params 是一个 Param 切片，由路由器返回。
// 该切片是有序的，第一个 URL 参数也是第一个切片值。
// 因此，通过索引读取值是安全的。
type Params []Param

// 路由树
type methodTree struct {
	method string // HTTP 方法
	root   *node  // 路由树的根节点
}

// methodTrees 是一个切片，存储不同 HTTP 方法对应的路由树
// 每个 HTTP 方法对应一棵路由树，最多存在 9 颗，初始化的时候会直接申请长度为 9 的切片
type methodTrees []methodTree

// 路由树上的节点
type node struct {
	path     string       // 路径
	children []*node      // 子节点
	handlers HandlerChain // 该节点对应的handler处理链
	fullPath string       // 完整路径，包含父节点的路径
}
