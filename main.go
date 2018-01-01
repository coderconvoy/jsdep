package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

type dep struct {
	name string
	deps []dep
	ex   bool
}

func ndep(n string, ex ...bool) dep {
	if len(ex) == 0 {
		return dep{name: n}
	}
	return dep{name: n, ex: ex[0]}
}

func (dp dep) Depends(s dep) bool {
	for _, v := range dp.deps {
		if v.name == s.name {
			return true
		}
	}
	return false
}

func Print(dpl []dep) {
	fmt.Println("--DEPLIST--")
	for _, v := range dpl {
		fmt.Println(v.name)
		for _, vv := range v.deps {
			fmt.Println("\t" + vv.name)
		}
	}
}

func HTMLString(dpl []dep, hpath string) string {
	res := ""
	for k := len(dpl) - 1; k >= 0; k-- {
		v := dpl[k]
		s := v.name
		if !v.ex {
			s = path.Join(hpath, v.name)
		}
		res += fmt.Sprintf("<script src=\"%s\"></script>\n", s)
	}
	return res
}

func PrintHTML(w io.Writer, dpl []dep, hpath string) {
	w.Write([]byte(HTMLString(dpl, hpath)))
}

func getDeps(d dep, root string) (dep, error) {

	if d.ex {
		return d, nil
	}

	f, err := ioutil.ReadFile(path.Join(root, d.name))
	if err != nil {
		return d, err
	}

	for _, v := range strings.Split(string(f), "\n") {
		if strings.HasPrefix(v, "//dep ") {
			s := strings.TrimPrefix(v, "//dep ")
			s = strings.TrimSpace(s)
			d.deps = append(d.deps, ndep(s, false))
		}
		if strings.HasPrefix(v, "//exdep ") {
			s := strings.TrimPrefix(v, "//exdep ")
			s = strings.TrimSpace(s)
			d.deps = append(d.deps, ndep(s, true))
		}
	}
	return d, nil
}

func Dig(root string, ss ...string) ([]dep, error) {
	res := []dep{}

	cur := []dep{}
	for _, v := range ss {
		cur = append(cur, ndep(v, false))
	}

lenny:
	for len(cur) > 0 {
		s := cur[0]
		cur = cur[1:]
		for _, v := range res {
			if v.name == s.name {
				continue lenny
			}
		}
		d, err := getDeps(s, root)
		cur = append(cur, d.deps...)
		if err != nil {
			return res, errors.Errorf("error in %s, %s\n", s, err)
		}
		res = append(res, d)
	}
	return res, nil
}

func main() {
	_rt := flag.String("r", "", "Root directory")
	_s := flag.String("s", "", "Start File")
	_hpath := flag.String("h", "", "html-path")
	_w := flag.Bool("w", false, "Add html wrapper")
	_tp := flag.String("tp", "", "templatefile")

	flag.Parse()

	if *_tp != "" {
		//run template file
		rt := *_rt
		if rt == "" {
			rt = path.Dir(*_tp)
		}

		t, err := template.New("bill").Funcs(template.FuncMap{
			"jsdep": func(ss ...string) (string, error) {
				comp, err := Dig(rt, ss...)
				if err != nil {
					return "", err
				}

				comp, err = sortDeps(comp)
				if err != nil {
					return "", err
				}
				return HTMLString(comp, *_hpath), nil
			},
		}).ParseFiles(*_tp)

		if err != nil {
			log.Fatal(err)
		}

		err = t.ExecuteTemplate(os.Stdout, path.Base(*_tp), nil)
		if err != nil {
			log.Fatal(err)
		}

		return
	}

	if *_s == "" {
		log.Fatal("Needs a start file See --help")
	}

	comp, err := Dig(*_rt, *_s)
	if err != nil {
		fmt.Println(err)
	}

	comp, err = sortDeps(comp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}

	if *_w {
		fmt.Fprintln(os.Stdout, `<!DOCTYPE html><head><meta charset="utf-8"></head><body>`)
	}

	PrintHTML(os.Stdout, comp, *_hpath)

	if *_w {
		fmt.Fprintln(os.Stdout, `</body><\html>`)
	}

}

func sortDeps(ls []dep) ([]dep, error) {
	for k, v := range ls {
		//Move lowest dep down.
		for p := k + 1; p < len(ls); p++ {
			if ls[p].Depends(v) {
				ls[k], ls[p] = ls[p], ls[k]
			}
		}
	}
	if !inorder(ls) {
		return ls, errors.Errorf("Cyclic dependency")
	}
	return ls, nil
}

func inorder(dl []dep) bool {
	for k, v := range dl {
		for i := 0; i < k; i++ {
			if v.Depends(dl[i]) {
				return false
			}
		}
	}
	return true
}
