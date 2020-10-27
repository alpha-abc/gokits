package snowflakev1_test

import (
	"fmt"
	"testing"

	"github.com/alpha-abc/gokits/snowflake/snowflakev1"
)

func Test(t *testing.T) {
	var node, e = snowflakev1.NewNode(2)
	if e != nil {
		fmt.Println(e.Error())
		return
	}

	for i := 0; i < 100; i++ {
		var id, err = node.Generate()
		fmt.Println(id, err)
		var head, tt, node, seq = snowflakev1.Extract(id)
		fmt.Println(head, tt, node, seq)
	}
}
