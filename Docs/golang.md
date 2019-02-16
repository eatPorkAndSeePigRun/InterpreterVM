# golang

## Go写测试

1. 文件名以`_test.go`结尾

2. 必须`import "testing"`

3. 测试函数以`Test`开头，格式为`func TestXxx(t *testing.T)`

4. 用`testing.T`的`Error`，`Errorf`，`FailNow`，`Fatal`，`FatalIf`

5. 压力测试，格式为`func BenchmarkXxx(b *testing.B)`，循环中用`B.N`

6. 压力测试时，用`go test -test.bench=".*"`

## 指针

1. `unsafe.Pointer`用于转换指针

2. `uintptr`用于指针运算

3. `unsafe`包的三个接口：`func Sizeof(x ArbitraryType) uintptr`， `func Offsetof(x ArbitraryType) uintptr`， `func Alignof(x ArbitraryType) uintptr`

- [golang中传递中值传递以及指针传递](https://blog.csdn.net/gavin_new/article/details/80268905)

## fmt包

[输出-fmt包用法详解](https://godoc.org/fmt)

## interface

1. Commma-ok断言，如`v, ok := element.(T)`