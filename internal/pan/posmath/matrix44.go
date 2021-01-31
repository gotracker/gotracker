package posmath

var (
	// IdentityMatrix44 is the identity 4x4 matrix
	IdentityMatrix44 = Matrix44{
		LeftVector4,
		UpVector4,
		ForwardVector4,
		WVector4,
	}
)

// Matrix44 is a 4x4 matrix
type Matrix44 [4]Vector4

// Determinant returns the determinant calculated using decomposition
// NOTE: the determinant will be zero if any of these is true:
//  * any row is zero
//  * any two rows or columns are equal
//  * any row or column is a constant multiple of another row or column
func (m Matrix44) Determinant() float64 {
	temp1 := m[2][2]*m[3][3] - m[2][3]*m[3][2]
	temp2 := m[2][1]*m[3][3] - m[2][3]*m[3][1]
	temp3 := m[2][1]*m[3][2] - m[2][2]*m[3][1]
	temp4 := m[2][0]*m[3][3] - m[2][3]*m[3][0]
	temp5 := m[2][0]*m[3][2] - m[2][2]*m[3][0]
	temp6 := m[2][0]*m[3][1] - m[2][1]*m[3][0]

	temp7 := m[1][1]*temp1 - m[1][2]*temp2
	temp8 := m[1][0]*temp1 - m[1][2]*temp4
	temp9 := m[1][0]*temp2 - m[1][1]*temp4
	temp10 := m[1][0]*temp3 - m[1][1]*temp5
	temp11 := m[1][3] * temp3
	temp12 := m[1][3] * temp5
	temp13 := m[1][3] * temp6
	temp14 := m[1][2] * temp6
	temp15 := m[0][0] * (temp7 + temp11)
	temp16 := m[0][1] * (temp8 + temp12)
	temp17 := m[0][2] * (temp9 + temp13)
	temp18 := m[0][3] * (temp10 + temp14)
	return temp15 - temp16 + temp17 - temp18
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
