package set

type setType interface {
    int | int8 | int16 | int32 | int64 | string
}

var exists = struct{}{}
type Set[T setType] struct {
    d map[T]struct{}
}

func New[T setType]() *Set[T] {
    s := &Set[T]{}
    s.d = make(map[T]struct{})
    return s
}

func (set *Set[T]) Add(val T) {
    set.d[val] = exists
} 

func (set *Set[T]) Includes(val T) bool {
    _, ok := set.d[val]
    return ok
} 

func (set *Set[T]) Remove(val T) {
    delete(set.d, val)
} 

func (set *Set[T]) Value() []T {
    keys := []T{}
    for key := range set.d {
        keys = append(keys, key)
    }
    return keys
} 
