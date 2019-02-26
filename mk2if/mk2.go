package mk2if

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"sync"
	"time"
)

type scaling struct {
	scale  float64
	offset float64
}

type mk2Ser struct {
	info   *Mk2Info
	report *Mk2Info
	p      io.ReadWriter
	sc     []scaling
	scN    int
	run    chan struct{}
	locked bool
	sync.RWMutex
	infochan chan *Mk2Info
	wg       sync.WaitGroup
}

func NewMk2Connection(dev io.ReadWriter) (Mk2If, error) {
	mk2 := &mk2Ser{}
	mk2.p = dev
	mk2.info = &Mk2Info{}
	mk2.report = &Mk2Info{}
	mk2.scN = 0
	mk2.locked = false
	mk2.sc = make([]scaling, 0)
	mk2.setTarget()
	mk2.run = make(chan struct{})
	mk2.infochan = make(chan *Mk2Info)
	mk2.wg.Add(1)
	go mk2.frameLock()
	return mk2, nil
}

// Locks to incoming frame.
func (mk2 *mk2Ser) frameLock() {

	frame := make([]byte, 256)
	var size byte
	for {
		select {
		case <-mk2.run:
			mk2.wg.Done()
			return
		default:
		}
		if mk2.locked {
			size = mk2.readByte()
			l, err := io.ReadFull(mk2.p, frame[0:int(size)+1])
			if err != nil {
				mk2.addError(fmt.Errorf("Read Error: %v", err))
				mk2.locked = false
			} else if l != int(size)+1 {
				mk2.addError(errors.New("Read Length Error"))
				mk2.locked = false
			} else {
				mk2.handleFrame(size, frame[0:int(size+1)])
			}
		} else {
			tmp := mk2.readByte()
			if tmp == 0xff || tmp == 0x20 {
				l, err := io.ReadFull(mk2.p, frame[0:int(size)])
				if err != nil {
					mk2.addError(fmt.Errorf("Read Error: %v", err))
					time.Sleep(1 * time.Second)
				} else if l != int(size) {
					mk2.addError(errors.New("Read Length Error"))
				} else {
					if checkChecksum(size, tmp, frame[0:int(size)]) {
						mk2.locked = true
						log.Printf("Locked")
					}
				}
			}
			size = tmp
		}
	}
}

// Close Mk2
func (mk2 *mk2Ser) Close() {
	close(mk2.run)
	mk2.wg.Wait()
}

// Returns last known state with all reported errors since previous poll.
// Mk2Info.Valid will be false if no polling has completed.
func (mk2 *mk2Ser) GetMk2Info() *Mk2Info {
	mk2.RLock()
	defer mk2.RUnlock()
	return mk2.report
}

func (mk2 *mk2Ser) C() chan *Mk2Info {
	return mk2.infochan
}

func (mk2 *mk2Ser) readByte() byte {
	buffer := make([]byte, 1)
	_, err := io.ReadFull(mk2.p, buffer)
	if err != nil {
		mk2.addError(fmt.Errorf("Read error: %v", err))
		return 0
	}
	return buffer[0]
}

// Adds error to error slice.
func (mk2 *mk2Ser) addError(err error) {
	if mk2.info.Errors == nil {
		mk2.info.Errors = make([]error, 0)
	}
	mk2.info.Errors = append(mk2.info.Errors, err)
	mk2.info.Valid = false
}

// Updates report.
func (mk2 *mk2Ser) updateReport() {
	mk2.Lock()
	defer mk2.Unlock()
	mk2.info.Timestamp = time.Now()
	mk2.report = mk2.info
	select {
	case mk2.infochan <- mk2.info:
	default:
	}
	mk2.info = &Mk2Info{}
}

// Checks for valid frame and chooses decoding.
func (mk2 *mk2Ser) handleFrame(l byte, frame []byte) {
	if checkChecksum(l, frame[0], frame[1:]) {
		switch frame[0] {
		case 0xff:
			switch frame[1] {
			case 0x56: // V
				mk2.versionDecode(frame[2:])
			case 0x57:
				switch frame[2] {
				case 0x8e:
					mk2.scaleDecode(frame[2:])
				case 0x85:
					mk2.stateDecode(frame[2:])
				}

			case 0x4C: // L
				mk2.ledDecode(frame[2:])
			}

		case 0x20:
			switch frame[5] {
			case 0x0C:
				mk2.dcDecode(frame[1:])
			case 0x08:
				mk2.acDecode(frame[1:])
			}
		}
	} else {
		log.Printf("Failed")
		mk2.locked = false
	}
}

// Set the target VBus device.
func (mk *mk2Ser) setTarget() {
	cmd := make([]byte, 3)
	cmd[0] = 0x41 // A
	cmd[1] = 0x01
	cmd[2] = 0x00
	mk.sendCommand(cmd)
}

// Request the scaling factor for entry 'in'.
func (mk *mk2Ser) reqScaleFactor(in byte) {
	cmd := make([]byte, 4)
	cmd[0] = 0x57 // W
	cmd[1] = 0x36
	cmd[2] = in
	mk.sendCommand(cmd)
}

// Decode the scale factor frame.
func (mk *mk2Ser) scaleDecode(frame []byte) {
	scl := uint16(frame[2])<<8 + uint16(frame[1])
	ofs := int16(uint16(frame[5])<<8 + uint16(frame[4]))

	tmp := scaling{}
	tmp.offset = float64(ofs)
	if scl >= 0x4000 {
		tmp.scale = math.Abs(1 / (0x8000 - float64(scl)))
	} else {
		tmp.scale = math.Abs(float64(scl))
	}
	mk.sc = append(mk.sc, tmp)

	mk.scN++
	if mk.scN < 14 {
		mk.reqScaleFactor(byte(mk.scN))
	} else {
		log.Print("Monitoring starting.")
	}

}

// Decode the version number
func (mk *mk2Ser) versionDecode(frame []byte) {
	mk.info.Version = 0
	mk.info.Valid = true
	for i := 0; i < 4; i++ {
		mk.info.Version += uint32(frame[i]) << uint(i) * 8
	}

	if mk.scN < 14 {
		log.Print("Get scaling factors.")
		mk.reqScaleFactor(byte(mk.scN))
	} else {
		// Send DC status request
		cmd := make([]byte, 2)
		cmd[0] = 0x46 //F
		cmd[1] = 0
		mk.sendCommand(cmd)
	}
}

// Apply scaling to float
func (mk *mk2Ser) applyScale(value float64, scale int) float64 {
	return mk.sc[scale].scale * (value + mk.sc[scale].offset)
}

// Convert bytes->int16->float
func getSigned(data []byte) float64 {
	return float64(int16(data[0]) + int16(data[1])<<8)
}

// Convert bytes->uint32->float
func getUnsigned(data []byte) float64 {
	return float64(uint32(data[0]) + uint32(data[1])<<8 + uint32(data[2])<<16)
}

// Decodes DC frame.
func (mk *mk2Ser) dcDecode(frame []byte) {
	mk.info.BatVoltage = mk.applyScale(getSigned(frame[5:7]), 4)

	usedC := mk.applyScale(getUnsigned(frame[7:10]), 5)
	chargeC := mk.applyScale(getUnsigned(frame[10:13]), 5)
	mk.info.BatCurrent = usedC - chargeC

	mk.info.OutFrequency = 10 / (mk.applyScale(float64(frame[13]), 7))

	// Send L1 status request
	cmd := make([]byte, 2)
	cmd[0] = 0x46 //F
	cmd[1] = 1
	mk.sendCommand(cmd)
}

// Decodes AC frame.
func (mk *mk2Ser) acDecode(frame []byte) {
	mk.info.InVoltage = mk.applyScale(getSigned(frame[5:7]), 0)
	mk.info.InCurrent = mk.applyScale(getSigned(frame[7:9]), 1)
	mk.info.OutVoltage = mk.applyScale(getSigned(frame[9:11]), 2)
	mk.info.OutCurrent = mk.applyScale(getSigned(frame[11:13]), 3)

	if frame[13] == 0xff {
		mk.info.InFrequency = 0
	} else {
		mk.info.InFrequency = 10 / (mk.applyScale(float64(frame[13]), 8))
	}

	// Send status request
	cmd := make([]byte, 1)
	cmd[0] = 0x4C //F
	mk.sendCommand(cmd)
}

// Decode charge state of battery.
func (mk *mk2Ser) stateDecode(frame []byte) {
	mk.info.ChargeState = mk.applyScale(getSigned(frame[1:3]), 13)
	mk.updateReport()
}

// Decode the LED state frame.
func (mk *mk2Ser) ledDecode(frame []byte) {

	mk.info.LEDs = getLEDs(frame[0], frame[1])
	// Send charge state request
	cmd := make([]byte, 4)
	cmd[0] = 0x57 //W
	cmd[1] = 0x30
	cmd[2] = 13
	mk.sendCommand(cmd)
}

// Adds active LEDs to list.
func getLEDs(ledsOn, ledsBlink byte) map[Led]LEDstate {

	leds := map[Led]LEDstate{}
	for i := 0; i < 8; i++ {
		on := (ledsOn >> uint(i)) & 1
		blink := (ledsBlink >> uint(i)) & 1
		if on == 1 {
			leds[Led(i)] = LedOn
		} else if blink == 1 {
			leds[Led(i)] = LedBlink
		} else {
			leds[Led(i)] = LedOff
		}
	}
	return leds
}

// Adds header and trailing crc for frame to send.
func (mk2 *mk2Ser) sendCommand(data []byte) {
	l := len(data)
	dataOut := make([]byte, l+3)
	dataOut[0] = byte(l + 1)
	dataOut[1] = 0xff
	cr := -dataOut[0] - dataOut[1]
	for i := 0; i < len(data); i++ {
		cr = cr - data[i]
		dataOut[i+2] = data[i]
	}
	dataOut[l+2] = cr

	_, err := mk2.p.Write(dataOut)
	if err != nil {
		mk2.addError(fmt.Errorf("Write error: %v", err))
	}
}

// Checks the frame crc.
func checkChecksum(l, t byte, d []byte) bool {
	cr := (uint16(l) + uint16(t)) % 256
	for i := 0; i < len(d); i++ {
		cr = (cr + uint16(d[i])) % 256
	}
	if cr == 0 {
		return true
	}
	return false
}
