// +build debug

package profiling

import (
	_ "net/http/pprof"
)

func init() {
	Allowed = true
}
