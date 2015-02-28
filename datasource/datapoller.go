/*
Copyright (c) 2015, Hendrik van Wyk
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

* Redistributions of source code must retain the above copyright notice, this
list of conditions and the following disclaimer.

* Redistributions in binary form must reproduce the above copyright notice,
this list of conditions and the following disclaimer in the documentation
and/or other materials provided with the distribution.

* Neither the name of invertergui nor the names of its
contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

package datasource

import (
	"sync"
	"time"
)

type DataPoller interface {
	C() chan *Status
	Stop()
}

type Status struct {
	MpStatus MultiplusStatus
	Time     time.Time
	Err      error
}

type poller struct {
	source     DataSource
	rate       time.Duration
	statusChan chan *Status
	stop       chan struct{}
	wg         sync.WaitGroup
}

func NewDataPoller(source DataSource, pollRate time.Duration) DataPoller {
	this := &poller{
		source:     source,
		rate:       pollRate,
		statusChan: make(chan *Status),
		stop:       make(chan struct{}),
	}
	this.wg.Add(1)
	go this.poll()
	return this
}

func (this *poller) C() chan *Status {
	return this.statusChan
}

func (this *poller) Stop() {
	close(this.stop)
	this.wg.Wait()
}

func (this *poller) poll() {
	ticker := time.NewTicker(this.rate)
	this.doPoll()
	for {
		select {
		case <-ticker.C:
			this.doPoll()
		case <-this.stop:
			ticker.Stop()
			close(this.statusChan)
			this.wg.Done()
			return
		}
	}
}

func (this *poller) doPoll() {
	tmp := new(Status)
	tmp.Err = this.source.GetData(&tmp.MpStatus)
	tmp.Time = time.Now()
	this.statusChan <- tmp
}
