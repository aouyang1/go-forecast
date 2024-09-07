package forecast

import "github.com/aouyang1/go-forecast/feature"

type FeatureLabels struct {
	idx    map[feature.Feature]int
	labels []feature.Feature
}

func NewFeatureLabels(labels []feature.Feature) *FeatureLabels {
	idx := make(map[feature.Feature]int)
	for i := 0; i < len(labels); i++ {
		idx[labels[i]] = i
	}
	fl := &FeatureLabels{
		labels: labels,
		idx:    idx,
	}
	return fl
}

func (f *FeatureLabels) Len() int {
	return len(f.labels)
}

func (f *FeatureLabels) Labels() []feature.Feature {
	labels := make([]feature.Feature, len(f.labels))
	copy(labels, f.labels)
	return labels
}

func (f *FeatureLabels) Index(label feature.Feature) (int, bool) {
	if idx, exists := f.idx[label]; exists {
		return idx, exists
	}
	return -1, false
}
