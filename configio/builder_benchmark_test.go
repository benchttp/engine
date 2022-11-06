package configio_test

import (
	"strconv"
	"testing"

	"github.com/benchttp/sdk/configio"
)

/*
Output:

Running tool: /Users/greg/sdk/go1.17.6/bin/go test -benchmem -run=^$ -bench ^(BenchmarkBuilder)$ github.com/benchttp/sdk/configio

goos: darwin
goarch: amd64
pkg: github.com/benchttp/sdk/configio
cpu: Intel(R) Core(TM) i5-8279U CPU @ 2.40GHz
BenchmarkBuilder/setters/Builder_pipe-8                 12966073                97.47 ns/op           40 B/op          2 allocs/op
BenchmarkBuilder/setters/Builder_append-8               21533343                81.65 ns/op           64 B/op          1 allocs/op
BenchmarkBuilder/build/Builder_pipe_100-8                3044132               396.5 ns/op            96 B/op          1 allocs/op
BenchmarkBuilder/build/Builder_append_100-8              5413728               222.9 ns/op            96 B/op          1 allocs/op
BenchmarkBuilder/build/Builder_pipe_10000-8                25195             41284 ns/op              96 B/op          1 allocs/op
BenchmarkBuilder/build/Builder_append_10000-8              71463             16838 ns/op              96 B/op          1 allocs/op
BenchmarkBuilder/build/Builder_pipe_1000000-8                156           6993338 ns/op              96 B/op          1 allocs/op
BenchmarkBuilder/build/Builder_append_1000000-8              451           2730270 ns/op              96 B/op          1 allocs/op
*/
func BenchmarkBuilder(b *testing.B) {
	b.Run("setters", func(b *testing.B) {
		b.Run("Builder_pipe", func(b *testing.B) {
			builder := configio.Builder{}
			for i := 0; i < b.N; i++ {
				builder.SetConcurrency(100)
			}
		})
		b.Run("Builder_append", func(b *testing.B) {
			builder := configio.Builder_append{}
			for i := 0; i < b.N; i++ {
				builder.SetConcurrency(100)
			}
		})
	})

	b.Run("build", func(b *testing.B) {
		for _, iter := range []int{100, 10_000, 1_000_000} {
			b.Run("Builder_pipe_"+strconv.Itoa(iter), func(b *testing.B) {
				builder := configio.Builder{}
				setupBuilder(b, &builder, iter)
				for i := 0; i < b.N; i++ {
					_ = builder.Runner()
				}
			})
			b.Run("Builder_append_"+strconv.Itoa(iter), func(b *testing.B) {
				builder := configio.Builder_append{}
				setupBuilder(b, &builder, iter)
				for i := 0; i < b.N; i++ {
					_ = builder.Runner()
				}
			})
		}
	})
}

func setupBuilder(
	b *testing.B,
	builder interface{ SetConcurrency(int) },
	iter int,
) {
	b.Helper()
	values := []int{-100, 100}
	for i := 0; i < iter; i++ {
		builder.SetConcurrency(values[i%2])
	}
	b.ResetTimer()
}
