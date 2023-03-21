// This is an timer-based relay controller
//
// It creates a UDP connection to request the current time and parse the
// response from a NTP server.  The system time is set to NTP time.
// The relay is turned on at startHHMM and turned off at stopHHMM.

package main

import (
	"fmt"
	"io"
	"log"
	"machine"
	"net"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	ssid string
	pass string
	startHHMM string
	stopHHMM string
	ntpHost string = "0.pool.ntp.org:123"
)

const NTP_PACKET_SIZE = 48

var response = make([]byte, NTP_PACKET_SIZE)

var startTimer *time.Timer
var stopTimer *time.Timer

var relay = machine.D4

func main() {

	//waitSerial()
	time.Sleep(2 * time.Second)

	relay.Configure(machine.PinConfig{machine.PinOutput})
	relayOff()

	if err := netdev.NetConnect(); err != nil {
		log.Fatal(err)
	}

	conn, err := net.Dial("udp", ntpHost)
	if err != nil {
		log.Fatal(err)
	}

	message("Requesting NTP time...")

	now, err := getCurrentTime(conn)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error getting current time: %v", err))
	} else {
		message("NTP time: %v", now)
	}

	conn.Close()
	netdev.NetDisconnect()

	runtime.AdjustTimeOffset(-1 * int64(time.Since(now)))

	for i := 0; i < 2; i++ {
		relayOn()
		time.Sleep(2 * time.Second)
		relayOff()
		time.Sleep(2 * time.Second)
	}

	if isCurrentTimeBetween(startHHMM, stopHHMM) {
		relayOn()
	}

	startTimer = newTimer(startHHMM, relayOn)
	stopTimer = newTimer(stopHHMM, relayOff)

	select {}
}

func isCurrentTimeBetween(start, stop string) bool {
	// Parse start and stop times
	startTime, err := time.Parse("15:04", start)
	if err != nil {
		panic(err)
	}
	stopTime, err := time.Parse("15:04", stop)
	if err != nil {
		panic(err)
	}

	// Get current time
	currentTime := time.Now().UTC()
	currentTime = time.Date(0, 1, 1, currentTime.Hour(), currentTime.Minute(), 0, 0, currentTime.Location())

	// Check if current time is between start and stop times
	if startTime.After(stopTime) {
		return currentTime.After(startTime) || currentTime.Before(stopTime)
	} else {
		return currentTime.After(startTime) && currentTime.Before(stopTime)
	}

	return false
}

func relayOn() {
	message("Relay ON")
	relay.High()
	if startTimer != nil {
		startTimer.Reset(24 * time.Hour)
	}
}

func relayOff() {
	message("Relay OFF")
	relay.Low()
	if stopTimer != nil {
		stopTimer.Reset(24 * time.Hour)
	}
}

func getHoursAndMinutes(start string) (hours, minutes int) {
	parts := strings.Split(start, ":")
	if len(parts) == 2 {
		hours, _ = strconv.Atoi(parts[0])
		minutes, _ = strconv.Atoi(parts[1])
	}
	return
}

func newTimer(when string, f func()) *time.Timer {
	now := time.Now()
	hours, minutes := getHoursAndMinutes(when)
	then := time.Date(now.Year(), now.Month(), now.Day(), hours, minutes, 0, 0, now.Location())
	if now.After(then) {
		then = then.Add(24 * time.Hour) // add 24 hours to "then" if it's already passed today
	}
	wait := then.Sub(now)
	message("firing in %s", wait)
	return time.AfterFunc(wait, f)
}

// Wait for user to open serial console
func waitSerial() {
	for !machine.Serial.DTR() {
		time.Sleep(100 * time.Millisecond)
	}
}

func getCurrentTime(conn net.Conn) (time.Time, error) {
	if err := sendNTPpacket(conn); err != nil {
		return time.Time{}, err
	}

	n, err := conn.Read(response)
	if err != nil && err != io.EOF {
		return time.Time{}, err
	}
	if n != NTP_PACKET_SIZE {
		return time.Time{}, fmt.Errorf("expected NTP packet size of %d: %d", NTP_PACKET_SIZE, n)
	}

	return parseNTPpacket(response), nil
}

func sendNTPpacket(conn net.Conn) error {
	var request = [48]byte{
		0xe3,
	}

	_, err := conn.Write(request[:])
	return err
}

func parseNTPpacket(r []byte) time.Time {
	// the timestamp starts at byte 40 of the received packet and is four bytes,
	// this is NTP time (seconds since Jan 1 1900):
	t := uint32(r[40])<<24 | uint32(r[41])<<16 | uint32(r[42])<<8 | uint32(r[43])
	const seventyYears = 2208988800
	return time.Unix(int64(t-seventyYears), 0)
}

func message(format string, args ...interface{}) {
	println(fmt.Sprintf(format, args...), "\r")
}
