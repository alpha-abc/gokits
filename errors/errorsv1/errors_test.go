package errorsv1_test

import (
	"fmt"
	"testing"

	"github.com/alpha-abc/gokits/errors/errorsv1"
)

func TestExample(t *testing.T) {

	var err = errorsv1.New("asmerr", "ker err")
	err = fmt.Errorf("file not found %w", err)

	err = errorsv1.Wrap("500", "system error", err)
	err = errorsv1.Wrap("1001", "operator failure", err)
	err = errorsv1.SimpleWrap("simple wrap", err)

	fmt.Println(errorsv1.DetailString(err))
	fmt.Println(err.Error())

	err = nil
	fmt.Println(errorsv1.DetailString(err))
}

func TestDetailString(t *testing.T) {
	var err = fmt.Errorf("err1")
	err = fmt.Errorf("err2: %w", err)
	err = fmt.Errorf("err3: %w", err)
	err = fmt.Errorf("err4: %w", err)

	fmt.Println(errorsv1.DetailString(err))
}
