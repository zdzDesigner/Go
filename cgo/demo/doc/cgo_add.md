# CGO 调用动态库实战

## 代码实现 (add.go)

```go
package main

// #cgo LDFLAGS: -L. -ladd
// #include "add.h"
import "C"
import "fmt"

func main() {
	result := C.add(10, 20)
	fmt.Printf("10 + 20 = %d\n", result)
}
```

## 关键要点

### 1. CGO 指令
```go
// #cgo LDFLAGS: -L. -ladd
```
- `#cgo`：CGO 编译指令
- `LDFLAGS`：链接器参数
- `-L.`：在当前目录查找库文件
- `-ladd`：链接 libadd.so（自动添加 lib 前缀和 .so 后缀）

### 2. 头文件引入
```go
// #include "add.h"
```
- 必须紧邻 `import "C"` 之前
- 引入 C 函数声明
- 头文件路径相对于当前 .go 文件

### 3. 特殊导入
```go
import "C"
```
- 伪包，触发 CGO 编译
- 与注释之间**不能有空行**
- 提供 `C` 命名空间访问 C 代码

### 4. 调用 C 函数
```go
result := C.add(10, 20)
```
- 通过 `C.` 前缀调用
- Go 的 int 自动转换为 C 的 int
- 返回值自动转换为 Go 类型

## 运行结果

```bash
$ go run add.go
10 + 20 = 30
```

## 编译说明

### 直接运行
```bash
go run add.go
```

### 编译为可执行文件
```bash
go build add.go
./add
```

### 运行时库路径
如果动态库不在系统路径，需要设置 `LD_LIBRARY_PATH`：
```bash
export LD_LIBRARY_PATH=.:$LD_LIBRARY_PATH
./add
```

或使用 `rpath` 嵌入路径：
```go
// #cgo LDFLAGS: -L. -ladd -Wl,-rpath=.
```

## 类型转换对照

| C 类型 | Go 类型 | 说明 |
|--------|---------|------|
| int    | C.int   | 32位整数 |
| long   | C.long  | 平台相关 |
| char*  | *C.char | 字符串指针 |
| void*  | unsafe.Pointer | 通用指针 |

## 常见错误

1. **找不到库文件**
   ```
   error: ld: library not found for -ladd
   ```
   解决：检查 `-L` 路径是否正确

2. **注释和 import 之间有空行**
   ```
   could not determine kind of name for C.add
   ```
   解决：删除 `import "C"` 前的空行

3. **运行时找不到 .so**
   ```
   error while loading shared libraries: libadd.so
   ```
   解决：设置 `LD_LIBRARY_PATH` 或使用 `-rpath`
