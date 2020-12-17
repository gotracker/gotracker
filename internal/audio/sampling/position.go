package sampling

// Pos stores the integer and fractional portions of a position within a sample
type Pos struct {
	Pos  int
	Frac float32
}

// Add increments the internal position values by the specified amount
func (p *Pos) Add(amt float32) {
	f := p.Frac + amt
	i := int(f)
	p.Frac = f - float32(i)
	p.Pos += i
}
