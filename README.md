# relay-timer

```
tinygo flash -monitor -target arduino-nano33 -size short -ldflags '-X "main.ssid=xxx" -X "main.pass=yyy" 
-X "main.startHHMM=14:00" -X "main.stopHHMM=02:00"' ~/work/relay-timer/
```

startHHMM and stopHHMM are in UTC time.  So to set start/stop for 7am/7pm PST, add 7 hrs to get startHHMM/stopHHMM:

```
7am: 07:00 + 07:00 = 14:00
7pm: 19:00 + 07:00 = 26:00 % 24:00 = 02:00
```

Relay input is connected to D4 on Arduino Nano 33 IoT.
