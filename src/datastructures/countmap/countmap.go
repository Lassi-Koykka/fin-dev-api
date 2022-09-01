package countmap

import (
	"sort"
)

type entry[T countMapType] struct {
	Key   string
	Value T
}

type countMapType interface {
	int | int8 | int16 | int32 | int64
}
type CountMap[T countMapType] map[string]T

func New[T countMapType]() *CountMap[T] {
	m := make(CountMap[T])
	return &m
}

func (countMap *CountMap[T]) Inc(key string) {
	m := *countMap
	_, ok := m[key]
	if ok {
		m[key]++
		return
	}
	m[key] = 1
}

func (countMap *CountMap[T]) IncAll(keys []string) {
	m := *countMap
	for _, key := range keys {
		_, ok := m[key]
		if ok {
			m[key]++
			continue
		}
		m[key] = 1
	}
}

func (countMap *CountMap[T]) Merge(anotherMap *CountMap[T]) {
	m := *countMap
	for key, val := range *anotherMap {
		_, ok := m[key]
		if ok {
			m[key] += val
			continue
		}
		m[key] = val
	}
}
func (countMap *CountMap[T]) Dec(key string) {
	m := *countMap
	_, ok := m[key]
	if ok {
		m[key]--
		return
	}
	m[key] = -1
}

func (countMap *CountMap[T]) SortedKeys() []string {
	m := *countMap
	keys := make([]string, 0, len(m))

	for key := range m {
		keys = append(keys, key)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return m[keys[i]] < m[keys[j]]
	})

	return keys
}

func (countMap *CountMap[T]) SortAsc() []entry[T] {
	m := *countMap
	sorted := []entry[T]{}
	
	for _, key := range countMap.SortedKeys() {
		val, ok := m[key]
		if(ok) {
			sorted = append(sorted, entry[T]{Key: key, Value: val})
		}
	}
	return sorted
}

func (countMap *CountMap[T]) SortDec() []entry[T] {
	sorted := countMap.SortAsc()
	for i, j := 0, len(sorted)-1; i < j; i, j = i+1, j-1 {
		sorted[i], sorted[j] = sorted[j], sorted[i]
	}
	return sorted
}
