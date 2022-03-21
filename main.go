package main

import (
	"flag"
	"fmt"
	"encoding/hex"
	"time"
	"os/signal"
	"os"

	"github.com/frioux/gq-gmc-320/internal/gqclient"
)

func main() {
	cl, err := gqclient.New(115200)
	if err != nil {
		panic(err)
	}
	defer cl.Close()

	var (
		version, cpm, voltage, key0, key1, key2, key3 bool
		serial, on, off, reboot, temp, gyro bool
		heartbeat bool
		dateTime string
	)

	flag.BoolVar(&version, "version", false, "show version")
	flag.BoolVar(&cpm, "cpm", false, "show cpm")
	flag.BoolVar(&voltage, "voltage", false, "show voltage")
	flag.BoolVar(&key0, "key-0", false, "press key 0")
	flag.BoolVar(&key1, "key-1", false, "press key 1")
	flag.BoolVar(&key2, "key-2", false, "press key 2")
	flag.BoolVar(&key3, "key-3", false, "press key 3")
	flag.BoolVar(&serial, "serial", false, "show serial number")
	flag.BoolVar(&on, "on", false, "turn device on")
	flag.BoolVar(&off, "off", false, "turn device off")
	flag.BoolVar(&reboot, "reboot", false, "reboot device")
	flag.BoolVar(&temp, "temp", false, "show device temp")
	flag.BoolVar(&gyro, "gyro", false, "show gyroscope data")
	flag.BoolVar(&heartbeat, "heartbeat", false, "record heartbeat each second")
	flag.StringVar(&dateTime, "date-time", "", "set date with <YYYY-MM-DDTHH:MM:SS> or - to use system clock; use 'c' to show current value")
	flag.Parse()
	switch {
	case version:
		v, err := cl.GetVer()
		if err != nil {
			panic(err)
		}
		fmt.Println(v)
	case cpm:
		v, err := cl.GetCPM()
		if err != nil {
			panic(err)
		}
		fmt.Println(v)
	case voltage:
		v, err := cl.GetVolt()
		if err != nil {
			panic(err)
		}
		fmt.Println(v)
	case key0:
		err := cl.SendKey(0)
		if err != nil {
			panic(err)
		}
	case key1:
		err := cl.SendKey(1)
		if err != nil {
			panic(err)
		}
	case key2:
		err := cl.SendKey(2)
		if err != nil {
			panic(err)
		}
	case key3:
		err := cl.SendKey(3)
		if err != nil {
			panic(err)
		}
	case serial:
		s, err := cl.GetSerial()
		if err != nil {
			panic(err)
		}
		fmt.Println(hex.EncodeToString(s))
	case on:
		err := cl.PowerOn()
		if err != nil {
			panic(err)
		}
	case off:
		err := cl.PowerOff()
		if err != nil {
			panic(err)
		}
	case reboot:
		err := cl.Reboot()
		if err != nil {
			panic(err)
		}
	case temp:
		v, err := cl.GetTemp()
		if err != nil {
			panic(err)
		}
		fmt.Println(v)
	case gyro:
		v, err := cl.GetGyro()
		if err != nil {
			panic(err)
		}
		fmt.Println(v)
	case dateTime == "c":
		v, err := cl.GetDateTime()
		if err != nil {
			panic(err)
		}
		fmt.Println(v)
	case dateTime == "-":
		t := time.Now()
		fmt.Println("setting to", t)
		err := cl.SetDateTime(t)
		if err != nil {
			panic(err)
		}
	case dateTime != "":
		t, err := time.Parse("2006-01-02T15:04:05", dateTime)
		if err != nil {
			panic(err)
		}
		fmt.Println("setting to", t)
		err = cl.SetDateTime(t)
		if err != nil {
			panic(err)
		}
	case heartbeat:
		if err := cl.HeartbeatOn(); err != nil {
			panic(err)
		}

		ch := make(chan os.Signal, 10)
		signal.Notify(ch, os.Interrupt)

		window := make([]int, 0, 60)
		total := 0

		fmt.Println("Press CTRL-C too stop polling")
		running := true
		go func() {
			<-ch
			fmt.Println("CTRL-C detected, shutting down...")
			running = false
		}()

		for running {
			beats, err := cl.ReadHeartBeat()
			if err != nil {
				cl.HeartbeatOff()
				panic(err)
			}
			if len(window) < 60 {
				total += beats
				window = append(window, beats) // memory leak?
			} else {
				total += beats - window[0]
				window = append(window[1:], beats)
			}
			fmt.Println("beats per second", beats, "per minute", total)

		}
		cl.HeartbeatOff()
	}

}
