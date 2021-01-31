package posmath

import "math"

var (
	// ZeroVector3 is a 3-dimensional vector with components equal to 0
	ZeroVector3 = Vector3{0, 0, 0}
	// OneVector3 is a 3-dimensional vector with components equal to 1
	OneVector3 = Vector3{1, 1, 1}
	// RightVector3 is a 3-dimensional vector that is 1 in the X direction and 0 in other directions
	RightVector3 = Vector3{1, 0, 0}
	// UpVector3 is a 3-dimensional vector that is 1 in the Y direction and 0 in other directions
	UpVector3 = Vector3{0, 1, 0}
	// ForwardVector3 is a 3-dimensional vector that is 1 in the Z direction and 0 in other directions
	ForwardVector3 = Vector3{0, 0, 1}
)

// Vector3 is a 3-dimensional vector
type Vector3 [3]float64

// Length returns the length of the vector
func (v Vector3) Length() float64 {
	return math.Sqrt(v.LengthSquared())
}

// LengthSquared returns the squared length of the vector
func (v Vector3) LengthSquared() float64 {
	return v[0]*v[0] + v[1]*v[1] + v[2]*v[2]
}

// Dot returns the dot product of the vector and the rhs vector
func (v Vector3) Dot(rhs Vector3) float64 {
	return v.Mul(rhs).Length()
}

// Normalize returns the normalized version of the vector
// will panic if the vector's length is 0
func (v Vector3) Normalize() Vector3 {
	length := v.Length()
	if length == 0 {
		panic("cannot normalize zero-length vector")
	}
	invLen := 1 / length
	return v.UniformMul(invLen)
}

// SafeNormalize returns the normalized version of the vector
// will return a zero-length vector if the starting vector length is 0
func (v Vector3) SafeNormalize() Vector3 {
	length := v.Length()
	if length == 0 {
		return Vector3{}
	}
	invLen := 1 / length
	return v.UniformMul(invLen)
}

// ToArray returns an array of the components of the vector
// []float64{X,Y,Z}
func (v Vector3) ToArray() []float64 {
	return v[:]
}

// Add combines two vectors together via addition
func (v Vector3) Add(rhs Vector3) Vector3 {
	return Vector3{
		v[0] + rhs[0],
		v[1] + rhs[1],
		v[2] + rhs[2],
	}
}

// Sub combines two vectors together via subtraction
func (v Vector3) Sub(rhs Vector3) Vector3 {
	return Vector3{
		v[0] - rhs[0],
		v[1] - rhs[1],
		v[2] - rhs[2],
	}
}

// Neg negates the vector components
func (v Vector3) Neg() Vector3 {
	return Vector3{
		-v[0],
		-v[1],
		-v[2],
	}
}

// UniformMul returns the vector after a uniform scaling is applied to
// all components
func (v Vector3) UniformMul(scale float64) Vector3 {
	return Vector3{
		v[0] * scale,
		v[1] * scale,
		v[2] * scale,
	}
}

// Mul multiplies two vectors together
func (v Vector3) Mul(rhs Vector3) Vector3 {
	return Vector3{
		v[0] * rhs[0],
		v[1] * rhs[1],
		v[2] * rhs[2],
	}
}

// Div divies the vector by the rhs vector and returns the result
func (v Vector3) Div(rhs Vector3) Vector3 {
	return Vector3{
		v[0] / rhs[0],
		v[1] / rhs[1],
		v[2] / rhs[2],
	}
}

// DistanceSquared returns the squared distance from v to rhs
func (v Vector3) DistanceSquared(rhs Vector3) float64 {
	return v.Sub(rhs).LengthSquared()
}

// Distance returns the distance from v to rhs
func (v Vector3) Distance(rhs Vector3) float64 {
	return math.Sqrt(v.DistanceSquared(rhs))
}

// LerpVector3 returns the linear interpolation of two vectors
func LerpVector3(t float64, lhs, rhs Vector3) Vector3 {
	return Vector3{
		Lerp(t, lhs[0], rhs[0]),
		Lerp(t, lhs[1], rhs[1]),
		Lerp(t, lhs[2], rhs[2]),
	}
}

// Cross returns the cross-product of two vectors
func (v Vector3) Cross(rhs Vector3) Vector3 {
	return Vector3{
		v[1]*rhs[2] - v[2]*rhs[1],
		v[2]*rhs[0] - v[0]*rhs[2],
		v[0]*rhs[1] - v[1]*rhs[0],
	}
}

// Equals returns true if the components of the vector match those in rhs
func (v Vector3) Equals(rhs Vector3) bool {
	return v[0] == rhs[0] && v[1] == rhs[1] && v[2] == rhs[2]
}
