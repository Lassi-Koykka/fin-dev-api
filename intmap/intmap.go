package intmap

import (
    "sort"
)

type intMapType interface {
    int | int8 | int16 | int32 | int64
}
type IntMap[T intMapType] map[string]T

func New[T intMapType]() *IntMap[T] {
    m := make(IntMap[T])
    return &m
}

func (intMap *IntMap[T]) Inc(key string) {
    m := *intMap
    _, ok := m[key]
    if ok {
        m[key]++
        return
    }
    m[key] = 1
}

func (intMap *IntMap[T]) Dec(key string) {
    m := *intMap
    _, ok := m[key]
    if ok {
        m[key]--
        return
    }
    m[key] = -1
}

func (intMap *IntMap[T]) SortedKeys() []string {
    m := *intMap
    keys := make([]string, 0, len(m))
  
    for key := range m {
        keys = append(keys, key)
    }
    sort.SliceStable(keys, func(i, j int) bool{
        return m[keys[i]] < m[keys[j]]
    })

    return keys
}
