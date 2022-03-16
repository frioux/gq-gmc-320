package main

import (
	"fmt"

	"github.com/goburrow/serial"
)

func main() {
	port, err := serial.Open(&serial.Config{
		Address: "/dev/ttyUSB0",
		BaudRate: 115200,
		Parity: "N",
	})
	if err != nil {
		panic(err)
	}
	defer port.Close()

	if _, err := fmt.Fprint(port, "<GETVER>>"); err != nil {
		panic(err)
	}

	var n int
	ver := make([]byte, 14)
	if n, err = port.Read(ver); err != nil {
		panic(err)
	}

	fmt.Println(string(ver[:n]))
}
