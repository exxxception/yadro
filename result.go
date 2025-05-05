package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
)

type kv struct {
	Key   string
	Value *ResultsTable
}

func SortByTime(m map[string]*ResultsTable) []string {
	ss := make([]kv, 0, len(m))
	for k, v := range m {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value.Time < ss[j].Value.Time
	})

	keys := make([]string, len(ss))
	for i, pair := range ss {
		keys[i] = pair.Key
	}
	return keys
}

func CreateResultFile(table map[string]*ResultsTable) {
	f, err := os.Create("result")
	if err != nil {
		log.Fatalf("cannot create result file: %v", err)
	}
	defer f.Close()
	writeResult := bufio.NewWriter(f)

	sortedKeys := SortByTime(table)

	for _, k := range sortedKeys {
		fmt.Fprintln(writeResult, TransformResultTable(table[k]))
	}

	if err := writeResult.Flush(); err != nil {
		log.Fatalf("cannot flush results: %v", err)
	}
}
