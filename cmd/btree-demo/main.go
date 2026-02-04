package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/wato787/b-tree/btree"
)

func main() {
	var (
		t        = flag.Int("t", 2, "Bツリーの最小次数 t（>=2）")
		ops      = flag.String("ops", "", "操作列。例: \"ins 10,ins 20,del 10,find 20\"")
		printAll = flag.Bool("print", false, "最後に木と昇順キー列を表示する")
		step     = flag.Bool("step", false, "各操作後に木を表示する（学習用）")
		validate = flag.Bool("validate", false, "各操作後に Validate() する")
	)
	flag.Parse()

	tr, err := btree.New(*t)
	if err != nil {
		log.Fatal(err)
	}

	if strings.TrimSpace(*ops) == "" {
		fmt.Println("ops が空です。例: -ops \"ins 10,ins 20,del 10,find 20\"")
		return
	}

	for _, raw := range splitOps(*ops) {
		kind, arg, ok := parseOp(raw)
		if !ok {
			log.Fatalf("操作を解釈できません: %q（例: ins 10）", raw)
		}

		switch kind {
		case "ins", "insert":
			tr.Insert(arg)
		case "del", "delete", "rm":
			tr.Delete(arg)
		case "find", "search", "has":
			fmt.Printf("find %d => %v\n", arg, tr.Search(arg))
		default:
			log.Fatalf("未知の操作: %q（ins/del/find）", kind)
		}

		if *validate {
			if err := tr.Validate(); err != nil {
				log.Fatalf("Validate 失敗（op=%q）: %v\n木:\n%s", raw, err, tr.String())
			}
		}
		if *step {
			fmt.Printf("op: %s\n%s\n\n", raw, tr.String())
		}
	}

	if *printAll {
		fmt.Println("tree:")
		fmt.Println(tr.String())
		fmt.Println()
		fmt.Println("keys:")
		fmt.Println(tr.Keys())
	}
}

func splitOps(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

func parseOp(raw string) (kind string, arg int, ok bool) {
	fs := strings.Fields(raw)
	if len(fs) != 2 {
		return "", 0, false
	}
	v, err := strconv.Atoi(fs[1])
	if err != nil {
		return "", 0, false
	}
	return strings.ToLower(fs[0]), v, true
}
