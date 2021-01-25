package validatorv1

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"testing"
	//"github.com/tslisp/gokits/validator/validatorv1"
)

func TestReflect(t *testing.T) {
	var p = 123

	var ty = reflect.TypeOf(p)
	fmt.Println(ty.Kind())
	ty.Elem()
}

func TestNil(t *testing.T) {

	var arr []bool = make([]bool, 3)

	fmt.Println(string([]rune("阿斯顿")[0]), arr)
}

type S struct {
	Name string   `valid:"^[a-z]{3,6}$" message:"名称不合法"`
	Age  int      `valid:"^(1|2){1,1}[0-9]{1,1}$" message:""`
	Lst  []string `valid:"^aa$,required," message:""`
	S1   *S1      `valid:"required" message:"error"`
}

type S1 struct {
	Name string
}

func TestDo(t *testing.T) {
	var inf = &S{
		Name: "hel啊o",
		Age:  12,
		Lst:  []string{"aab", "aabbb"},
		//Lst: nil,
		S1: nil,
	}

	// var inf1 *S = nil
	// var inf **S = &inf1
	var errs = Validate(&inf)
	fmt.Println("result ====>")
	for _, err := range errs {
		fmt.Println(err)
	}
	fmt.Println("result ====>")

}

func TestExtract(t *testing.T) {
	var text = ",required,^,->,^,abc$,$"
	fmt.Println(RuleExtractFromTag(text))

}

func TestAst(t *testing.T) {
	var src = `
package main

import "fmt"

func main() {
	fmt.Println("hello world")
}
	`

	var fset = token.NewFileSet()
	var f, e = parser.ParseFile(fset, "", src, 0)
	if e != nil {
		panic(e)
	}
	ast.Print(fset, f)
}

/* valid 格式设计

<rule> , / , <rule>#1#2#3, <rule>


*/

func TestPattern(t *testing.T) {
	var tagStr = "abc,\\,hello"

	var s = Validator{}
	fmt.Println(s.ExtractFromTag(tagStr))
}

func TestExtractFromTag(t *testing.T) {
	var validator = &Validator{}

	var resp = validator.ExtractFromTag("user; email; min=10; message=名称不合法; -> ;")
	fmt.Println(resp)
}
