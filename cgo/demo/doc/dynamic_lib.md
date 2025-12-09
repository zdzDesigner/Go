# C 动态库创建步骤

## 实现过程

### 1. 创建头文件 (add.h)
```c
#ifndef ADD_H
#define ADD_H

int add(int a, int b);

#endif
```

### 2. 创建实现文件 (add.c)
```c
#include "add.h"

int add(int a, int b) {
    return a + b;
}
```

### 3. 编译为动态库
```bash
gcc -shared -fPIC -o libadd.so add.c
```

**参数说明：**
- `-shared`：生成共享库（动态库）
- `-fPIC`：生成位置无关代码（Position Independent Code），动态库必需
- `-o libadd.so`：输出文件名（Linux 约定以 lib 前缀，.so 后缀）

### 4. 验证结果
```bash
$ ls -lh libadd.so
-rwxrwxr-x 1 zdz zdz 16K 12月  9 10:07 libadd.so

$ file libadd.so
libadd.so: ELF 64-bit LSB shared object, x86-64, version 1 (SYSV), dynamically linked
```

## 关键点

1. **命名规范**：Linux 动态库命名为 `lib<name>.so`，CGO 链接时用 `-l<name>`
2. **PIC 必需**：动态库必须使用 `-fPIC` 编译，否则无法被多进程共享
3. **头文件**：提供接口声明，CGO 需要通过 `#include` 引用
4. **符号导出**：默认所有非 static 函数都会导出，可用 `nm -D libadd.so` 查看

## 下一步：CGO 调用

```go
// #cgo LDFLAGS: -L. -ladd
// #include "add.h"
import "C"

result := C.add(1, 2)
```
