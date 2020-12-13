package feature

// Feature is an enumeration of player features that can be disabled by a device
type Feature int

const (
	featureUnknown = Feature(iota)

	// FeaturePatternLoop describes the pattern loop feature
	FeaturePatternLoop
)
