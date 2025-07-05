package bytesconv

import "unsafe"

// StringToBytes 将字符串转换为字节切片
// unsafe 包从 1.20 起正式支持 SliceData 、 String 和 StringData
func StringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// BytesToString 将字节切片转换为字符串
// unsafe 包从 1.20 起正式支持 SliceData 、 String 和 StringData
func BytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}