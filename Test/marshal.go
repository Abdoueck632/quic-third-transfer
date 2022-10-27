package main

import (
	"encoding/json"
	"fmt"
)

type Num struct {
	N int
}

type Packed struct {
	pNum *Num
	Name string
}

func main() {
	num := &Num{N: 100}

	packed := &Packed{pNum: num, Name: "xx-packed-xy"}

	fmt.Printf(" %+v\n", packed)

	dataInBytes, err := json.Marshal(packed)
	

	unpacked := &Packed{}
	err = json.Unmarshal(dataInBytes, unpacked)
	fmt.Printf("%v, %+v\n", err, unpacked)
}
