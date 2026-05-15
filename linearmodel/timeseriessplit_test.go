package linearmodel

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeSeriesSplit(t *testing.T) {
	testData := map[string]struct {
		splits       int
		maxTrainSize int
		testSize     int
		gap          int

		ts          []float64
		expected    []SplitGroup
		expectedErr error
	}{
		"no time series with defaults": {
			expectedErr: ErrSplitsFewerThanSamples,
		},
		"invalid splits": {
			splits:      1,
			expectedErr: ErrInvalidNSplits,
		},
		"series less than splits": {
			ts:          []float64{1, 2, 3},
			expectedErr: ErrSplitsFewerThanSamples,
		},
		"incompatible gap with splits": {
			gap:         2,
			ts:          []float64{1, 2, 3, 4, 5, 6, 7},
			expectedErr: ErrTooManySplits,
		},
		"simple valid": {
			splits: 5,
			ts:     []float64{1, 2, 3, 4, 5, 6},
			expected: []SplitGroup{
				{
					Train: []int{0},
					Test:  []int{1},
				},
				{
					Train: []int{0, 1},
					Test:  []int{2},
				},
				{
					Train: []int{0, 1, 2},
					Test:  []int{3},
				},
				{
					Train: []int{0, 1, 2, 3},
					Test:  []int{4},
				},
				{
					Train: []int{0, 1, 2, 3, 4},
					Test:  []int{5},
				},
			},
		},
		"test size": {
			splits:   3,
			testSize: 2,
			ts:       []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
			expected: []SplitGroup{
				{
					Train: []int{0, 1, 2, 3, 4, 5},
					Test:  []int{6, 7},
				},
				{
					Train: []int{0, 1, 2, 3, 4, 5, 6, 7},
					Test:  []int{8, 9},
				},
				{
					Train: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
					Test:  []int{10, 11},
				},
			},
		},
		"test size with gap": {
			splits:   3,
			testSize: 2,
			gap:      2,
			ts:       []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
			expected: []SplitGroup{
				{
					Train: []int{0, 1, 2, 3},
					Test:  []int{6, 7},
				},
				{
					Train: []int{0, 1, 2, 3, 4, 5},
					Test:  []int{8, 9},
				},
				{
					Train: []int{0, 1, 2, 3, 4, 5, 6, 7},
					Test:  []int{10, 11},
				},
			},
		},
	}

	for name, td := range testData {
		t.Run(name, func(t *testing.T) {
			tsSplit := NewDefaultTimeSeriesSplit()
			if td.splits > 0 {
				tsSplit.NSplits = td.splits
			}
			tsSplit.MaxTrainSize = td.maxTrainSize
			tsSplit.TestSize = td.testSize
			tsSplit.Gap = td.gap

			res, err := tsSplit.Split(td.ts)
			if td.expectedErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, td.expectedErr)
				return
			}

			assert.Equal(t, td.expected, res)
		})
	}
}
