package set

var exists = struct{}{}
type StrSet map[string]struct{}

func NewStrSet() StrSet {
    return make(map[string]struct{})
}

func (setPtr *StrSet) Add(val string) {
    set := *setPtr
    set[val] = exists
} 

func (setPtr *StrSet) Includes(val string) bool {
    set := *setPtr
    _, ok := set[val]
    return ok
} 

func (setPtr *StrSet) Remove(val string) {
    set := *setPtr
    delete(set, val)
} 

func (setPtr *StrSet) ToSlice() []string {
    set := *setPtr
    keys := []string{}
    for key := range set {
        keys = append(keys, key)
    }
    return keys
} 
