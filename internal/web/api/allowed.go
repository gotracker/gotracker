//go:build web || debug
// +build web debug

package api

func init() {
	allowed = true
}
