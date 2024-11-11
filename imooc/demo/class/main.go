package main

type Person struct {
	Sex bool
}

func (p *Person) GetName() {
	println("get name")
}

type Zdz struct {
	Person
}

func main() {
	zdz := Zdz{}
	zdz.GetName()
}
