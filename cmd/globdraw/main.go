package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode/utf8"

	"github.com/gobwas/glob"
	"github.com/gobwas/glob/match"
)

func main() {
	var (
		pattern  = flag.String("p", "", "pattern to draw")
		sep      = flag.String("s", "", "comma separated list of separators characters")
		filepath = flag.String("file", "", "path for patterns file")
		auto     = flag.Bool("auto", false, "autoopen result")
		offset   = flag.Int("offset", 0, "patterns to skip")
	)
	flag.Parse()

	var patterns []string
	if *pattern != "" {
		patterns = append(patterns, *pattern)
	}
	if *filepath != "" {
		file, err := os.Open(*filepath)
		if err != nil {
			fmt.Printf("could not open file: %v\n", err)
			os.Exit(1)
		}
		s := bufio.NewScanner(file)
		for s.Scan() {
			fmt.Println(*offset)
			if *offset > 0 {
				*offset--
				fmt.Println("skipped")
				continue
			}
			patterns = append(patterns, s.Text())
		}
		file.Close()
	}
	if len(patterns) == 0 {
		return
	}

	var separators []rune
	if len(*sep) > 0 {
		for _, c := range strings.Split(*sep, ",") {
			r, w := utf8.DecodeRuneInString(c)
			if len(c) > w {
				fmt.Println("only single charactered separators are allowed: %+q", c)
				os.Exit(1)
			}
			separators = append(separators, r)
		}
	}

	br := bufio.NewReader(os.Stdin)
	for _, p := range patterns {
		g, err := glob.Compile(p, separators...)
		if err != nil {
			fmt.Printf("could not compile pattern %+q: %v\n", p, err)
			os.Exit(1)
		}
		s := match.Graphviz(p, g.(match.Matcher))
		if *auto {
			fmt.Fprintf(os.Stdout, "pattern: %+q: ", p)
			if err := open(s); err != nil {
				fmt.Printf("could not open graphviz: %v", err)
				os.Exit(1)
			}
			if !next(br) {
				return
			}
		} else {
			fmt.Fprintln(os.Stdout, s)
		}
	}
}

func open(s string) error {
	file, err := os.Create("glob.graphviz.png")
	if err != nil {
		return err
	}
	defer file.Close()
	cmd := exec.Command("dot", "-Tpng")
	cmd.Stdin = strings.NewReader(s)
	cmd.Stdout = file
	if err := cmd.Run(); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	cmd = exec.Command("open", file.Name())
	return cmd.Run()
}

func next(in *bufio.Reader) bool {
	fmt.Fprint(os.Stdout, "cancel? [Y/n]: ")
	p, err := in.ReadBytes('\n')
	if err != nil {
		return false
	}
	if p[0] == 'Y' {
		return false
	}
	return true
}
