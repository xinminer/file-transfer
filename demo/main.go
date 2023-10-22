package main

import (
	"fmt"
	"github.com/gogf/gf/v2/text/gstr"
)

func main() {
	comp := gstr.Explode(".", "10.0.13.15")
	comp = comp[:3]
	prefix := gstr.Implode(".", comp)
	fmt.Println(prefix)
}
