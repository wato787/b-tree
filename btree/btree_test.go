package btree_test

import (
	"math/rand"
	"sort"
	"testing"

	"github.com/wato787/b-tree/btree"
)

func TestInsertSearchDelete_Small(t *testing.T) {
	tr, err := btree.New(2)
	if err != nil {
		t.Fatal(err)
	}

	ins := []int{10, 20, 5, 6, 12, 30, 7, 17}
	for _, k := range ins {
		tr.Insert(k)
		if err := tr.Validate(); err != nil {
			t.Fatalf("validate after insert %d: %v\n%s", k, err, tr.String())
		}
	}

	for _, k := range ins {
		if !tr.Search(k) {
			t.Fatalf("expected to find %d", k)
		}
	}
	if tr.Search(999) {
		t.Fatalf("did not expect to find 999")
	}

	// 重複は入らない
	tr.Insert(10)
	if err := tr.Validate(); err != nil {
		t.Fatalf("validate after duplicate insert: %v\n%s", err, tr.String())
	}

	// 削除
	dels := []int{6, 7, 10, 12, 30, 20, 17, 5}
	for _, k := range dels {
		tr.Delete(k)
		if err := tr.Validate(); err != nil {
			t.Fatalf("validate after delete %d: %v\n%s", k, err, tr.String())
		}
		if tr.Search(k) {
			t.Fatalf("expected %d to be deleted", k)
		}
	}

	if got := tr.Keys(); len(got) != 0 {
		t.Fatalf("expected empty keys, got %v", got)
	}
}

func TestRandomOps_ValidateAndOrder(t *testing.T) {
	for _, tc := range []struct {
		name string
		t    int
	}{
		{name: "t2", t: 2},
		{name: "t3", t: 3},
		{name: "t8", t: 8},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tr, err := btree.New(tc.t)
			if err != nil {
				t.Fatal(err)
			}

			r := rand.New(rand.NewSource(1))
			set := map[int]struct{}{}

			const (
				ops   = 5000
				space = 200
			)

			for i := 0; i < ops; i++ {
				k := r.Intn(space) - space/2
				switch r.Intn(3) {
				case 0, 1:
					tr.Insert(k)
					set[k] = struct{}{}
				case 2:
					tr.Delete(k)
					delete(set, k)
				}

				if err := tr.Validate(); err != nil {
					t.Fatalf("validate failed at i=%d: %v\n%s", i, err, tr.String())
				}

				// Keys() は昇順で、setと一致する
				want := sortedKeys(set)
				got := tr.Keys()
				if !equalInts(got, want) {
					t.Fatalf("keys mismatch at i=%d\n got=%v\nwant=%v\n%s", i, got, want, tr.String())
				}

				// Search も一致する（ランダムに数点チェック）
				for j := 0; j < 5; j++ {
					q := r.Intn(space) - space/2
					_, ok := set[q]
					if tr.Search(q) != ok {
						t.Fatalf("search mismatch q=%d want=%v", q, ok)
					}
				}
			}
		})
	}
}

func sortedKeys(set map[int]struct{}) []int {
	out := make([]int, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Ints(out)
	return out
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
