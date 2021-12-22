package mk2core

import (
	"github.com/diebietse/invertergui/mk2driver"
)

type Core struct {
	mk2driver.Mk2
	plugins  map[*subscription]bool
	register chan *subscription
}

func NewCore(m mk2driver.Mk2) *Core {
	core := &Core{
		Mk2:      m,
		register: make(chan *subscription, 255),
		plugins:  map[*subscription]bool{},
	}
	go core.run()
	return core
}

func (c *Core) NewSubscription() mk2driver.Mk2 {
	sub := &subscription{
		send: make(chan *mk2driver.Mk2Info),
	}
	c.register <- sub
	return sub
}

func (c *Core) run() {
	for {
		select {
		case r := <-c.register:
			c.plugins[r] = true
		case e := <-c.C():
			for plugin := range c.plugins {
				select {
				case plugin.send <- e:
				default:
				}
			}
		}
	}
}

type subscription struct {
	send chan *mk2driver.Mk2Info
}

func (s *subscription) C() chan *mk2driver.Mk2Info {
	return s.send
}


func (s *subscription) SendCommand(data []byte) {

}


func (s *subscription) Close() {
	close(s.send)
}
