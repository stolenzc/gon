package gon

import (
	"bytes"
	"strings"

	"github.com/stolenzc/gon/internal/bytesconv"
)

var (
	strColon = []byte(":") // 路由中的变量
	strStar  = []byte("*") // 路由中的模糊匹配
	strSlash = []byte("/") // 路由中分组符号
)

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

// get 方法用来在 methodTrees 中查找指定请求方式的路由树
func (trees methodTrees) get(method string) *node {
	// 遍历 methodTrees，查找对应的 method
	for _, tree := range trees {
		if tree.method == method {
			return tree.root
		}
	}
	return nil
}

// addChild 方法用于向节点中添加一个子节点
// 如果当前节点中已经存在通配符节点，那么通配符节点一定会在children 的最后一个位置，那么插入的节点就会在通配符节点之前
// 否则直接添加在 children 的末尾
func (n *node) addChild(child *node) {
	if n.wildChild && len(n.children) > 0 {
		wildChild := n.children[len(n.children)-1]
		n.children = append(n.children[:len(n.children)-1], child, wildChild)
	} else {
		n.children = append(n.children, child)
	}
}

// countParams 用来计算路径中参数的数量， 包括用 ":" 开头的参数和 "*" 通配符参数
func countParams(path string) uint16 {
	var n uint16
	s := bytesconv.StringToBytes(path)
	n += uint16(bytes.Count(s, strColon))
	n += uint16(bytes.Count(s, strStar))
	return n
}

func countSections(path string) uint16 {
	var n uint16
	s := bytesconv.StringToBytes(path)
	n += uint16(bytes.Count(s, strSlash))
	return n
}

type nodeType uint8

const (
	static   nodeType = iota // 静态节点， 可能存在于任何位置， path 需要完全匹配字符串
	root                     // 路由树的根节点，path 以 "/" 开头
	param                    // 参数节点，path 以 ":" 开头，例如："/users/:id" 中的 ":id"。如果不是末尾，需要以 "/" 结尾
	catchAll                 // 捕获所有路径的节点，使用 * 通配符，必须位于路径的末尾，例如："/files/*filepath" 中的 "*filepath"。
)

// 路由树上的节点
type node struct {
	path      string       // 当前节点的段路径，例如 "/users" 或 ":id" 或 "*filepath"
	indices   string       // 每个子节点path的首字符，顺序和children一致
	wildChild bool         // 是否包含通配符子节点，通配符子节点是指 path 以 ":" 或 "*" 开头的子节点，如果为true，那么通配符的子节点一定会是children 中的最后一个节点
	nType     nodeType     // 节点类型
	priority  int          // 经过该节点的路径数量，该数量会影响该 node 在父 node 的 children 中的顺序，数量越大，在父 node 的 children 中越靠前
	children  []*node      // 子节点
	handlers  HandlerChain // 该节点对应的handler处理链
	fullPath  string       // 完整路径，所有父节点的路径 + 当前节点的路径的拼接
}

// addRoute 方法用于在当前节点下添加一个路由
func (n *node) addRoute(path string, handlers HandlerChain) {

}

// findWildCard 用于在路径中查找通配符 "*"，返回通配符的字符串、位置和是否有效
// wildcard: 通配符
// wildcard: 通配符字符串，(: 或 * 以及后面的字符串，到 / 或到末尾，)
// n: 通配符在路径中的位置
// valid: 是否有效，true 表示有效，false 表示无效（例如出现了多个通配符 "/users/:id:name"）
// 如果没有找到，返回空字符串和 -1
func findWildCard(path string) (wildcard string, n int, valid bool) {
	for start, c := range []byte(path) {
		// 找到第一个通配符字符，确定通配符的起始位置
		if c != ':' && c != '*' {
			continue
		}

		// 寻找通配符的结束位置
		// 1. 如果找到 /, 则通配符到此结束
		// 2. 如果遇到了 : 或 *, 则无效，则次通配符不合法
		// 3. 如果循环完没有遇到上述任何一个情况。则通配符到路径的末尾结束
		valid = true
		for end, c := range []byte(path[start+1:]) {
			switch c {
			case '/':
				return path[start : start+1+end], start, valid
			case ':', '*':
				valid = false
			}
		}
		return path[start:], start, valid
	}
	// 到此处，说明path中没有通配符
	return "", -1, false
}

// insertChild 用于在当前节点下插入一个子节点
// 进入改方法时，n 本质上是一个 path 为空的 node，故某些条件下，可以直接操作 n 节点，不需要新建节点
func (n *node) insertChild(path string, fullPath string, handlers HandlerChain) {
	for {
		// 查找通配符位置
		wildcard, i, valid := findWildCard(path)
		// 没有找到通配符，直接退出循环
		if i < 0 {
			break
		}

		// 通配符不合法，直接 panic
		if !valid {
			panic("only one wildcard per path segment is allowed, has: '" +
				wildcard + "' in path '" + fullPath + "'")
		}

		// 如果存在通配符，那么至少有两个字符，一个通配符号 + 至少一个字符
		if len(wildcard) < 2 {
			panic("wildcards must be named with a non-empty name in path '" + fullPath + "'")
		}

		// 通配符为 : ，说明是一个参数节点
		if wildcard[0] == ':' {
			// i > 0 ，说明前面还有内容，修改path
			if i > 0 {
				n.path = path[:i]
				path = path[i:]
			}

			// 新建一个参数节点，使用 wildcard 作为 path
			child := &node{
				nType:    param,
				path:     wildcard,
				fullPath: fullPath,
			}

			n.addChild(child)
			n.wildChild = true // 标记当前节点包含通配符子节点，通配符子节点位于 children 的最后一个位置
			n = child
			n.priority++

			// 如果通配符后面还有内容，将 path 的内容修改为参数后面的内容，然后新建一个 path 为空的节点，继续循环
			if len(wildcard) < len(path) {
				path = path[len(wildcard):]
				child := &node{
					priority: 1,
					fullPath: fullPath,
				}
				n.addChild(child)
				n = child
				continue
			}

			// 否则，说明通配符是路径的最后一个节点，直接赋值 handlers
			n.handlers = handlers
			break
		}

		// 运行到此处，说明通配符是 *
		// 检查 * 通配符是否是路径的最后一个节点
		if i+len(wildcard) != len(path) {
			panic("catch-all routes are only allowed at the end of the path in path '" + fullPath + "'")
		}

		// 节点的路径结尾是 "/", 说明后面还有内容，会导致路径冲突不合法
		// 例如存在 /api/users/:id 后添加 /api/*filename
		if len(n.path) > 0 && n.path[len(n.path)-1] == '/' {
			pathSeg := ""
			if len(n.children) != 0 {
				pathSeg = strings.SplitN(n.children[0].path, "/", 2)[0]
			}
			panic("catch-all wildcard '" + path +
				"' in new path '" + fullPath +
				"' conflicts with existing path segment '" + pathSeg +
				"' in existing prefix '" + n.path + pathSeg +
				"'")
		}

		// 检查通配符前面的符号是否是 /
		i--
		if path[i] != '/' {
			panic("no / before catch-all in path '" + fullPath + "'")
		}

		// 把当前的空节点的path设置为通配符前 / 前面的内容
		n.path = path[:i]

		// 新建两个节点，分别用来做通配符前面的 "/" + 通配符节点
		// TODO 此处为什么需要一个 path 为空的节点
		child := &node{
			wildChild: true,
			nType:     catchAll,
			fullPath:  fullPath,
		}

		n.addChild(child)
		n.indices = string('/')
		n = child
		n.priority++

		// 通配符节点
		child = &node{
			path:     path[i:],
			nType:    catchAll,
			handlers: handlers,
			priority: 1,
			fullPath: fullPath,
		}

		n.children = []*node{child}

		return
	}

	// 到此处，说明没有通配符，直接把内容赋值给当前节点
	n.path = path
	n.handlers = handlers
	n.fullPath = fullPath
}
