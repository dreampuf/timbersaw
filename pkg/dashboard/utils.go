package dashboard

import (
	"regexp"
	"sort"
	"sync"
)

var (
	reSection = regexp.MustCompile("^(/[^/]+)/.*$")
)

// MergeMap combine multi *sync.Map to one map[string]uint64
func MergeMap(maps... *sync.Map) map[string]uint64 {
	mergeMap := map[string]uint64{}
	for _, i := range maps {
		i.Range(func(key, value interface{}) bool {
			skey := key.(string)
			if v, ok := mergeMap[skey]; ok {
				mergeMap[skey] = v + *value.(*uint64)
			} else {
				mergeMap[skey] = *value.(*uint64)
			}
			return true
		})
	}
	return mergeMap
}

// URLExtraSection turns a parent folder as the section
func URLExtraSection(url string) string {
	if matches := reSection.FindStringSubmatch(url); matches != nil {
		return matches[1]
	}
	return "/"
}

type pair struct {
	key string
	val uint64
}
type pairlist []pair

func (p pairlist) Len() int { return len(p) }
func (p pairlist) Less(i, j int) bool { return p[i].val > p[j].val }
func (p pairlist) Swap(i, j int){ p[i], p[j] = p[j], p[i] }

// SortedMapByValue converts a map to a pairlist with value's descending order
func SortedMapByValue(kv map[string]uint64) pairlist {
	pl := make(pairlist, len(kv))
	i := 0
	for k, v := range kv {
		pl[i] = pair{k, v}
		i ++
	}
	sort.Sort(pl)
	return pl
}

