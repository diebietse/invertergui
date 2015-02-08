package webgui

import (
	"time"
)

type ChargeTracker struct {
	fullLevel    float64
	currentLevel float64
	lastUpdate   time.Time
}

// Creates a new charge tracker. fullLevel is in A h.
func NewChargeTracker(fullLevel float64) *ChargeTracker {
	return &ChargeTracker{
		fullLevel:    fullLevel,
		currentLevel: fullLevel, // Have to start somewhere.
		lastUpdate:   time.Now(),
	}
}

func (c *ChargeTracker) Update(amp float64) {
	newNow := time.Now()
	elapsed := newNow.Sub(c.lastUpdate).Hours()
	c.lastUpdate = newNow
	c.currentLevel -= elapsed * amp
}

func (c *ChargeTracker) CurrentLevel() float64 {
	return c.currentLevel
}

func (c *ChargeTracker) Reset() {
	c.currentLevel = c.fullLevel
}
