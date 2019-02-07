# golang

## Go写测试

1. 文件名以`_test.go`结尾

2. 必须`import "testing"`

3. 测试函数以`Test`开头，格式为`func TestXxx(t *testing.T)`

4. 用`testing.T`的`Error`，`Errorf`，`FailNow`，`Fatal`，`FatalIf`

5. 压力测试，格式为`func BenchmarkXxx(b *testing.B)`，循环中用`B.N`

6. 压力测试时，用`go test -test.bench=".*"`
