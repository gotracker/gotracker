package posmath

import "math"

var (
	// ZeroVector2 is a 2-dimensional vector with components equal to 0
	ZeroVector2 = Vector2{0, 0}
	// OneVector2 is a 2-dimensional vector with components equal to 1
	OneVector2 = Vector2{1, 1}
	// RightVector2 is a 2-dimensional vector that is 1 in the X direction and 0 in other directions
	RightVector2 = Vector2{1, 0}
	// UpVector2 is a 2-dimensional vector that is 1 in the Y direction and 0 in other directions
	UpVector2 = Vector2{0, 1}
)

// Vector2 is a 2-dimensional vector
type Vector2 [2]float64

// Length returns the length of the vector
func (v Vector2) Length() float64 {
	return math.Sqrt(v.LengthSquared())
}

// LengthSquared returns the squared length of the vector
func (v Vector2) LengthSquared() float64 {
	return v[0]*v[0] + v[1]*v[1]
}

// Dot returns the dot product of the vector and the rhs vector
func (v Vector2) Dot(rhs Vector2) float64 {
	return v.Mul(rhs).Length()
}

// Normalize returns the normalized version of the vector
// will panic if the vector's length is 0
func (v Vector2) Normalize() Vector2 {
	length := v.Length()
	if length == 0 {
		panic("cannot normalize zero-length vector")
	}
	invLen := 1 / length
	return v.UniformMul(invLen)
}

// SafeNormalize returns the normalized version of the vector
// will return a zero-length vector if the starting vector length is 0
func (v Vector2) SafeNormalize() Vector2 {
	length := v.Length()
	if length == 0 {
		return Vector2{}
	}
	invLen := 1 / length
	return v.UniformMul(invLen)
}

// ToArray returns an array of the components of the vector
// []float64{X,Y,Z}
func (v Vector2) ToArray() []float64 {
	return v[:]
}

// Add combines two vectors together via addition
func (v Vector2) Add(rhs Vector2) Vector2 {
	return Vector2{
		v[0] + rhs[0],
		v[1] + rhs[1],
	}
}

// Sub combines two vectors together via subtraction
func (v Vector2) Sub(rhs Vector2) Vector2 {
	return Vector2{
		v[0] - rhs[0],
		v[1] - rhs[1],
	}
}

// Neg negates the vector components
func (v Vector2) Neg() Vector2 {
	return Vector2{
		-v[0],
		-v[1],
	}
}

// UniformMul returns the vector after a uniform scaling is applied to
// all components
func (v Vector2) UniformMul(scale float64) Vector2 {
	return Vector2{
		v[0] * scale,
		v[1] * scale,
	}
}

// Mul multiplies two vectors together
func (v Vector2) Mul(rhs Vector2) Vector2 {
	return Vector2{
		v[0] * rhs[0],
		v[1] * rhs[1],
	}
}

// Div divies the vector by the rhs vector and returns the result
func (v Vector2) Div(rhs Vector2) Vector2 {
	return Vector2{
		v[0] / rhs[0],
		v[1] / rhs[1],
	}
}

// DistanceSquared returns the squared distance from v to rhs
func (v Vector2) DistanceSquared(rhs Vector2) float64 {
	return v.Sub(rhs).LengthSquared()
}

// Distance returns the distance from v to rhs
func (v Vector2) Distance(rhs Vector2) float64 {
	return math.Sqrt(v.DistanceSquared(rhs))
}

// LerpVector2 returns the linear interpolation of two vectors
func LerpVector2(t float64, lhs, rhs Vector2) Vector2 {
	return Vector2{
		Lerp(t, lhs[0], rhs[0]),
		Lerp(t, lhs[1], rhs[1]),
	}
}

// Cross returns the cross-product of two vectors
func (v Vector2) Cross(rhs Vector2) float64 {
	return v[0]*rhs[1] - v[1]*rhs[0]
}

// Equals returns true if the components of the vector match those in rhs
func (v Vector2) Equals(rhs Vector2) bool {
	return v[0] == rhs[0] && v[1] == rhs[1]
}
