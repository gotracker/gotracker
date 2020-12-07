package oscillator

type WaveTableSelect uint8

const (
	WaveTableSelectSine = WaveTableSelect(iota)
	WaveTableSelectSawtooth
	WaveTableSelectSquare
	WaveTableSelectRandom
)

type Oscillator struct {
	Table WaveTableSelect
	Pos   int8
}
