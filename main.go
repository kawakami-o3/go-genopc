package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/rakyll/statik/fs"
	"golang.org/x/tools/go/packages"

	_ "github.com/kawakami-o3/go-genopc/statik"
)

type opcode struct {
	id    int
	name  string
	arity int
}

func parseOpcode(ln string) opcode {
	ss := strings.Split(ln, ": ")
	ts := strings.Split(ss[1], "/")

	id, _ := strconv.Atoi(ss[0])
	name := ts[0]
	arity, _ := strconv.Atoi(ts[1])
	return opcode{id, name, arity}
}

type genOp struct {
	beamFormatNumber int
	opcodes          []opcode
}

func newGenOp() *genOp {
	ret := &genOp{}
	ret.opcodes = append(ret.opcodes, parseOpcode("0: /0"))
	return ret
}

func readGenOpTab() (string, error) {
	statikFS, err := fs.New()
	if err != nil {
		return "", err
	}
	file, err := statikFS.Open("/genop.tab")
	if err != nil {
		return "", err
	}
	cnt, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(cnt), nil
}

func parseGenOpTab(cnt string) *genOp {
	genOp := newGenOp()
	for _, ln := range strings.Split(string(cnt), "\n") {
		if len(ln) == 0 {
			continue
		}
		if ln[0] == '#' {
			continue
		}
		if strings.Index(ln, "BEAM_FORMAT_NUMBER") >= 0 {
			genOp.beamFormatNumber, _ = strconv.Atoi(ln[len("BEAM_FORMAT_NUMBER="):])
		}
		if strings.Index(ln, ": ") >= 0 {
			genOp.opcodes = append(genOp.opcodes, parseOpcode(ln))
		}
	}
	return genOp
}

type Buffer struct {
	buf bytes.Buffer
}

func newBuffer() *Buffer {
	return &Buffer{
		buf: *bytes.NewBuffer([]byte{}),
	}
}

func (this *Buffer) Printf(format string, args ...interface{}) {
	fmt.Fprintf(&this.buf, format, args...)
}

func (this *Buffer) format() ([]byte, error) {
	src, err := format.Source(this.buf.Bytes())
	if err != nil {
		return []byte{}, err
	}
	return src, nil
}

func parsePackage() (string, error) {
	cfg := &packages.Config{
		Mode:  packages.LoadSyntax,
		Tests: false,
	}
	pkgs, err := packages.Load(cfg)
	if err != nil {
		return "", err
	}
	if len(pkgs) == 0 {
		return "", errors.New("package not found")
	}
	return pkgs[0].Name, nil
}

func main() {
	cnt, err := readGenOpTab()
	if err != nil {
		panic(err)
	}

	packageName, err := parsePackage()
	if err != nil {
		panic(err)
	}

	genOp := parseGenOpTab(cnt)

	buf := newBuffer()
	//buf.Println("package", packageName)

	buf.Printf("package %s\n", packageName)

	buf.Printf("const BEAM_FORMAT_NUMBER=%d\n", genOp.beamFormatNumber)

	buf.Printf("type Opcode struct {\n")
	buf.Printf("  Id int\n")
	buf.Printf("  Name string\n")
	buf.Printf("  Arity int\n")
	buf.Printf("}\n")

	buf.Printf("func op(id int, name string, arity int) Opcode {\n")
	buf.Printf("  return Opcode{ Id: id, Name: name, Arity: arity }\n")
	buf.Printf("}\n")

	buf.Printf("var Opcodes = []Opcode{\n")
	for _, o := range genOp.opcodes {
		buf.Printf("  op(%d, \"%s\", %d),\n", o.id, o.name, o.arity)
	}
	buf.Printf("}\n")

	src, err := buf.format()
	if err != nil {
		panic(err)
	}

	outName := packageName + "_gen.go"
	err = ioutil.WriteFile(outName, src, 0644)
	if err != nil {
		panic(err)
	}
}
