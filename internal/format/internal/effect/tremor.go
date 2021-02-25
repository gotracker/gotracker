package effect

// Tremor is the storage object for the tremor values
type Tremor struct {
	off  bool
	tick int
}

// IsActive returns the on-state of the tremor object
func (t *Tremor) IsActive() bool {
	return !t.off
}

// Advance updates the tremor counter and resets it to a value if it triggers
func (t *Tremor) Advance() int {
	t.tick++
	return t.tick
}

// ToggleAndReset toggles the on-state of the tremor object and resets the counter to 0
func (t *Tremor) ToggleAndReset() {
	t.off = !t.off
	t.tick = 0
}

// Reset resets the tremor to zeroes
func (t *Tremor) Reset() {
	t.tick = 0
	t.off = false
}
