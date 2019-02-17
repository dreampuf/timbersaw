package dashboard

import (
	"sync"
	"testing"
)

func uint64pointer(i uint64) *uint64 {
	return &i
}

func TestMergeMap(t *testing.T) {
	mapA := &sync.Map{}
	mapA.Store("a", uint64pointer(1))
	mapA.Store("b", uint64pointer(1))
	mapB := &sync.Map{}
	mapB.Store("b", uint64pointer(1))
	mapB.Store("c", uint64pointer(3))

	expectedMap := map[string]uint64{"a": 1, "b":2, "c":3}
	mapResult := MergeMap(mapA, mapB)
	//reflect.DeepEqual(mapResult, expectedMap)
	for rk, rv := range mapResult {
		for ek, ev := range expectedMap {
			if rk == ek {
				if rv != ev {
					t.Errorf("key '%s' doesn't meet expectation. original: '%d', expect: '%d'", rk, rv, ev)
				}
				break
			}
		}
	}
}

func TestURLExtraSection(t *testing.T) {
	for caseurl, caseresult := range map[string]string {
		"/app/main/posts": "/app",
		"/posts/abc": "/posts",
		"/wp-content": "/",
		"/apps/cart.jsp?appID=9768": "/apps",
		"/apps/cart.jsp?appID=7507": "/apps",
	} {
		if val := URLExtraSection(caseurl); val != caseresult {
			t.Errorf("'%s' and '%s' doesn't match", val, caseresult)
		}
	}
}
