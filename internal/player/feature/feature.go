package feature

// Feature is an enumeration of player features that can be disabled by a device
type Feature int

const (
	featureUnknown = Feature(iota)

	// OrderLoop describes the pattern loop feature
	OrderLoop

	// PlayerSleepInterval describes the player sleep interval feature
	PlayerSleepInterval

	// IgnoreUnknownEffect describes a bypass/ignore of unknown effects
	IgnoreUnknownEffect
)

func init() {
	_ = featureUnknown // lint
}
