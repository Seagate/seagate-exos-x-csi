package controller

import storageapi "github.com/Seagate/seagate-exos-x-api-go"

// Volumes : convenience alias fort sorting purposes
type Volumes []storageapi.Volume

func (v Volumes) Len() int {
	return len(v)
}

func (v Volumes) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v Volumes) Less(i, j int) bool {
	return v[i].LUN < v[j].LUN
}
