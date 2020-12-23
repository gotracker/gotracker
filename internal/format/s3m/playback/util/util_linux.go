// +build linux

package util

// BE16ToLE16 converts a big-endian uint16 to a little-endian uint16
func BE16ToLE16(be uint16) uint16 {
	return uint16(be>>8) | (be << 8)
}
