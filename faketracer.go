//go:build !armtracer
// +build !armtracer

package armtracer

func Begin() {}
func End()   {}
func BeginTrace(name string) int64 {
	return 0
}
func EndTrace(id int64) {}
