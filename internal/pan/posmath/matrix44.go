package posmath

var (
	// IdentityMatrix44 is the identity 4x4 matrix
	IdentityMatrix44 = Matrix44{
		RightVector4,
		UpVector4,
		ForwardVector4,
		WVector4,
	}
	// ZeroMatrix44 is a matrix with all its components set to 0
	ZeroMatrix44 = Matrix44{}
)

// Matrix44 is a 4x4 matrix
type Matrix44 [4]Vector4

// Determinant returns the determinant calculated using decomposition
// NOTE: the determinant will be zero if any of these is true:
//  * any row is zero
//  * any two rows or columns are equal
//  * any row or column is a constant multiple of another row or column
func (m Matrix44) Determinant() float64 {
	b0 := m[2][2]*m[3][3] - m[2][3]*m[3][2]
	b1 := m[2][1]*m[3][3] - m[2][3]*m[3][1]
	b2 := m[2][1]*m[3][2] - m[2][2]*m[3][1]
	b3 := m[2][0]*m[3][3] - m[2][3]*m[3][0]
	b4 := m[2][0]*m[3][2] - m[2][2]*m[3][0]
	b5 := m[2][0]*m[3][1] - m[2][1]*m[3][0]

	d00 := m[1][1]*b5 + m[1][2]*b4 + m[1][3]*b3
	d01 := m[1][0]*b5 + m[1][2]*b2 + m[1][3]*b1
	d02 := m[1][0]*-b4 + m[1][1]*b2 + m[1][3]*b0
	d03 := m[1][0]*b3 + m[1][1]*-b1 + m[1][2]*b0
	return m[0][0]*d00 - m[0][1]*d01 + m[0][2]*d02 - m[0][3]*d03
}

// Add returns the result of an addition of two matrices
func (m Matrix44) Add(rhs Matrix44) Matrix44 {
	return Matrix44{
		m[0].Add(rhs[0]),
		m[1].Add(rhs[1]),
		m[2].Add(rhs[2]),
		m[3].Add(rhs[3]),
	}
}

// Sub returns the result of a subtraction of two matrices
func (m Matrix44) Sub(rhs Matrix44) Matrix44 {
	return Matrix44{
		m[0].Sub(rhs[0]),
		m[1].Sub(rhs[1]),
		m[2].Sub(rhs[2]),
		m[3].Sub(rhs[3]),
	}
}

// Mul returns the result of a multiplication of two matrices
func (m Matrix44) Mul(rhs Matrix44) Matrix44 {
	return Matrix44{
		m[0].Transform(rhs),
		m[1].Transform(rhs),
		m[2].Transform(rhs),
		m[3].Transform(rhs),
	}
}

// UniformMul returns the result of a uniformly-scaled matrix
func (m Matrix44) UniformMul(scale float64) Matrix44 {
	return Matrix44{
		m[0].UniformMul(scale),
		m[1].UniformMul(scale),
		m[2].UniformMul(scale),
		m[3].UniformMul(scale),
	}
}

// Neg returns the negated matrix
func (m Matrix44) Neg() Matrix44 {
	return Matrix44{
		m[0].Neg(),
		m[1].Neg(),
		m[2].Neg(),
		m[3].Neg(),
	}
}

// T returns the transposed representation of the matrix
func (m Matrix44) T() Matrix44 {
	return Matrix44{
		Vector4{m[0][0], m[1][0], m[2][0], m[3][0]},
		Vector4{m[0][1], m[1][1], m[2][1], m[3][1]},
		Vector4{m[0][2], m[1][2], m[2][2], m[3][2]},
		Vector4{m[0][3], m[1][3], m[2][3], m[3][3]},
	}
}

// Invert returns the inverted matrix
// NOTE: if the determinant of the original matrix is 0, then the ZeroMatrix will be returned
func (m Matrix44) Invert() Matrix44 {
	b0 := m[2][2]*m[3][3] - m[2][3]*m[3][2]
	b1 := m[2][1]*m[3][3] - m[2][3]*m[3][1]
	b2 := m[2][1]*m[3][2] - m[2][2]*m[3][1]
	b3 := m[2][0]*m[3][3] - m[2][3]*m[3][0]
	b4 := m[2][0]*m[3][2] - m[2][2]*m[3][0]
	b5 := m[2][0]*m[3][1] - m[2][1]*m[3][0]

	d00 := m[1][1]*+b5 + m[1][2]*+b4 + m[1][3]*+b3
	d01 := m[1][0]*+b5 + m[1][2]*+b2 + m[1][3]*+b1
	d02 := m[1][0]*-b4 + m[1][1]*+b2 + m[1][3]*+b0
	d03 := m[1][0]*+b3 + m[1][1]*-b1 + m[1][2]*+b0

	d := m[0][0]*d00 - m[0][1]*d01 + m[0][2]*d02 - m[0][3]*d03
	if d == 0 {
		return ZeroMatrix44
	}

	det := 1 / d

	a0 := m[0][0]*m[1][1] - m[0][1]*m[1][0]
	a1 := m[0][0]*m[1][2] - m[0][2]*m[1][0]
	a2 := m[0][3]*m[1][0] - m[0][0]*m[1][3]
	a3 := m[0][1]*m[1][2] - m[0][2]*m[1][1]
	a4 := m[0][3]*m[1][1] - m[0][1]*m[1][3]
	a5 := m[0][2]*m[1][3] - m[0][3]*m[1][2]

	d10 := m[0][1]*+b5 + m[0][2]*+b4 + m[0][3]*+b3
	d11 := m[0][0]*+b5 + m[0][2]*+b2 + m[0][3]*+b1
	d12 := m[0][0]*-b4 + m[0][1]*+b2 + m[0][3]*+b0
	d13 := m[0][0]*+b3 + m[0][1]*-b1 + m[0][2]*+b0

	d20 := m[3][1]*+a5 + m[3][2]*+a4 + m[3][3]*+a3
	d21 := m[3][0]*+a5 + m[3][2]*+a2 + m[3][3]*+a1
	d22 := m[3][0]*-a4 + m[3][1]*+a2 + m[3][3]*+a0
	d23 := m[3][0]*+a3 + m[3][1]*-a1 + m[3][2]*+a0

	d30 := m[2][1]*+a5 + m[2][2]*+a4 + m[2][3]*+a3
	d31 := m[2][0]*+a5 + m[2][2]*+a2 + m[2][3]*+a1
	d32 := m[2][0]*-a4 + m[2][1]*+a2 + m[2][3]*+a0
	d33 := m[2][0]*+a3 + m[2][1]*-a1 + m[2][2]*+a0

	return Matrix44{
		Vector4{+d00, -d10, +d20, -d30},
		Vector4{-d01, +d11, -d21, +d31},
		Vector4{+d02, -d12, +d22, -d32},
		Vector4{-d03, +d13, -d23, +d33},
	}.UniformMul(det)
}

// LerpMatrix44 returns the linear interpolation of two matrices
func LerpMatrix44(t float64, lhs, rhs Matrix44) Matrix44 {
	return Matrix44{
		LerpVector4(t, lhs[0], rhs[0]),
		LerpVector4(t, lhs[1], rhs[1]),
		LerpVector4(t, lhs[2], rhs[2]),
		LerpVector4(t, lhs[3], rhs[3]),
	}
}
