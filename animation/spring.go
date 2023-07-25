// SPDX-License-Identifier: Unlicensed OR MIT
// Reference: https://medium.com/hackernoon/the-spring-factory-4c3d988e7129

package animation

import (
	"math"
)

type Easing struct {
	Damping         float64
	HalfCycles      int
	InitialPosition float64
	InitialVelocity int
}

type ZetaK struct {
	Zeta float64
	K    int
	Y0   float64
	V0   int
}

type OmegaB struct {
	Omega float64 `json:"omega"`
	B     float64 `json:"B"`
}

const (
	NEG = true
	POS = false
)

func SpringFactory(args Easing) func(float64) float64 {
	var (
		zeta     = args.Damping
		k        = args.HalfCycles
		y0       = args.InitialPosition
		v0       = args.InitialVelocity
		A        = y0
		B, omega float64
	)

	// If v0 is 0, an analytical solution exists, otherwise,
	// we need to numerically solve it.
	if math.Abs(float64(v0)) < 1e-6 {
		B = zeta * y0 / math.Sqrt(1-zeta*zeta)
		omega = computeOmega(A, B, float64(k), zeta)
	} else {
		result := numericallySolveOmegaAndB(ZetaK{
			Zeta: zeta,
			K:    k,
			Y0:   y0,
			V0:   v0,
		})

		B = result.B
		omega = result.Omega
	}

	omega *= 2 * math.Pi
	omegaD := omega * math.Sqrt(1-zeta*zeta)

	return func(t float64) float64 {
		var sinusoid = A*math.Cos(omegaD*t) + B*math.Sin(omegaD*t)
		return math.Exp(-t*zeta*omega) * sinusoid
	}
}

func Clamp(x, min, max float64) float64 {
	return math.Min(math.Max(x, min), max)
}

func computeOmega(A, B, k, zeta float64) float64 {
	// Haven't quite figured out why yet, but to ensure same behavior of
	// k when argument of arctangent is negative, need to subtract pi
	// otherwise an extra halfcycle occurs.
	//
	// It has something to do with -atan(-x) = atan(x),
	// the range of atan being (-pi/2, pi/2) which is a difference of pi
	//
	// The other way to look at it is that for every integer k there is a
	// solution and the 0 point for k is arbitrary, we just want it to be
	// equal to the thing that gives us the same number of halfcycles as k.
	if A*B < 0 && k >= 1 {
		k--
	}

	return (-math.Atan(A/B) + math.Pi*k) / (2 * math.Pi * math.Sqrt(1-zeta*zeta))
}

// Resolve recursive definition of omega an B using bisection method
func numericallySolveOmegaAndB(args ZetaK) OmegaB {
	zeta := args.Zeta
	k := args.K
	y0 := args.Y0
	v0 := args.V0

	// See https://en.wikipedia.org/wiki/Damping#Under-damping_.280_.E2.89.A4_.CE.B6_.3C_1.29
	// B and omega are recursively defined in solution. Know omega in terms of B, will numerically
	// solve for B.

	errorFn := func(B, omega float64) float64 {
		omegaD := omega * math.Sqrt(1-zeta*zeta)
		return B - ((zeta*omega*y0)+float64(v0))/omegaD
	}

	A := y0
	B := zeta // initial guess that's pretty close

	var omega, error_ float64
	var direction bool

	step := func() {
		omega = computeOmega(A, B, float64(k), zeta)
		error_ = errorFn(B, omega)
		direction = !math.Signbit(error_)
	}

	step()

	var lower, upper float64
	tolerance := 1e-6
	ct := 0.0
	maxct := 1e3

	if direction == POS {
		for direction == POS {
			ct++
			if ct > maxct {
				break
			}

			lower = B
			B *= 2
			step()
		}
		upper = B
	} else {
		upper = B
		B *= -1

		for direction == NEG {
			ct++
			if ct > maxct {
				break
			}

			lower = B
			B *= 2
			step()
		}
		lower = B
	}

	for math.Abs(error_) > tolerance {
		ct++
		if ct > maxct {
			break
		}

		B = (upper + lower) / 2
		step()

		if direction == POS {
			lower = B
		} else {
			upper = B
		}
	}

	return OmegaB{Omega: omega, B: B}
}
