package models

import (
	"errors"
	"fmt"
	"math"

	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/mat"
)

// OLS computes ordinary least squares using QR factorization
func OLS(obs, y mat.Matrix) (float64, []float64) {
	_, n := obs.Dims()
	qr := new(mat.QR)
	qr.Factorize(obs)

	q := new(mat.Dense)
	r := new(mat.Dense)

	qr.QTo(q)
	qr.RTo(r)
	yq := new(mat.Dense)
	yq.Mul(y, q)

	c := make([]float64, n)
	for i := n - 1; i >= 0; i-- {
		c[i] = yq.At(0, i)
		for j := i + 1; j < n; j++ {
			c[i] -= c[j] * r.At(i, j)
		}
		c[i] /= r.At(i, i)
	}
	if len(c) == 0 {
		return math.NaN(), nil
	}
	if len(c) == 1 {
		return c[0], nil
	}
	return c[0], c[1:]
}

var (
	ErrObsYSizeMismatch  = errors.New("observation and y matrix have different number of features")
	ErrWarmStartBetaSize = errors.New("warm start beta does not have the same dimensions as X")
)

type LassoOptions struct {
	WarmStartBeta []float64
	Lambda        float64
	Iterations    int
	Tolerance     float64
}

func NewDefaultLassoOptions() *LassoOptions {
	return &LassoOptions{
		Lambda:        1.0,
		Iterations:    1000,
		Tolerance:     1e-4,
		WarmStartBeta: nil,
	}
}

// LassoRegression computes the lasso regression using coordinate descent. lambda = 0 converges to OLS
func LassoRegression(obs, y *mat.Dense, opt *LassoOptions) (float64, []float64, error) {
	if opt == nil {
		opt = NewDefaultLassoOptions()
	}

	m, n := obs.Dims()

	_, ym := y.Dims()
	if m != ym {
		return 0, nil, fmt.Errorf("observation matrix has %d observations and y matrix as %d observations, %w", m, ym, ErrObsYSizeMismatch)
	}
	if opt.WarmStartBeta != nil && len(opt.WarmStartBeta) != n {
		return 0, nil, fmt.Errorf("warm start beta has %d features instead of %d, %w", len(opt.WarmStartBeta), n, ErrWarmStartBetaSize)
	}

	// tracks current betas
	beta := mat.NewDense(1, n, opt.WarmStartBeta)

	// precompute the per feature dot product
	xdot := make([]float64, n)
	for i := 0; i < n; i++ {
		xi := obs.ColView(i)
		xdot[i] = mat.Dot(xi, xi)
	}

	// tracks the per coordinate residual
	residual := mat.NewDense(1, m, nil)

	for i := 0; i < opt.Iterations; i++ {
		maxCoef := 0.0
		maxUpdate := 0.0

		// loop through all features and minimize loss function
		for j := 0; j < n; j++ {
			betaCurr := beta.At(0, j)
			if i != 0 {
				// early terminate if current beta has already been set to 0
				if betaCurr == 0 {
					continue
				}
			}

			residual.Mul(beta, obs.T())
			residual.Sub(y, residual)

			num := mat.Dot(obs.ColView(j), residual.RowView(0))
			betaNext := num/xdot[j] + betaCurr

			gamma := opt.Lambda / xdot[j]
			betaNext = SoftThreshold(betaNext, gamma)

			maxCoef = math.Max(maxCoef, betaNext)
			maxUpdate = math.Max(maxUpdate, math.Abs(betaNext-betaCurr))
			beta.Set(0, j, betaNext)
		}

		// break early if we've achieved the desired tolerance
		if maxUpdate < opt.Tolerance*maxCoef {
			break
		}
	}

	c := beta.RawRowView(0)
	if len(c) == 0 {
		return math.NaN(), nil, nil
	}
	if len(c) == 1 {
		return c[0], nil, nil
	}
	return c[0], c[1:], nil
}

// LassoRegression2 computes the lasso regression using coordinate descent. lambda = 0 converges to OLS
func LassoRegression2(obs, y *mat.Dense, opt *LassoOptions) (float64, []float64, error) {
	if opt == nil {
		opt = NewDefaultLassoOptions()
	}

	m, n := obs.Dims()

	_, ym := y.Dims()
	if m != ym {
		return 0, nil, fmt.Errorf("observation matrix has %d observations and y matrix as %d observations, %w", m, ym, ErrObsYSizeMismatch)
	}
	if opt.WarmStartBeta != nil && len(opt.WarmStartBeta) != n {
		return 0, nil, fmt.Errorf("warm start beta has %d features instead of %d, %w", len(opt.WarmStartBeta), n, ErrWarmStartBetaSize)
	}

	// tracks current betas
	beta := make([]float64, n)
	if opt.WarmStartBeta != nil {
		copy(beta, opt.WarmStartBeta)
	}

	// precompute the per feature dot product
	xdot := make([]float64, n)
	for i := 0; i < n; i++ {
		xi := obs.ColView(i)
		xdot[i] = mat.Dot(xi, xi)
	}

	// tracks the per coordinate residual
	residual := mat.NewDense(1, m, nil)
	betaX := mat.NewDense(1, m, nil)
	betaXDelta := mat.NewDense(1, m, nil)

	for i := 0; i < opt.Iterations; i++ {
		maxCoef := 0.0
		maxUpdate := 0.0
		betaDiff := 0.0
		// loop through all features and minimize loss function
		for j := 0; j < n; j++ {
			betaCurr := beta[j]
			if i != 0 {
				if betaCurr == 0 {
					continue
				}
			}

			betaX.Add(betaX, betaXDelta)
			residual.Sub(y, betaX)

			obsCol := obs.ColView(j)
			num := mat.Dot(obsCol, residual.RowView(0))
			betaNext := num/xdot[j] + betaCurr

			gamma := opt.Lambda / xdot[j]
			betaNext = SoftThreshold(betaNext, gamma)

			maxCoef = math.Max(maxCoef, betaNext)
			maxUpdate = math.Max(maxUpdate, math.Abs(betaNext-betaCurr))
			betaDiff = betaNext - betaCurr
			betaXDelta.Scale(betaDiff, obsCol.T())
			beta[j] = betaNext
		}

		// break early if we've achieved the desired tolerance
		if maxUpdate < opt.Tolerance*maxCoef {
			break
		}
	}

	if len(beta) == 0 {
		return math.NaN(), nil, nil
	}
	if len(beta) == 1 {
		return beta[0], nil, nil
	}
	return beta[0], beta[1:], nil
}

// LassoRegression3 computes the lasso regression using coordinate descent. lambda = 0 converges to OLS
func LassoRegression3(obs [][]float64, y []float64, opt *LassoOptions) (float64, []float64, error) {
	if opt == nil {
		opt = NewDefaultLassoOptions()
	}

	n := len(obs)
	if n == 0 {
		return 0, nil, errors.New("observation matrix has no observations")
	}
	m := len(obs[0])

	ym := len(y)
	if m != ym {
		return 0, nil, fmt.Errorf("observation matrix has %d observations and y matrix as %d observations, %w", m, ym, ErrObsYSizeMismatch)
	}
	if opt.WarmStartBeta != nil && len(opt.WarmStartBeta) != n {
		return 0, nil, fmt.Errorf("warm start beta has %d features instead of %d, %w", len(opt.WarmStartBeta), n, ErrWarmStartBetaSize)
	}

	// tracks current betas
	beta := make([]float64, n)
	if opt.WarmStartBeta != nil {
		copy(beta, opt.WarmStartBeta)
	}

	// precompute the per feature dot product
	xdot := make([]float64, n)
	for i := 0; i < n; i++ {
		xi := obs[i]
		xdot[i] = floats.Dot(xi, xi)
	}

	// tracks the per coordinate residual
	residual := make([]float64, m)
	betaX := make([]float64, m)
	betaXDelta := make([]float64, m)

	for i := 0; i < opt.Iterations; i++ {
		maxCoef := 0.0
		maxUpdate := 0.0
		betaDiff := 0.0
		// loop through all features and minimize loss function
		for j := 0; j < n; j++ {
			betaCurr := beta[j]
			if i != 0 {
				if betaCurr == 0 {
					continue
				}
			}

			floats.Add(betaX, betaXDelta)
			floats.SubTo(residual, y, betaX)

			obsCol := obs[j]
			num := floats.Dot(obsCol, residual)
			betaNext := num/xdot[j] + betaCurr

			gamma := opt.Lambda / xdot[j]
			betaNext = SoftThreshold(betaNext, gamma)

			maxCoef = math.Max(maxCoef, betaNext)
			maxUpdate = math.Max(maxUpdate, math.Abs(betaNext-betaCurr))
			betaDiff = betaNext - betaCurr
			floats.ScaleTo(betaXDelta, betaDiff, obsCol)
			beta[j] = betaNext
		}

		// break early if we've achieved the desired tolerance
		if maxUpdate < opt.Tolerance*maxCoef {
			break
		}
	}

	if len(beta) == 0 {
		return math.NaN(), nil, nil
	}
	if len(beta) == 1 {
		return beta[0], nil, nil
	}
	return beta[0], beta[1:], nil
}

func SoftThreshold(x, gamma float64) float64 {
	res := math.Max(0, math.Abs(x)-gamma)
	if math.Signbit(x) {
		return -res
	}
	return res
}
