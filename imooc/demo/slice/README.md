# slice
type slice struct {
	array unsafe.Pointer
	len   int
	cap   int
}



## append
> 当cap容量不足时会，重新生成array指针（找另一个空地占着）, 再把老的内容复制到新的内容中
grow cap 
