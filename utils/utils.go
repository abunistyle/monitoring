package utils

import (
    "crypto/md5"
    "encoding/hex"
    "fmt"
    "strings"
)

func ArrayContains[T comparable](s []T, e T) bool {
    for _, v := range s {
        if v == e {
            return true
        }
    }
    return false
}

func ArrayUnique[T comparable](s []T) []T {
    result := make([]T, 0, len(s))
    temp := map[T]struct{}{}
    for _, item := range s {
        if _, ok := temp[item]; !ok {
            temp[item] = struct{}{}
            result = append(result, item)
        }
    }
    return result
}

func MapCopy[T comparable](m map[string]T) map[string]T {
    m2 := make(map[string]T, len(m))
    var id string
    for id, m2[id] = range m {
    }
    return m2
}

func MD5(v string) string {
    d := []byte(v)
    m := md5.New()
    m.Write(d)
    return hex.EncodeToString(m.Sum(nil))
}

func GenDbIndexKey(moduleName string, groupName string) string {
    key := fmt.Sprintf("%s|%s", strings.ToLower(moduleName), strings.ToLower(groupName))
    return key
}
