package gon

import "path"

// assert1 用来实现错误断言功能，不满足条件触发 panic
func assert1(guard bool, text string) {
	if !guard {
		panic(text)
	}
}

// lastChar 返回字符串的最后一个字符
func lastChar(str string) uint8 {
	if str == "" {
		panic("The length of the string can't be 0")
	}
	return str[len(str)-1]
}


// joinPaths 用于将绝对路径和相对路径合并成一个完整的路径
// absolutePath: 是前缀绝对路径
// relativePath: 是需要进行拼接的路径
func joinPaths(absolutePath, relativePath string) string {
	if relativePath == "" {
		return absolutePath
	}

	finalPath := path.Join(absolutePath, relativePath)
	if lastChar(finalPath) != '/' && lastChar(relativePath) == '/' {
		return finalPath + "/"
	}
	return finalPath
}