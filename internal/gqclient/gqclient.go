package gqclient

import (
	"fmt"
	"time"

	"github.com/goburrow/serial"
)

type Client struct{
	port serial.Port
}

func New(baudRate int) (Client, error) {
	var (
		cl Client
		err error
	)

	cl.port, err = serial.Open(&serial.Config{
		Address: "/dev/ttyUSB0",
		BaudRate: baudRate,
		Parity: "N",
	})
	if err != nil {
		return cl, err
	}

	return cl, nil
}

// Close shuts down the serial connection to the unit.
func (cl Client) Close() error { return cl.port.Close() }

// GetVer returns the version string.
func (cl Client) GetVer() (string, error) {
	if _, err := fmt.Fprint(cl.port, "<GETVER>>"); err != nil {
		return "", err
	}

	var (
		n int
		err error
	)
	ver := make([]byte, 14)
	if n, err = cl.port.Read(ver); err != nil {
		return "", err
	}

	return string(ver[:n]), nil
}

// GetCPM returns the current CPM.
func (cl Client) GetCPM() (int, error) {
	if _, err := fmt.Fprint(cl.port, "<GETCPM>>"); err != nil {
		return 0, err
	}

	buf := make([]byte, 2)
	if _, err := cl.port.Read(buf); err != nil {
		return 0, err
	}

	return int(buf[0])*256 + int(buf[1]), nil
}

func (cl Client) heartbeatOn() error {
	_, err := fmt.Fprint(cl.port, "<HEARTBEAT1>>")
	return err
}

func (cl Client) heartbeatOff() error {
	_, err := fmt.Fprint(cl.port, "<HEARTBEAT0>>")
	return err
}

func (cl Client) Heartbeat() (chan int, error) {
	if err := cl.heartbeatOn(); err != nil {
		return nil, err
	}

	ch := make(chan int, 10)

	i := 0
	go func() {
		defer cl.heartbeatOff()
		defer close(ch)
		for {
			buf := make([]byte, 2)
			i++
			if _, err := cl.port.Read(buf); err != nil {
				fmt.Println("error polling", err)
				return
			}

			fmt.Println("raw", buf)
			buf[0] = buf[0] &^ 128
			buf[0] = buf[0] &^ 64
			fmt.Println("masked", buf)
			ch <- int(buf[0])*256 + int(buf[1]) // XXX mask off top two bits
			if i > 10 {
				return
			}

		}
	}()

	return ch, nil
}

func (cl Client) GetVolt() (float32, error) {
	if _, err := fmt.Fprint(cl.port, "<GETVOLT>>"); err != nil {
		return 0, err
	}

	buf := make([]byte, 1)
	if _, err := cl.port.Read(buf); err != nil {
		return 0, err
	}

	return float32(buf[0])/10, nil
}

func (cl Client) ReadFlash(a2, a1, a0, l1, l0 byte) ([]byte, error) {
	cmd := []byte("<SPIR     >>")
	cmd[5] = a2
	cmd[6] = a1
	cmd[7] = a0
	cmd[8] = l1
	cmd[9] = l0

	if _, err := cl.port.Write(cmd); err != nil {
		return nil, err
	}

	buf := make([]byte, int(l1) * 256 + int(l0))
	if _, err := cl.port.Read(buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func (cl Client) GetCFG() ([]byte, error) {
	if _, err := fmt.Fprint(cl.port, "<GETCFG>>"); err != nil {
		return nil, err
	}

	buf := make([]byte, 256)
	if _, err := cl.port.Read(buf); err != nil {
		return nil, err
	}

	return buf, nil
}

func (cl Client) hasAck() error {
	buf := make([]byte, 1)
	if _, err := cl.port.Read(buf); err != nil {
		return err
	}

	if b := buf[0]; b != 0xAA {
		return fmt.Errorf("unexpected response: %X", b)
	}

	return nil
}

// XXX Check this
func (cl Client) EraseCFG() error {
	if _, err := fmt.Fprint(cl.port, "<ECFG>>"); err != nil {
		return err
	}

	return cl.hasAck()
}

// XXX Check this
func (cl Client) WriteCFG(a0, d0 byte) error {
	cmd := []byte("<WCFG  >>")
	cmd[5] = a0
	cmd[6] = d0

	if _, err := cl.port.Write(cmd); err != nil {
		return err
	}

	return cl.hasAck()
}

func (cl Client) SendKey(k byte) error {
	cmd := []byte("<KEY >>")
	cmd[4] = k

	_, err := cl.port.Write(cmd)
	return err
}

func (cl Client) GetSerial() ([]byte, error) {
	if _, err := fmt.Fprint(cl.port, "<GETSERIAL>>"); err != nil {
		return nil, err
	}

	buf := make([]byte, 7)
	if _, err := cl.port.Read(buf); err != nil {
		return nil, err
	}

	return buf, nil
}

func (cl Client) PowerOff() error {
	_, err := fmt.Fprint(cl.port, "<POWEROFF>>")
	return err
}

// XXX ???
func (cl Client) CfgUpdate() error {
	if _, err := fmt.Fprint(cl.port, "<CFGUPDATE>>"); err != nil {
		return err
	}

	return cl.hasAck()
}

func (cl Client) dateSet(which, b byte) error {
	cmd := []byte("<SETDATE   >>")
	cmd[8] = which
	cmd[9] = which
	cmd[10] = b

	if _, err := cl.port.Write(cmd); err != nil {
		return err
	}

	return cl.hasAck()
}

func (cl Client) SetYear(year byte) error { return cl.dateSet('Y', year) }
func (cl Client) SetMonth(month byte) error { return cl.dateSet('M', month) }
func (cl Client) SetDay(day byte) error { return cl.dateSet('D', day) }

func (cl Client) timeSet(which, b byte) error {
	cmd := []byte("<SETTIME   >>")
	cmd[8] = which
	cmd[9] = which
	cmd[10] = b

	if _, err := cl.port.Write(cmd); err != nil {
		return err
	}

	return cl.hasAck()
}

func (cl Client) SetHour(hour byte) error { return cl.timeSet('H', hour) }
func (cl Client) SetMinute(minute byte) error { return cl.timeSet('M', minute) }
func (cl Client) SetSecond(second byte) error { return cl.timeSet('S', second) }

func (cl Client) FactoryReset() error {
	if _, err := fmt.Fprint(cl.port, "<FACTORYRESET>>"); err != nil {
		return err
	}

	return cl.hasAck()
}

func (cl Client) Reboot() error {
	_, err := fmt.Fprint(cl.port, "<REBOOT>>")
	return err
}

func (cl Client) setDateTime(year, month, day, hour, minute, second byte) error {
	cmd := []byte("<SETDATETIMEYMDHMS>>")
	cmd[12] = year
	cmd[13] = month
	cmd[14] = day
	cmd[15] = hour
	cmd[16] = minute
	cmd[17] = second
	if _, err := cl.port.Write(cmd); err != nil {
		return err
	}

	return cl.hasAck()
}

func (cl Client) SetDateTime(t time.Time) error {
	return cl.setDateTime(
		byte(t.Year() - 2000),
		byte(t.Month()),
		byte(t.Day()),
		byte(t.Hour()),
		byte(t.Minute()),
		byte(t.Second()),
	)
}

func (cl Client) getDateTime() ([]byte, error) {
	if _, err := fmt.Fprint(cl.port, "<GETDATETIME>>"); err != nil {
		return nil, err
	}

	buf := make([]byte, 7)
	if _, err := cl.port.Read(buf); err != nil {
		return nil, err
	}

	if b := buf[6]; b != 0xAA {
		return nil, fmt.Errorf("unexpected trailer response: %X", b)
	}

	return buf[:6], nil
}

func (cl Client) GetDateTime() (time.Time, error) {
	raw, err := cl.getDateTime()
	if err != nil {
		return time.Now(), err
	}

	return time.Date(
		2000 + int(raw[0]), time.Month(raw[1]), int(raw[2]),
		int(raw[3]), int(raw[4]), int(raw[5]), 0, time.UTC), nil
}

func (cl Client) getTemp() ([]byte, error) {
	if _, err := fmt.Fprint(cl.port, "<GETTEMP>>"); err != nil {
		return nil, err
	}

	buf := make([]byte, 4)
	if _, err := cl.port.Read(buf); err != nil {
		return nil, err
	}

	if b := buf[3]; b != 0xAA {
		return nil, fmt.Errorf("unexpected trailer response: %X", b)
	}

	return buf[:3], nil
}

func (cl Client) GetTemp() (float32, error) {
	b, err := cl.getTemp()
	if err != nil {
		return 0, err
	}

	ret := float32(b[0]) + float32(b[1])/100
	if b[2] != 0 {
		ret = -ret
	}

	return ret, nil
}

func (cl Client) GetGyro() ([]byte, error) {
	if _, err := fmt.Fprint(cl.port, "<GETGYRO>>"); err != nil {
		return nil, err
	}

	buf := make([]byte, 7)
	if _, err := cl.port.Read(buf); err != nil {
		return nil, err
	}

	if b := buf[6]; b != 0xAA {
		return nil, fmt.Errorf("unexpected trailer response: %X", b)
	}

	return buf[:6], nil
}

func (cl Client) PowerOn() error {
	_, err := fmt.Fprint(cl.port, "<POWERON>>")
	return err
}
