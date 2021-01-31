package posmath

import "math"

var (
	// ZeroVector4 is a 4-dimensional vector with components equal to 0
	ZeroVector4 = Vector4{0, 0, 0, 0}
	// OneVector4 is a 4-dimensional vector with components equal to 1
	OneVector4 = Vector4{1, 1, 1, 1}
	// LeftVector4 is a 4-dimensional vector that is 1 in the X direction and 0 in other directions
	LeftVector4 = Vector4{1, 0, 0, 0}
	// UpVector4 is a 4-dimensional vector that is 1 in the Y direction and 0 in other directions
	UpVector4 = Vector4{0, 1, 0, 0}
	// ForwardVector4 is a 4-dimensional vector that is 1 in the Z direction and 0 in other directions
	ForwardVector4 = Vector4{0, 0, 1, 0}
	// WVector4 is a 4-dimensional vector that is 1 in the W direction and 0 in other directions
	WVector4 = Vector4{0, 0, 0, 1}
)

// Vector4 is a 4-dimensional vector with W component last
type Vector4 [4]float64

// SetVector3 returns a new vector with the X,Y,Z components from rhs
// and the W component from the current vector
func (v Vector4) SetVector3(rhs Vector3) Vector4 {
	return Vector4{
		rhs[0],
		rhs[1],
		rhs[2],
		v[3],
	}
}

// Vector3 returns a 3-dimensional vector with the X,Y,Z components of this vector
func (v Vector4) Vector3() Vector3 {
	return Vector3{
		v[0],
		v[1],
		v[2],
	}
}

// Length returns the length of the vector
func (v Vector4) Length() float64 {
	return math.Sqrt(v.LengthSquared())
}

// LengthSquared returns the squared length of the vector
func (v Vector4) LengthSquared() float64 {
	return v[0]*v[0] + v[1]*v[1] + v[2]*v[2] + v[3]*v[3]
}

// Dot returns the dot product of the vector and the rhs vector
func (v Vector4) Dot(rhs Vector4) float64 {
	return v.Mul(rhs).Length()
}

// Normalize returns the normalized version of the vector
// will panic if the vector's length is 0
func (v Vector4) Normalize() Vector4 {
	length := v.Length()
	if length == 0 {
		panic("cannot normalize zero-length vector")
	}
	invLen := 1 / length
	return v.UniformMul(invLen)
}

// SafeNormalize returns the normalized version of the vector
// will return a zero-length vector if the starting vector length is 0
func (v Vector4) SafeNormalize() Vector4 {
	length := v.Length()
	if length == 0 {
		return Vector4{}
	}
	invLen := 1 / length
	return v.UniformMul(invLen)
}

// ToArray returns an array of the components of the vector
// []float64{X,Y,Z,W}
func (v Vector4) ToArray() []float64 {
	return v[:]
}

// Add combines two vectors together via addition
func (v Vector4) Add(rhs Vector4) Vector4 {
	return Vector4{
		v[0] + rhs[0],
		v[1] + rhs[1],
		v[2] + rhs[2],
		v[3] + rhs[3],
	}
}

// Sub combines two vectors together via subtraction
func (v Vector4) Sub(rhs Vector4) Vector4 {
	return Vector4{
		v[0] - rhs[0],
		v[1] - rhs[1],
		v[2] - rhs[2],
		v[3] - rhs[3],
	}
}

// Neg negates the vector components
func (v Vector4) Neg() Vector4 {
	return Vector4{
		-v[0],
		-v[1],
		-v[2],
		-v[3],
	}
}

// UniformMul returns the vector after a uniform scaling is applied to
// all components
func (v Vector4) UniformMul(scale float64) Vector4 {
	return Vector4{
		v[0] * scale,
		v[1] * scale,
		v[2] * scale,
		v[3] * scale,
	}
}

// Mul multiplies two vectors together
func (v Vector4) Mul(rhs Vector4) Vector4 {
	return Vector4{
		v[0] * rhs[0],
		v[1] * rhs[1],
		v[2] * rhs[2],
		v[3] * rhs[3],
	}
}

// Div divies the vector by the rhs vector and returns the result
func (v Vector4) Div(rhs Vector4) Vector4 {
	return Vector4{
		v[0] / rhs[0],
		v[1] / rhs[1],
		v[2] / rhs[2],
		v[3] / rhs[3],
	}
}

// DistanceSquared returns the squared distance from v to rhs
func (v Vector4) DistanceSquared(rhs Vector4) float64 {
	return v.Sub(rhs).LengthSquared()
}

// Distance returns the distance from v to rhs
func (v Vector4) Distance(rhs Vector4) float64 {
	return math.Sqrt(v.DistanceSquared(rhs))
}

// LerpVector4 returns the linear interpolation of two vectors
func LerpVector4(t float64, lhs, rhs Vector4) Vector4 {
	return Vector4{
		Lerp(t, lhs[0], rhs[0]),
		Lerp(t, lhs[1], rhs[1]),
		Lerp(t, lhs[2], rhs[2]),
		Lerp(t, lhs[3], rhs[3]),
	}
}

// Transform returns a vector after transformation with the 4x4 matrix
func (v Vector4) Transform(m Matrix44) Vector4 {
	return Vector4{
		v[0]*m[0][0] + v[1]*m[1][0] + v[2]*m[2][0] + v[3]*m[3][0],
		v[0]*m[0][1] + v[1]*m[1][1] + v[2]*m[2][1] + v[3]*m[3][1],
		v[0]*m[0][2] + v[1]*m[1][2] + v[2]*m[2][2] + v[3]*m[3][2],
		v[0]*m[0][3] + v[1]*m[1][3] + v[2]*m[2][3] + v[3]*m[3][3],
	}
}

// Equals returns true if the components of the vector match those in rhs
func (v Vector4) Equals(rhs Vector4) bool {
	return v[0] == rhs[0] && v[1] == rhs[1] && v[2] == rhs[2] && v[3] == rhs[3]
}
