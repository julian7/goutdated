package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type UpdateInfo struct {
	Dir       string
	GoMod     string
	GoVersion string
	Indirect  bool
	Main      bool
	Path      string
	Time      string
	Update    NewerVersion
	Version   string
	visited   bool
}

type NewerVersion struct {
	Path    string
	Time    string
	Version string
}

func main() {
	dryrun := flag.Bool("n", false, "dryrun")
	all := flag.Bool("a", false, "show all updates")
	flag.Parse()

	golistcmd := exec.Command("go", "list", "-u", "-json", "-m", "all")
	data, err := golistcmd.Output()
	if err != nil {
		panic(err)
	}

	values := []UpdateInfo{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	for {
		var value UpdateInfo
		err = decoder.Decode(&value)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			panic(err)
		}
		values = append(values, value)
	}

	var me UpdateInfo
	var outdated []UpdateInfo

	for _, item := range values {
		if item.Main {
			me = item
			continue
		}
		if item.Update.Version != "" {
			if *all {
				fmt.Printf("Found outdated version of %s: %s -> %s\n", item.Path, item.Version, item.Update.Version)
			}
			outdated = append(outdated, item)
		}
	}

	if *all {
		return
	}

	modfile, err := os.Open(me.GoMod)
	if err != nil {
		panic(err)
	}
	gomoddata, err := io.ReadAll(modfile)
	if err != nil {
		panic(err)
	}
	modfile.Close()

	if *dryrun {
		var fname, version string
		scanner := bufio.NewScanner(bytes.NewReader(gomoddata))
		for scanner.Scan() {
			indirect := false
			n, err := fmt.Sscanf(scanner.Text(), "\t%s %s // indirect", &fname, &version)
			if err == nil && n == 2 {
				indirect = true
			} else {
				n, err := fmt.Sscanf(scanner.Text(), "\t%s %s", &fname, &version)
				if err != nil || n != 2 {
					continue
				}
			}

			indir := map[bool]string{
				false: "",
				true:  "// indirect",
			}
			for _, item := range outdated {
				if item.Path != fname {
					continue
				}
				if item.Version != version {
					panic(fmt.Sprintf("%s: version mismatch (want: %s, got: %s)", item.Path, item.Version, version))
				}
				if item.Indirect != indirect {
					panic(fmt.Sprintf("%s: indirection mismatch (%s vs %#v)", item.Path, scanner.Text(), item))
				}

				fmt.Printf(
					"%s %s -> %s %s\n",
					item.Path,
					item.Version,
					item.Update.Version,
					indir[item.Indirect],
				)
				item.visited = true
			}
		}
		return
	}
	orig := append(gomoddata[:0:0], gomoddata...)
	for _, item := range outdated {
		gomoddata = bytes.Replace(
			gomoddata,
			[]byte(fmt.Sprintf("\t%s %s", item.Path, item.Version)),
			[]byte(fmt.Sprintf("\t%s %s", item.Path, item.Update.Version)),
			1,
		)
	}
	if bytes.Compare(orig, gomoddata) != 0 {
		fi, err := os.Stat(me.GoMod)
		if err != nil {
			panic(err)
		}
		if err := os.WriteFile(me.GoMod, gomoddata, fi.Mode()); err != nil {
			panic(err)
		}
	}
}
