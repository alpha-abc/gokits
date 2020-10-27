package lruv1_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/alpha-abc/gokits/lru/lruv1"
)

type String string

func (s String) Len() int {
	return len(s)
}

func TestCache(t *testing.T) {
	var cache = lruv1.New(6, func(key string, val lruv1.Value) {
		//fmt.Println("remove", key, val)
	})
	cache.Add("a", String("A"))
	fmt.Println(cache)

	cache.Add("b", String("B"))
	fmt.Println(cache)

	cache.Add("c", String("C"))
	fmt.Println(cache)

	cache.Add("d", String("D"))
	fmt.Println(cache)

	cache.Get("a")
	fmt.Println(cache)

	cache.Get("b")
	fmt.Println(cache)

	cache.Add("c", String("C"))
	fmt.Println(cache)

	cache.Add("a", String("A"))
	fmt.Println(cache)
}

type S struct {
	name string
}

type S1 struct {
	name string
}

func TestEqual(t *testing.T) {
	var s1 = &S{
		name: "",
	}

	var s2 = &S{
		name: "",
	}

	var s3 *S = nil

	var _ = &S1{
		name: "",
	}

	fmt.Println(s1 == s2)
	fmt.Println(s1 == s3)
	fmt.Println(reflect.DeepEqual(s1, s2))

	fmt.Println(len(""))
}
