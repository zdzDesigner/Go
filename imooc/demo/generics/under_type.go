package main


// `BasicInterface`基本接口类型，由于其仅包含方法元素，我们依旧可以基于之前讲过的方法集合，来确定一个类型是否实现了接口，以及是否可以作为*类型实参* 传递给 *约束下的类型形参*。
// `NonBasicInterface`对于只能作为约束的非基本接口类型，既有方法元素，也有类型元素，我们如何判断一个类型是否满足约束，并作为类型实参传给类型形参呢？
// 这时候我们就需要 Go 泛型落地时引入的新概念：类型集合（type set），类型集合将作为后续判断类型是否满足约束的基本手段。

type BasicInterface interface { // 基本接口类型
	M1()
}

type NonBasicInterface interface { // 非基本接口类型
	BasicInterface
	~int | ~string // 包含类型元素
}

type MyString string

func (MyString) M1() {
}

func foo[T NonBasicInterface](a T) { // 非基本接口类型作为约束
}

func bar[T BasicInterface](a T) { // 基本接口类型作为约束
}

type MyStruct[T interface{*int}] struct{} 
// type MyStruct[T *int,] struct{} // 简化版本


func main() {
	s := MyString("hello")
	var bi BasicInterface = s // 基本接口类型支持常规用法
	bi.M1()
	// var nbi NonBasicInterface = s // 非基本接口不支持常规用法，导致编译器错误：cannot use type NonBasicInterface outside a type constraint: interface contains type constraints
	// nbi.M1()
	foo(s)
	bar(s)
}
