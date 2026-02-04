package btree

import (
	"fmt"
	"strings"
)

// BTree は最小次数 t (minimum degree) を持つBツリーです。
//
// - 1ノードあたりのキー数: [t-1, 2t-1]（rootを除く）
// - 子の数: keys+1（内部ノードの場合）
// - すべての葉は同じ深さ
//
// t は 2 以上が必要です（t=2 は 2-3-4木相当）。
type BTree struct {
	t    int
	root *node
}

// New は最小次数 t のBツリーを作成します。
func New(t int) (*BTree, error) {
	if t < 2 {
		return nil, fmt.Errorf("btree: minimum degree t must be >= 2 (got %d)", t)
	}
	return &BTree{t: t, root: newLeaf()}, nil
}

// MinDegree は最小次数 t を返します。
func (tr *BTree) MinDegree() int { return tr.t }

// Search はキー k を探索し、見つかれば true を返します。
func (tr *BTree) Search(k int) bool {
	if tr.root == nil {
		return false
	}
	_, found := tr.root.search(k)
	return found
}

// Insert はキー k を挿入します。すでに存在する場合は何もしません（重複なし）。
func (tr *BTree) Insert(k int) {
	if tr.root == nil {
		tr.root = newLeaf()
	}
	if tr.Search(k) {
		return
	}

	// root が満杯なら高さを1増やして分割してから挿入する。
	if len(tr.root.keys) == 2*tr.t-1 {
		oldRoot := tr.root
		newRoot := newInternal()
		newRoot.children = []*node{oldRoot}
		newRoot.splitChild(0, tr.t)
		tr.root = newRoot
	}
	tr.root.insertNonFull(k, tr.t)
}

// Delete はキー k を削除します。存在しない場合は何もしません。
func (tr *BTree) Delete(k int) {
	if tr.root == nil {
		return
	}
	tr.root.delete(k, tr.t)

	// root が空になったら高さを下げる。
	if len(tr.root.keys) == 0 && !tr.root.leaf {
		tr.root = tr.root.children[0]
	}
	// 全削除時は葉の空ノードを維持
	if tr.root == nil {
		tr.root = newLeaf()
	}
}

// Keys は全キーを昇順で返します。
func (tr *BTree) Keys() []int {
	if tr.root == nil {
		return nil
	}
	var out []int
	tr.root.inorder(&out)
	return out
}

// String は学習用の簡易表示です（各レベルごとにノードのキー列を表示）。
func (tr *BTree) String() string {
	if tr.root == nil {
		return "<nil>"
	}
	var b strings.Builder
	type item struct {
		n     *node
		level int
	}
	q := []item{{tr.root, 0}}
	cur := 0
	firstInLevel := true
	for len(q) > 0 {
		it := q[0]
		q = q[1:]
		if it.level != cur {
			b.WriteByte('\n')
			cur = it.level
			firstInLevel = true
		} else if !firstInLevel {
			b.WriteString("  ")
		}
		b.WriteString(it.n.keysString())
		firstInLevel = false
		if !it.n.leaf {
			for _, ch := range it.n.children {
				q = append(q, item{ch, it.level + 1})
			}
		}
	}
	return b.String()
}

// Validate はBツリーの不変条件を検査します。学習用に詳細なエラーを返します。
func (tr *BTree) Validate() error {
	if tr.root == nil {
		return fmt.Errorf("btree: root is nil")
	}
	if tr.t < 2 {
		return fmt.Errorf("btree: invalid minimum degree t=%d", tr.t)
	}
	return tr.root.validate(tr.t, true)
}
