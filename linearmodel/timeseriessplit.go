package linearmodel

import (
	"errors"
	"fmt"
)

const defaultNSplits = 5

var (
	ErrUninitializedTimeSeriesSplit = errors.New("uninitialized time series split")
	ErrInvalidNSplits               = errors.New("number of splits must be at least 2")
	ErrSplitsFewerThanSamples       = errors.New("cannot have splits fewer than number of samples")
	ErrTooManySplits                = errors.New("too many splits for number of samples")
)

type TimeSeriesSplit struct {
	// Number of splits
	NSplits int

	// Max size of each train set
	MaxTrainSize int

	// Limit size of the test set
	TestSize int

	// Number of samples to exclude from the end of each train set before the test set
	Gap int
}

func NewDefaultTimeSeriesSplit() *TimeSeriesSplit {
	return &TimeSeriesSplit{
		NSplits: defaultNSplits,
	}
}

func (t *TimeSeriesSplit) Split(ts []float64) ([]SplitGroup, error) {
	var groups []SplitGroup
	if t == nil {
		return groups, ErrUninitializedTimeSeriesSplit
	}

	if t.NSplits < 2 {
		return groups, fmt.Errorf("with splits, %d, %w", t.NSplits, ErrInvalidNSplits)
	}

	if len(ts) < t.NSplits {
		return groups, fmt.Errorf("with %d samples and %d splits, %w", len(ts), t.NSplits, ErrSplitsFewerThanSamples)
	}

	testSize := t.TestSize
	if testSize <= 0 {
		testSize = max(len(ts)/(t.NSplits+1), 1)
	}

	if len(ts)-t.NSplits*testSize-t.Gap <= 0 {
		return groups, fmt.Errorf("with %d samples and %d test size and %d gap, %w", len(ts), testSize, t.Gap, ErrTooManySplits)
	}

	groups = make([]SplitGroup, 0, t.NSplits)
	for i := t.NSplits - 1; i >= 0; i-- {
		var group SplitGroup
		split := make([]int, 0, len(ts)-i*testSize)
		for j := 0; j < cap(split); j++ {
			split = append(split, j)
		}
		fmt.Println(split)
		group.Train = split[:len(split)-testSize-t.Gap]
		group.Test = split[len(split)-testSize:]
		groups = append(groups, group)
	}

	return groups, nil
}

type SplitGroup struct {
	Train []int
	Test  []int
}
