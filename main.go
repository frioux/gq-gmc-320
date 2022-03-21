package main

import (
	"fmt"

	"github.com/frioux/gq-gmc-320/internal/gqclient"
)

func main() {
	cl, err := gqclient.New(115200)
	if err != nil {
		panic(err)
	}
	defer cl.Close()

	ver, err := cl.GetVer()
	if err != nil {
		panic(err)
	}

	fmt.Println(ver)

	cpm, err := cl.GetCPM()
	if err != nil {
		panic(err)
	}
	fmt.Println("cpm", cpm)

	// fmt.Println(cl.SetDateTime(time.Now()))
	// fmt.Println(cl.Reboot())
	// s, err := cl.GetSerial()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("serial", s)

	// if err := cl.PowerOff(); err != nil {
	// 	panic(err)
	// }
	// ch, err := cl.Heartbeat()
	// if err != nil {
	// 	panic(err)
	// }

	// for v := range ch {
	// 	fmt.Println("val", v)
	// }
}
