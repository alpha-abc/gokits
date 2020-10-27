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

	fmt.Println(errorsv1.DetailString(err))
	fmt.Println(err.Error())

	err = nil
	fmt.Println(errorsv1.DetailString(err))
}
