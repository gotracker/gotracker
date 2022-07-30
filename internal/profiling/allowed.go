//go:build debug
// +build debug

package profiling

import (
	_ "net/http/pprof"
)

func init() {
	allowed = true
}
