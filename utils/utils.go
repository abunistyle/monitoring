package utils

func ArrayContains[T comparable](s []T, e T) bool {
    for _, v := range s {
        if v == e {
            return true
        }
    }
    return false
}

func MapCopy[T comparable](m map[string]T) map[string]T {
    m2 := make(map[string]T, len(m))
    var id string
    for id, m2[id] = range m {
    }
    return m2
}
