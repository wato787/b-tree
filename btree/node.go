package btree

import (
	"fmt"
	"sort"
	"strings"
)

type node struct {
	leaf     bool
	keys     []int
	children []*node // leaf=false のとき len(children)==len(keys)+1
}

func newLeaf() *node     { return &node{leaf: true} }
func newInternal() *node { return &node{leaf: false} }

func (n *node) keysString() string {
	if len(n.keys) == 0 {
		return "[]"
	}
	var b strings.Builder
	b.WriteByte('[')
	for i, k := range n.keys {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "%d", k)
	}
	b.WriteByte(']')
	return b.String()
}

// search は n 内のキー配列で二分探索し、(index, found) を返します。
// index は「k を挿入するならこの位置」という意味でも使えます。
func (n *node) search(k int) (int, bool) {
	i := sort.SearchInts(n.keys, k)
	if i < len(n.keys) && n.keys[i] == k {
		return i, true
	}
	if n.leaf {
		return i, false
	}
	return n.children[i].search(k)
}

// splitChild は parent=n が children[i] を分割し、中央値キーを parent に昇格させます。
//
// 前提: children[i] は満杯（2t-1 keys）
func (n *node) splitChild(i int, t int) {
	y := n.children[i]
	z := &node{leaf: y.leaf}

	// y.keys = [0..2t-2]
	// median = y.keys[t-1]
	median := y.keys[t-1]

	// z に末尾 t-1 個を移す
	z.keys = append(z.keys, y.keys[t:]...)
	// y は先頭 t-1 個を残す
	y.keys = y.keys[:t-1]

	// 子も分割（内部ノードの場合）
	if !y.leaf {
		z.children = append(z.children, y.children[t:]...)
		y.children = y.children[:t]
	}

	// parent に child を1つ増やす（i+1 に z を挿入）
	n.children = append(n.children, nil)
	copy(n.children[i+2:], n.children[i+1:])
	n.children[i+1] = z

	// parent に median を挿入（i に）
	n.keys = append(n.keys, 0)
	copy(n.keys[i+1:], n.keys[i:])
	n.keys[i] = median
}

func (n *node) insertNonFull(k int, t int) {
	i := sort.SearchInts(n.keys, k)
	if n.leaf {
		n.keys = append(n.keys, 0)
		copy(n.keys[i+1:], n.keys[i:])
		n.keys[i] = k
		return
	}

	// 子が満杯なら先に分割して、降りる先を決め直す
	if len(n.children[i].keys) == 2*t-1 {
		n.splitChild(i, t)
		if k > n.keys[i] {
			i++
		}
	}
	n.children[i].insertNonFull(k, t)
}

func (n *node) inorder(out *[]int) {
	if n.leaf {
		*out = append(*out, n.keys...)
		return
	}
	for i, k := range n.keys {
		n.children[i].inorder(out)
		*out = append(*out, k)
	}
	n.children[len(n.children)-1].inorder(out)
}

func (n *node) delete(k int, t int) {
	idx := sort.SearchInts(n.keys, k)

	// ケース1: このノードに k がある
	if idx < len(n.keys) && n.keys[idx] == k {
		if n.leaf {
			// 1a: 葉ならそのまま削除
			n.keys = append(n.keys[:idx], n.keys[idx+1:]...)
			return
		}

		// 1b: 内部ノード
		if len(n.children[idx].keys) >= t {
			// predecessor で置換
			pred := n.children[idx].maxKey()
			n.keys[idx] = pred
			n.children[idx].delete(pred, t)
			return
		}
		if len(n.children[idx+1].keys) >= t {
			// successor で置換
			succ := n.children[idx+1].minKey()
			n.keys[idx] = succ
			n.children[idx+1].delete(succ, t)
			return
		}

		// 1c: 両隣が t-1 ならマージしてから削除
		n.mergeChildren(idx)
		n.children[idx].delete(k, t)
		return
	}

	// ケース2: このノードには無い（降りる）
	if n.leaf {
		return
	}

	// children[idx] に降りる前に、キー数を >= t に保つ
	if len(n.children[idx].keys) == t-1 {
		idx = n.ensureChildHasAtLeastTKeys(idx, t)
	}

	n.children[idx].delete(k, t)
}

func (n *node) minKey() int {
	cur := n
	for !cur.leaf {
		cur = cur.children[0]
	}
	return cur.keys[0]
}

func (n *node) maxKey() int {
	cur := n
	for !cur.leaf {
		cur = cur.children[len(cur.children)-1]
	}
	return cur.keys[len(cur.keys)-1]
}

// ensureChildHasAtLeastTKeys は child=children[idx] が t-1 keys のとき、
// 兄弟から借りるかマージして、降りる前に child のキー数を増やします。
// 戻り値は「次に降りるべき child index」です（マージにより idx が左にずれる場合がある）。
func (n *node) ensureChildHasAtLeastTKeys(idx int, t int) int {
	// 左兄弟から借りられる
	if idx > 0 && len(n.children[idx-1].keys) >= t {
		n.borrowFromPrev(idx)
		return idx
	}
	// 右兄弟から借りられる
	if idx < len(n.children)-1 && len(n.children[idx+1].keys) >= t {
		n.borrowFromNext(idx)
		return idx
	}
	// どちらも t-1 ならマージ
	if idx < len(n.children)-1 {
		n.mergeChildren(idx)
		return idx
	}
	// 末尾なら左とマージ（idx-1 に吸収）
	n.mergeChildren(idx - 1)
	return idx - 1
}

// mergeChildren は children[i], keys[i], children[i+1] を children[i] にまとめ、
// 親から keys[i] と children[i+1] を削除します。
func (n *node) mergeChildren(i int) {
	left := n.children[i]
	right := n.children[i+1]

	// left.keys + parentKey + right.keys
	left.keys = append(left.keys, n.keys[i])
	left.keys = append(left.keys, right.keys...)

	if !left.leaf {
		left.children = append(left.children, right.children...)
	}

	// 親のキーを削除
	n.keys = append(n.keys[:i], n.keys[i+1:]...)
	// 親の子（右）を削除
	n.children = append(n.children[:i+1], n.children[i+2:]...)
}

// borrowFromPrev は左兄弟から1キー借りて child[idx] を増やします。
func (n *node) borrowFromPrev(idx int) {
	child := n.children[idx]
	sib := n.children[idx-1]

	// child に parentKey を先頭へ、parentKey を sib の末尾キーで置換
	child.keys = append([]int{n.keys[idx-1]}, child.keys...)
	n.keys[idx-1] = sib.keys[len(sib.keys)-1]
	sib.keys = sib.keys[:len(sib.keys)-1]

	if !child.leaf {
		// sib の末尾子を child の先頭へ
		move := sib.children[len(sib.children)-1]
		sib.children = sib.children[:len(sib.children)-1]
		child.children = append([]*node{move}, child.children...)
	}
}

// borrowFromNext は右兄弟から1キー借りて child[idx] を増やします。
func (n *node) borrowFromNext(idx int) {
	child := n.children[idx]
	sib := n.children[idx+1]

	// child に parentKey を末尾へ、parentKey を sib の先頭キーで置換
	child.keys = append(child.keys, n.keys[idx])
	n.keys[idx] = sib.keys[0]
	sib.keys = sib.keys[1:]

	if !child.leaf {
		// sib の先頭子を child の末尾へ
		move := sib.children[0]
		sib.children = sib.children[1:]
		child.children = append(child.children, move)
	}
}

func (n *node) validate(t int, isRoot bool) error {
	leafDepth := -1
	return n.validateRec(t, isRoot, nil, nil, 0, &leafDepth)
}

func (n *node) validateRec(t int, isRoot bool, minExclusive *int, maxExclusive *int, depth int, leafDepth *int) error {
	if t < 2 {
		return fmt.Errorf("btree: invalid t=%d", t)
	}
	if n == nil {
		return fmt.Errorf("btree: nil node")
	}

	// keys 数の制約
	if isRoot {
		// root は 0..2t-1 を許容（空の木も表現したい）
		if len(n.keys) > 2*t-1 {
			return fmt.Errorf("btree: root has too many keys: %d", len(n.keys))
		}
	} else {
		if len(n.keys) < t-1 || len(n.keys) > 2*t-1 {
			return fmt.Errorf("btree: node has invalid key count: %d (must be in [%d,%d])", len(n.keys), t-1, 2*t-1)
		}
	}

	// ソート＆重複なし
	for i := 1; i < len(n.keys); i++ {
		if n.keys[i-1] >= n.keys[i] {
			return fmt.Errorf("btree: keys not strictly increasing: %v", n.keys)
		}
	}

	// 範囲制約
	if minExclusive != nil {
		for _, k := range n.keys {
			if k <= *minExclusive {
				return fmt.Errorf("btree: key %d violates minExclusive %d", k, *minExclusive)
			}
		}
	}
	if maxExclusive != nil {
		for _, k := range n.keys {
			if k >= *maxExclusive {
				return fmt.Errorf("btree: key %d violates maxExclusive %d", k, *maxExclusive)
			}
		}
	}

	// 子の整合性
	if n.leaf {
		if len(n.children) != 0 {
			return fmt.Errorf("btree: leaf has children: %d", len(n.children))
		}
		if *leafDepth == -1 {
			*leafDepth = depth
		} else if *leafDepth != depth {
			return fmt.Errorf("btree: leaves at different depths: got %d want %d", depth, *leafDepth)
		}
		return nil
	}

	if len(n.children) != len(n.keys)+1 {
		return fmt.Errorf("btree: internal node children mismatch: children=%d keys=%d", len(n.children), len(n.keys))
	}

	// 子ごとの範囲を組み立てて再帰
	for i := 0; i < len(n.children); i++ {
		var childMin *int
		var childMax *int

		// child i の範囲は:
		// - i==0: (minExclusive, keys[0])
		// - 0<i<len(keys): (keys[i-1], keys[i])
		// - i==len(keys): (keys[last], maxExclusive)
		if i == 0 {
			childMin = minExclusive
			if len(n.keys) > 0 {
				childMax = &n.keys[0]
			} else {
				childMax = maxExclusive
			}
		} else if i == len(n.keys) {
			childMin = &n.keys[i-1]
			childMax = maxExclusive
		} else {
			childMin = &n.keys[i-1]
			childMax = &n.keys[i]
		}

		if err := n.children[i].validateRec(t, false, childMin, childMax, depth+1, leafDepth); err != nil {
			return err
		}
	}
	return nil
}
