package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"strings"

	"github.com/pkg/errors"
)

type depList struct {
	f    string
	deps []string
}

func (dp depList) Depends(s string) bool {
	for _, v := range dp.deps {
		if v == s {
			return true
		}
	}
	return false
}

func Print(dpl []depList) {
	fmt.Println("--DEPLIST--")
	for _, v := range dpl {
		fmt.Println(v.f)
		for _, vv := range v.deps {
			fmt.Println("\t" + vv)
		}
	}
}

func list(fname string) ([]string, error) {
	res := []string{}
	f, err := ioutil.ReadFile(fname)
	if err != nil {
		return res, err
	}

	for _, v := range strings.Split(string(f), "\n") {
		if strings.HasPrefix(v, "//dep ") {
			s := strings.TrimPrefix(v, "//dep ")
			s = strings.TrimSpace(s)
			res = append(res, s)
		}
	}
	return res, nil
}

func main() {
	_rt := flag.String("r", "", "Root directory")
	_s := flag.String("s", "", "Start File")

	flag.Parse()

	if *_s == "" {
		log.Fatal("Needs a start file See --help")
	}
	cur := []string{*_s}
	comp := []depList{}

lenny:
	for len(cur) > 0 {
		s := cur[0]
		cur = cur[1:]
		for _, v := range comp {
			if v.f == s {
				continue lenny
			}
		}
		deps, err := list(path.Join(*_rt, s))
		cur = append(cur, deps...)
		if err != nil {
			fmt.Printf("error in %s, %s\n", s, err)
			continue
		}
		comp = append(comp, depList{s, deps})
	}
	Print(comp)
	comp, _ = sortDeps(comp)
	Print(comp)

}

func sortDeps(ls []depList) ([]depList, error) {
	for k, v := range ls {
		//Move lowest dep down.
		for p := k + 1; p < len(ls); p++ {
			if ls[p].Depends(v.f) {
				ls[k], ls[p] = ls[p], ls[k]
			}
		}
		//Look for cycle.
		for p := k + 1; p < len(ls); p++ {
			if ls[p].Depends(v.f) {
				return ls, errors.Errorf("Cyclic Depedency %s,%s", ls[p].f, v.f)
			}
		}
	}
	return ls, nil
}
