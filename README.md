# Bツリー（B-Tree）学習リポジトリ（Go）

このリポジトリは **Bツリーを「読める・動かせる・壊して直せる」** ことを目標にした学習用教材です。
Goで、教科書的なアルゴリズム（CLRS系）に沿って **検索/挿入/削除** を実装し、さらに **不変条件の検証（Validate）** を用意しています。

## これは何？

### Bツリーの要点（まずここだけ）

- **1ノードが複数のキーを持つ**（配列のようにまとまっている）
- ノードが満杯になったら **分割（split）** して高さの増加を抑える
- 削除でスカスカになったら **借用（borrow）/併合（merge）** してバランスを保つ
- **すべての葉が同じ深さ** になるため、探索は常に \(O(\log n)\)

この実装では **最小次数 \(t\)**（minimum degree）を使います。

- **各（root以外の）ノードのキー数**: \(t-1 \le \#keys \le 2t-1\)
- **内部ノードの子の数**: \(\#children = \#keys + 1\)
- \(t \ge 2\)（\(t=2\) は 2-3-4木相当）

## 使い方

### テストを回す

```bash
go test ./...
```

### 触ってみる（簡易デモ）

```bash
go run ./cmd/btree-demo -t 2 -ops "ins 10,ins 20,ins 5,ins 6,ins 12,ins 30,ins 7,ins 17" -print
```

削除もできます。

```bash
go run ./cmd/btree-demo -t 3 -ops "ins 1,ins 2,ins 3,ins 4,ins 5,del 3,del 1" -print -validate
```

## ディレクトリ構成

- `btree/`: 学習用Bツリー実装（`Insert/Search/Delete/Validate`）
- `cmd/btree-demo/`: 手動で操作して挙動を観察するための簡易CLI
- `docs/`: 解説（図・用語・アルゴリズムの分岐を文章で追えるように）

## どこから読む？

- まずは `docs/01_overview.md`（用語と不変条件の全体像）
- 次に `docs/02_insert.md`（分割の気持ちよさ）
- その後 `docs/03_delete.md`（借用・併合の分岐が本番）
- 実コードは `btree/` を上から追う（`BTree.Insert` → `node.splitChild` → `node.insertNonFull`）

## 注意（学習用としての割り切り）

- **キー型は `int` 固定**（理解のため）
- **重複キーは許しません**（`Insert` は既存なら何もしない）
- 実務向けに必要な最適化（永続化・ページング・並行制御等）は扱いません

