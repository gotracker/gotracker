package state

type Memory struct {
	portaToNote   uint8
	vibrato       uint8
	sampleOffset  uint8
	tempoDecrease uint8
	tempoIncrease uint8
}

func (m *Memory) getEffectMemory(input uint8, reg *uint8) uint8 {
	if input == 0 {
		return *reg
	}
	if input != 0 {
		*reg = input
	}
	return input
}

func (m *Memory) PortaToNote(input uint8) uint8 {
	return m.getEffectMemory(input, &m.portaToNote)
}

func (m *Memory) Vibrato(input uint8) uint8 {
	return m.getEffectMemory(input, &m.vibrato)
}

func (m *Memory) SampleOffset(input uint8) uint8 {
	return m.getEffectMemory(input, &m.sampleOffset)
}

func (m *Memory) TempoDecrease(input uint8) uint8 {
	return m.getEffectMemory(input, &m.tempoDecrease)
}

func (m *Memory) TempoIncrease(input uint8) uint8 {
	return m.getEffectMemory(input, &m.tempoIncrease)
}
