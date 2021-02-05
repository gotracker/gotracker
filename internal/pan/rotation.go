package pan

import (
	"math"

	"gotracker/internal/pan/posmath"
)

// RotationOrder defines the order of rotation of 3 Euler angles when generating a rotation matrix
// NOTE: really they're Tait-Bryan angles, but who cares that much?
type RotationOrder int

const (
	// RotationOrderXYZ intrinsically rotates Pitch, Yaw, Roll in that order
	RotationOrderXYZ = RotationOrder(iota)
	// RotationOrderXZY intrinsically rotates Pitch, Roll, Yaw in that order
	RotationOrderXZY
	// RotationOrderYXZ intrinsically rotates Yaw, Pitch, Roll in that order
	RotationOrderYXZ
	// RotationOrderYZX intrinsically rotates Yaw, Roll, Pitch in that order
	RotationOrderYZX
	// RotationOrderZYX intrinsically rotates Roll, Yaw, Pitch in that order
	RotationOrderZYX
	// RotationOrderZXY intrinsically rotates Roll, Pitch, Yaw in that order
	RotationOrderZXY
)

// Rotation is the details for a 3-axis rotation
type Rotation posmath.Vector3

// SetPitch sets the pitch (X-axis) rotation
func (r Rotation) SetPitch(rad float64) {
	r[0] = rad
}

// GetPitch returns the pitch (X-axis) rotation
func (r Rotation) GetPitch() float64 {
	return r[0]
}

// SetYaw sets the yaw (Y-axis) rotation
func (r Rotation) SetYaw(rad float64) {
	r[1] = rad
}

// GetYaw returns the yaw (Y-axis) rotation
func (r Rotation) GetYaw() float64 {
	return r[1]
}

// SetRoll sets the roll (Z-axis) rotation
func (r Rotation) SetRoll(rad float64) {
	r[2] = rad
}

// GetRoll returns the roll (Z-axis) rotation
func (r Rotation) GetRoll() float64 {
	return r[2]
}

// ToMatrix44 generates a rotation matrix based on the requested rotation order
func (r Rotation) ToMatrix44(order RotationOrder) posmath.Matrix44 {
	switch order {
	case RotationOrderXYZ:
		return r.ToMatrix44XYZ()
	case RotationOrderXZY:
		return r.ToMatrix44XZY()
	case RotationOrderYXZ:
		return r.ToMatrix44YXZ()
	case RotationOrderYZX:
		return r.ToMatrix44YZX()
	case RotationOrderZYX:
		return r.ToMatrix44ZYX()
	case RotationOrderZXY:
		return r.ToMatrix44ZXY()
	default:
		panic("unhandled rotation order")
	}
}

// ToMatrix44XYZ returns a 4x4 matrix based on a rotation order XYZ
func (r Rotation) ToMatrix44XYZ() posmath.Matrix44 {
	// Rz(A) * Ry(B) * Rx(C)
	cosA, sinA := math.Sincos(r.GetRoll())
	cosB, sinB := math.Sincos(r.GetYaw())
	cosC, sinC := math.Sincos(r.GetPitch())
	m00 := cosA * cosB
	m01 := cosA*sinB*sinC - sinA*cosC
	m02 := cosA*sinB*cosC + sinA*sinC
	m10 := sinA * cosB
	m11 := sinA*sinB*sinC + cosA*cosC
	m12 := sinA*sinB*cosC - cosA*sinC
	m20 := sinB
	m21 := cosB * sinC
	m22 := cosB * cosC
	return posmath.Matrix44{
		posmath.Vector4{m00, m01, m02, 0},
		posmath.Vector4{m10, m11, m12, 0},
		posmath.Vector4{m20, m21, m22, 0},
		posmath.WVector4,
	}
}

// ToMatrix44XZY returns a 4x4 matrix based on a rotation order XZY
func (r Rotation) ToMatrix44XZY() posmath.Matrix44 {
	// Ry(A) * Rz(B) * Rx(C)
	cosA, sinA := math.Sincos(r.GetYaw())
	cosB, sinB := math.Sincos(r.GetRoll())
	cosC, sinC := math.Sincos(r.GetPitch())
	m00 := cosA * cosB
	m01 := sinA*sinC - cosA*cosC*sinB
	m02 := cosC*sinA + cosA*sinB*sinC
	m10 := sinB
	m11 := cosB * cosC
	m12 := -cosB * sinC
	m20 := -cosB * sinA
	m21 := cosA*sinC + cosC*sinA*sinB
	m22 := cosA*cosC - sinA*sinB*sinC
	return posmath.Matrix44{
		posmath.Vector4{m00, m01, m02, 0},
		posmath.Vector4{m10, m11, m12, 0},
		posmath.Vector4{m20, m21, m22, 0},
		posmath.WVector4,
	}
}

// ToMatrix44YXZ returns a 4x4 matrix based on a rotation order YXZ
func (r Rotation) ToMatrix44YXZ() posmath.Matrix44 {
	// Rz(A) * Rx(B) * Ry(C)
	cosA, sinA := math.Sincos(r.GetRoll())
	cosB, sinB := math.Sincos(r.GetPitch())
	cosC, sinC := math.Sincos(r.GetYaw())
	m00 := cosA*cosC + sinA*sinB*sinC
	m01 := cosC*sinA*sinB - cosA*sinC
	m02 := cosB * sinA
	m10 := cosB * sinC
	m11 := cosB * cosC
	m12 := -sinB
	m20 := cosA*sinB*sinC - cosC*sinA
	m21 := cosA*cosC*sinB + sinA*sinC
	m22 := cosA * cosB
	return posmath.Matrix44{
		posmath.Vector4{m00, m01, m02, 0},
		posmath.Vector4{m10, m11, m12, 0},
		posmath.Vector4{m20, m21, m22, 0},
		posmath.WVector4,
	}
}

// ToMatrix44YZX returns a 4x4 matrix based on a rotation order YZX
func (r Rotation) ToMatrix44YZX() posmath.Matrix44 {
	// Rx(A) * Rz(B) * Ry(C)
	cosA, sinA := math.Sincos(r.GetPitch())
	cosB, sinB := math.Sincos(r.GetRoll())
	cosC, sinC := math.Sincos(r.GetYaw())
	m00 := cosB * cosC
	m01 := -sinB
	m02 := cosB * sinC
	m10 := sinA*sinC + cosA*cosC*sinB
	m11 := cosA * cosB
	m12 := cosA*sinB*sinC - cosC*sinA
	m20 := cosC*sinA*sinB - cosA*sinC
	m21 := cosB * sinA
	m22 := cosA*cosC + sinA*sinB*sinC
	return posmath.Matrix44{
		posmath.Vector4{m00, m01, m02, 0},
		posmath.Vector4{m10, m11, m12, 0},
		posmath.Vector4{m20, m21, m22, 0},
		posmath.WVector4,
	}
}

// ToMatrix44ZYX returns a 4x4 matrix based on a rotation order ZYX
func (r Rotation) ToMatrix44ZYX() posmath.Matrix44 {
	// Rx(A) * Ry(B) * Rz(C)
	cosA, sinA := math.Sincos(r.GetPitch())
	cosB, sinB := math.Sincos(r.GetYaw())
	cosC, sinC := math.Sincos(r.GetRoll())
	m00 := cosA*cosC - sinA*sinB*sinC
	m01 := -cosB * sinA
	m02 := cosA*sinC + cosC*sinA*sinB
	m10 := cosC*sinA + cosA*sinB*sinC
	m11 := cosA * cosB
	m12 := sinA*sinC - cosA*cosC*sinB
	m20 := -cosB * sinC
	m21 := sinB
	m22 := cosB * cosC
	return posmath.Matrix44{
		posmath.Vector4{m00, m01, m02, 0},
		posmath.Vector4{m10, m11, m12, 0},
		posmath.Vector4{m20, m21, m22, 0},
		posmath.WVector4,
	}
}

// ToMatrix44ZXY returns a 4x4 matrix based on a rotation order ZXY
func (r Rotation) ToMatrix44ZXY() posmath.Matrix44 {
	// Ry(A) * Rx(B) * Rz(C)
	cosA, sinA := math.Sincos(r.GetYaw())
	cosB, sinB := math.Sincos(r.GetPitch())
	cosC, sinC := math.Sincos(r.GetRoll())
	m00 := cosA*cosC + sinA*sinB*sinC
	m01 := cosC*sinA*sinB - cosA*sinC
	m02 := cosB * sinA
	m10 := cosB * sinC
	m11 := cosB * cosC
	m12 := -sinB
	m20 := cosA*sinB*sinC - cosC*sinA
	m21 := cosA*cosC*sinB + sinA*sinC
	m22 := cosA * cosB
	return posmath.Matrix44{
		posmath.Vector4{m00, m01, m02, 0},
		posmath.Vector4{m10, m11, m12, 0},
		posmath.Vector4{m20, m21, m22, 0},
		posmath.WVector4,
	}
}
