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
	info       *Mk2Info
	p          io.ReadWriter
	scales     []scaling
	scaleCount int
	run        chan struct{}
	frameLock  bool
	infochan   chan *Mk2Info
	wg         sync.WaitGroup
}

func NewMk2Connection(dev io.ReadWriter) (Mk2If, error) {
	mk2 := &mk2Ser{}
	mk2.p = dev
	mk2.info = &Mk2Info{}
	mk2.scaleCount = 0
	mk2.frameLock = false
	mk2.scales = make([]scaling, 0, 14)
	mk2.setTarget()
	mk2.run = make(chan struct{})
	mk2.infochan = make(chan *Mk2Info)
	mk2.wg.Add(1)
	go mk2.frameLocker()
	return mk2, nil
}

// Locks to incoming frame.
func (m *mk2Ser) frameLocker() {

	frame := make([]byte, 256)
	var size byte
	for {
		select {
		case <-m.run:
			m.wg.Done()
			return
		default:
		}
		if m.frameLock {
			size = m.readByte()
			l, err := io.ReadFull(m.p, frame[0:int(size)+1])
			if err != nil {
				m.addError(fmt.Errorf("Read Error: %v", err))
				m.frameLock = false
			} else if l != int(size)+1 {
				m.addError(errors.New("Read Length Error"))
				m.frameLock = false
			} else {
				m.handleFrame(size, frame[0:int(size+1)])
			}
		} else {
			tmp := m.readByte()
			if tmp == 0xff || tmp == 0x20 {
				l, err := io.ReadFull(m.p, frame[0:int(size)])
				if err != nil {
					m.addError(fmt.Errorf("Read Error: %v", err))
					time.Sleep(1 * time.Second)
				} else if l != int(size) {
					m.addError(errors.New("Read Length Error"))
				} else {
					if checkChecksum(size, tmp, frame[0:int(size)]) {
						m.frameLock = true
						log.Printf("Locked")
					}
				}
			}
			size = tmp
		}
	}
}

// Close Mk2
func (m *mk2Ser) Close() {
	close(m.run)
	m.wg.Wait()
}

func (m *mk2Ser) C() chan *Mk2Info {
	return m.infochan
}

func (m *mk2Ser) readByte() byte {
	buffer := make([]byte, 1)
	_, err := io.ReadFull(m.p, buffer)
	if err != nil {
		m.addError(fmt.Errorf("Read error: %v", err))
		return 0
	}
	return buffer[0]
}

// Adds error to error slice.
func (m *mk2Ser) addError(err error) {
	if m.info.Errors == nil {
		m.info.Errors = make([]error, 0)
	}
	m.info.Errors = append(m.info.Errors, err)
	m.info.Valid = false
}

// Updates report.
func (m *mk2Ser) updateReport() {
	m.info.Timestamp = time.Now()
	select {
	case m.infochan <- m.info:
	default:
	}
	m.info = &Mk2Info{}
}

// Checks for valid frame and chooses decoding.
func (m *mk2Ser) handleFrame(l byte, frame []byte) {
	if checkChecksum(l, frame[0], frame[1:]) {
		switch frame[0] {
		case 0xff:
			switch frame[1] {
			case 0x56: // V
				m.versionDecode(frame[2:])
			case 0x57:
				switch frame[2] {
				case 0x8e:
					m.scaleDecode(frame[2:])
				case 0x85:
					m.stateDecode(frame[2:])
				}

			case 0x4C: // L
				m.ledDecode(frame[2:])
			}

		case 0x20:
			switch frame[5] {
			case 0x0C:
				m.dcDecode(frame[1:])
			case 0x08:
				m.acDecode(frame[1:])
			}
		}
	} else {
		log.Printf("Invalid incoming frame checksum: %x", frame)
		m.frameLock = false
	}
}

// Set the target VBus device.
func (m *mk2Ser) setTarget() {
	cmd := make([]byte, 3)
	cmd[0] = 0x41 // A
	cmd[1] = 0x01
	cmd[2] = 0x00
	m.sendCommand(cmd)
}

// Request the scaling factor for entry 'in'.
func (m *mk2Ser) reqScaleFactor(in byte) {
	cmd := make([]byte, 4)
	cmd[0] = 0x57 // W
	cmd[1] = 0x36
	cmd[2] = in
	m.sendCommand(cmd)
}

// Decode the scale factor frame.
func (m *mk2Ser) scaleDecode(frame []byte) {
	scl := uint16(frame[2])<<8 + uint16(frame[1])
	ofs := int16(uint16(frame[5])<<8 + uint16(frame[4]))

	tmp := scaling{}
	tmp.offset = float64(ofs)
	if scl >= 0x4000 {
		tmp.scale = math.Abs(1 / (0x8000 - float64(scl)))
	} else {
		tmp.scale = math.Abs(float64(scl))
	}
	m.scales = append(m.scales, tmp)

	m.scaleCount++
	if m.scaleCount < 14 {
		m.reqScaleFactor(byte(m.scaleCount))
	} else {
		log.Print("Monitoring starting.")
	}

}

// Decode the version number
func (m *mk2Ser) versionDecode(frame []byte) {
	m.info.Version = 0
	m.info.Valid = true
	for i := 0; i < 4; i++ {
		m.info.Version += uint32(frame[i]) << uint(i) * 8
	}

	if m.scaleCount < 14 {
		log.Print("Get scaling factors.")
		m.reqScaleFactor(byte(m.scaleCount))
	} else {
		// Send DC status request
		cmd := make([]byte, 2)
		cmd[0] = 0x46 //F
		cmd[1] = 0
		m.sendCommand(cmd)
	}
}

// Apply scaling to float
func (m *mk2Ser) applyScale(value float64, scale int) float64 {
	return m.scales[scale].scale * (value + m.scales[scale].offset)
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
func (m *mk2Ser) dcDecode(frame []byte) {
	m.info.BatVoltage = m.applyScale(getSigned(frame[5:7]), 4)

	usedC := m.applyScale(getUnsigned(frame[7:10]), 5)
	chargeC := m.applyScale(getUnsigned(frame[10:13]), 5)
	m.info.BatCurrent = usedC - chargeC

	m.info.OutFrequency = 10 / (m.applyScale(float64(frame[13]), 7))

	// Send L1 status request
	cmd := make([]byte, 2)
	cmd[0] = 0x46 //F
	cmd[1] = 1
	m.sendCommand(cmd)
}

// Decodes AC frame.
func (m *mk2Ser) acDecode(frame []byte) {
	m.info.InVoltage = m.applyScale(getSigned(frame[5:7]), 0)
	m.info.InCurrent = m.applyScale(getSigned(frame[7:9]), 1)
	m.info.OutVoltage = m.applyScale(getSigned(frame[9:11]), 2)
	m.info.OutCurrent = m.applyScale(getSigned(frame[11:13]), 3)

	if frame[13] == 0xff {
		m.info.InFrequency = 0
	} else {
		m.info.InFrequency = 10 / (m.applyScale(float64(frame[13]), 8))
	}

	// Send status request
	cmd := make([]byte, 1)
	cmd[0] = 0x4C //F
	m.sendCommand(cmd)
}

// Decode charge state of battery.
func (m *mk2Ser) stateDecode(frame []byte) {
	m.info.ChargeState = m.applyScale(getSigned(frame[1:3]), 13)
	m.updateReport()
}

// Decode the LED state frame.
func (m *mk2Ser) ledDecode(frame []byte) {

	m.info.LEDs = getLEDs(frame[0], frame[1])
	// Send charge state request
	cmd := make([]byte, 4)
	cmd[0] = 0x57 //W
	cmd[1] = 0x30
	cmd[2] = 13
	m.sendCommand(cmd)
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
func (m *mk2Ser) sendCommand(data []byte) {
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

	_, err := m.p.Write(dataOut)
	if err != nil {
		m.addError(fmt.Errorf("Write error: %v", err))
	}
}

// Checks the frame crc.
func checkChecksum(l, t byte, d []byte) bool {
	cr := (uint16(l) + uint16(t)) % 256
	for i := 0; i < len(d); i++ {
		cr = (cr + uint16(d[i])) % 256
	}
	return cr == 0
}
