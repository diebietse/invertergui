package mk2driver

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type scaling struct {
	scale     float64
	offset    float64
	signed    bool
	supported bool
}

//nolint:deadcode,varcheck
const (
	ramVarVMains = iota
	ramVarIMains
	ramVarVInverter
	ramVarIInverter
	ramVarVBat
	ramVarIBat
	ramVarVBatRipple
	ramVarInverterPeriod
	ramVarMainPeriod
	ramVarIACLoad
	ramVarVirSwitchPos
	ramVarIgnACInState
	ramVarMultiFuncRelay
	ramVarChargeState
	ramVarInverterPower1
	ramVarInverterPower2
	ramVarOutPower

	ramVarMaxOffset = 14
)

const (
	infoFrameHeader = 0x20
	frameHeader     = 0xff
)

const (
	acL1InfoFrame  = 0x08
	dcInfoFrame    = 0x0C
	setTargetFrame = 0x41
	infoReqFrame   = 0x46
	ledFrame       = 0x4C
	vFrame         = 0x56
	winmonFrame    = 0x57
)

// info frame types
const (
	infoReqAddrDC   = 0x00
	infoReqAddrACL1 = 0x01
)

// winmon frame commands
const (
	commandReadRAMVar    = 0x30
	commandGetRAMVarInfo = 0x36

	commandReadRAMResponse       = 0x85
	commandGetRAMVarInfoResponse = 0x8E
)

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

func NewMk2Connection(dev io.ReadWriter) (Mk2, error) {
	mk2 := &mk2Ser{}
	mk2.p = dev
	mk2.info = &Mk2Info{}
	mk2.scaleCount = 0
	mk2.frameLock = false
	mk2.scales = make([]scaling, 0, ramVarMaxOffset)
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
	var frameLength byte
	for {
		select {
		case <-m.run:
			m.wg.Done()
			return
		default:
		}
		if m.frameLock {
			frameLength = m.readByte()
			frameLengthOffset := int(frameLength) + 1
			l, err := io.ReadFull(m.p, frame[:frameLengthOffset])
			if err != nil {
				m.addError(fmt.Errorf("Read Error: %v", err))
				m.frameLock = false
			} else if l != frameLengthOffset {
				m.addError(errors.New("Read Length Error"))
				m.frameLock = false
			} else {
				m.handleFrame(frameLength, frame[:frameLengthOffset])
			}
		} else {
			tmp := m.readByte()
			frameLengthOffset := int(frameLength)
			if tmp == frameHeader || tmp == infoFrameHeader {
				l, err := io.ReadFull(m.p, frame[:frameLengthOffset])
				if err != nil {
					m.addError(fmt.Errorf("Read Error: %v", err))
					time.Sleep(1 * time.Second)
				} else if l != frameLengthOffset {
					m.addError(errors.New("Read Length Error"))
				} else {
					if checkChecksum(frameLength, tmp, frame[:frameLengthOffset]) {
						m.frameLock = true
						logrus.Info("Locked")
					}
				}
			}
			frameLength = tmp
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
	logrus.Debugf("frame %#v", frame)
	if checkChecksum(l, frame[0], frame[1:]) {
		switch frame[0] {
		case frameHeader:
			switch frame[1] {
			case vFrame:
				m.versionDecode(frame[2:])
			case winmonFrame:
				switch frame[2] {
				case commandGetRAMVarInfoResponse:
					m.scaleDecode(frame[2:])
				case commandReadRAMResponse:
					m.stateDecode(frame[2:])
				}

			case ledFrame:
				m.ledDecode(frame[2:])
			}

		case infoFrameHeader:
			switch frame[5] {
			case dcInfoFrame:
				m.dcDecode(frame[1:])
			case acL1InfoFrame:
				m.acDecode(frame[1:])
			}
		}
	} else {
		logrus.Errorf("Invalid incoming frame checksum: %x", frame)
		m.frameLock = false
	}
}

// Set the target VBus device.
func (m *mk2Ser) setTarget() {
	cmd := make([]byte, 3)
	cmd[0] = setTargetFrame
	cmd[1] = 0x01
	cmd[2] = 0x00
	m.sendCommand(cmd)
}

// Request the scaling factor for entry 'in'.
func (m *mk2Ser) reqScaleFactor(in byte) {
	cmd := make([]byte, 4)
	cmd[0] = winmonFrame
	cmd[1] = commandGetRAMVarInfo
	cmd[2] = in
	m.sendCommand(cmd)
}

func int16Abs(in int16) uint16 {
	if in < 0 {
		return uint16(-in)
	}
	return uint16(in)
}

// Decode the scale factor frame.
func (m *mk2Ser) scaleDecode(frame []byte) {
	tmp := scaling{}
	logrus.Debugf("Scale frame(%d): 0x%x", len(frame), frame)
	if len(frame) < 6 {
		tmp.supported = false
		logrus.Warnf("Skiping scaling factors for: %d", m.scaleCount)
	} else {
		tmp.supported = true
		var scl int16
		var ofs int16
		if len(frame) == 6 {
			scl = int16(frame[2])<<8 + int16(frame[1])
			ofs = int16(uint16(frame[4])<<8 + uint16(frame[3]))
		} else {
			scl = int16(frame[2])<<8 + int16(frame[1])
			ofs = int16(uint16(frame[5])<<8 + uint16(frame[4]))
		}
		if scl < 0 {
			tmp.signed = true
		}
		tmp.offset = float64(ofs)
		scale := int16Abs(scl)
		if scale >= 0x4000 {
			tmp.scale = 1 / (0x8000 - float64(scale))
		} else {
			tmp.scale = float64(scale)
		}
	}
	logrus.Debugf("scalecount %v: %#v \n", m.scaleCount, tmp)
	m.scales = append(m.scales, tmp)
	m.scaleCount++
	if m.scaleCount < ramVarMaxOffset {
		m.reqScaleFactor(byte(m.scaleCount))
	} else {
		logrus.Info("Monitoring starting.")
	}
}

// Decode the version number
func (m *mk2Ser) versionDecode(frame []byte) {
	logrus.Debugf("versiondecode %v", frame)
	m.info.Version = 0
	m.info.Valid = true
	for i := 0; i < 4; i++ {
		m.info.Version += uint32(frame[i]) << uint(i) * 8
	}

	if m.scaleCount < ramVarMaxOffset {
		logrus.Info("Get scaling factors.")
		m.reqScaleFactor(byte(m.scaleCount))
	} else {
		// Send DC status request
		cmd := make([]byte, 2)
		cmd[0] = infoReqFrame
		cmd[1] = infoReqAddrDC
		m.sendCommand(cmd)
	}
}

// Decode with correct signedness and apply scale
func (m *mk2Ser) applyScaleAndSign(data []byte, scale int) float64 {
	var value float64
	if !m.scales[scale].supported {
		return 0
	}
	if m.scales[scale].signed {
		value = getSigned(data)
	} else {
		value = getUnsigned16(data)
	}
	return m.applyScale(value, scale)
}

// Apply scaling to float
func (m *mk2Ser) applyScale(value float64, scale int) float64 {
	if !m.scales[scale].supported {
		return value
	}
	return m.scales[scale].scale * (value + m.scales[scale].offset)
}

// Convert bytes->int16->float
func getSigned(data []byte) float64 {
	return float64(int16(data[0]) + int16(data[1])<<8)
}

// Convert bytes->int16->float
func getUnsigned16(data []byte) float64 {
	return float64(uint16(data[0]) + uint16(data[1])<<8)
}

// Convert bytes->uint32->float
func getUnsigned(data []byte) float64 {
	return float64(uint32(data[0]) + uint32(data[1])<<8 + uint32(data[2])<<16)
}

// Decodes DC frame.
func (m *mk2Ser) dcDecode(frame []byte) {
	m.info.BatVoltage = m.applyScaleAndSign(frame[5:7], ramVarVBat)

	usedC := m.applyScale(getUnsigned(frame[7:10]), ramVarIBat)
	chargeC := m.applyScale(getUnsigned(frame[10:13]), ramVarIBat)
	m.info.BatCurrent = usedC - chargeC

	m.info.OutFrequency = 10 / (m.applyScale(float64(frame[13]), ramVarInverterPeriod))
	logrus.Debugf("dcDecode %#v", m.info)

	// Send L1 status request
	cmd := make([]byte, 2)
	cmd[0] = infoReqFrame
	cmd[1] = infoReqAddrACL1
	m.sendCommand(cmd)
}

// Decodes AC frame.
func (m *mk2Ser) acDecode(frame []byte) {
	m.info.InVoltage = m.applyScaleAndSign(frame[5:7], ramVarVMains)
	m.info.InCurrent = m.applyScaleAndSign(frame[7:9], ramVarIMains)
	m.info.OutVoltage = m.applyScaleAndSign(frame[9:11], ramVarVInverter)
	m.info.OutCurrent = m.applyScaleAndSign(frame[11:13], ramVarIInverter)

	if frame[13] == 0xff {
		m.info.InFrequency = 0
	} else {
		m.info.InFrequency = 10 / (m.applyScale(float64(frame[13]), ramVarMainPeriod))
	}
	logrus.Debugf("acDecode %#v", m.info)

	// Send status request
	cmd := make([]byte, 1)
	cmd[0] = ledFrame
	m.sendCommand(cmd)
}

// Decode charge state of battery.
func (m *mk2Ser) stateDecode(frame []byte) {
	m.info.ChargeState = m.applyScaleAndSign(frame[1:3], ramVarChargeState)
	logrus.Debugf("battery state decode %#v", m.info)
	m.updateReport()
}

// Decode the LED state frame.
func (m *mk2Ser) ledDecode(frame []byte) {

	m.info.LEDs = getLEDs(frame[0], frame[1])
	// Send charge state request
	cmd := make([]byte, 4)
	cmd[0] = winmonFrame
	cmd[1] = commandReadRAMVar
	cmd[2] = ramVarChargeState
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
// CHARGER ON
//  07 ff  58 37 01 00 14 81  d5
//  07 ff  5a 37 01 00 14 81  d3

// CHARGER OFF
//  07 ff  5a 37 01 00 54 81  93
//  07 ff  59 37 01 00 54 81  94
func (m *mk2Ser) sendCommand(data []byte) {
	l := len(data)
	dataOut := make([]byte, l+3)
	dataOut[0] = byte(l + 1)
	dataOut[1] = frameHeader
	cr := -dataOut[0] - dataOut[1]
	for i := 0; i < len(data); i++ {
		cr = cr - data[i]
		dataOut[i+2] = data[i]
	}
	dataOut[l+2] = cr

	logrus.Debugf("sendCommand %#v", dataOut)
	_, err := m.p.Write(dataOut)
	if err != nil {
		m.addError(fmt.Errorf("Write error: %v", err))
	}
}

func (m *mk2Ser) SendCommand(data []byte) {
	m.sendCommand(data)
}

// Checks the frame crc.
func checkChecksum(l, t byte, d []byte) bool {
	cr := (uint16(l) + uint16(t)) % 256
	for i := 0; i < len(d); i++ {
		cr = (cr + uint16(d[i])) % 256
	}
	return cr == 0
}
