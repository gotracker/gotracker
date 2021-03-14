package index

// Pattern is an index into the pattern list
type Pattern uint8

const (
	// NextPattern allows the order system the ability to kick to the next pattern
	NextPattern = Pattern(254)
	// InvalidPattern specifies an invalid pattern
	InvalidPattern = Pattern(255)
)
