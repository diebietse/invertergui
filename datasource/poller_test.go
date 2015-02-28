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
	"errors"
	"testing"
	"time"
)

type mockDataSource struct {
	currentMock int
	shouldBreak bool
}

func (this *mockDataSource) GetData(data *MultiplusStatus) error {
	if this.shouldBreak {
		return errors.New("Do not be alarmed. This is only a test.")
	}
	data.BatCurrent = float64(this.currentMock)
	this.currentMock++
	return nil
}

func TestOnePoll(t *testing.T) {
	poller := NewDataPoller(&mockDataSource{currentMock: 100}, 1*time.Millisecond)
	statChan := poller.C()
	status := <-statChan
	if status.MpStatus.BatCurrent != 100 {
		t.Errorf("Incorrect data passed from data source.")
	}
	if status.Time.IsZero() {
		t.Errorf("Time not set.")
	}
	poller.Stop()
}

func TestMultiplePolls(t *testing.T) {
	poller := NewDataPoller(&mockDataSource{currentMock: 100}, 1*time.Millisecond)
	statChan := poller.C()
	expect := 100
	for i := 0; i < 100; i++ {
		status := <-statChan
		if status.MpStatus.BatCurrent != float64(expect) {
			t.Errorf("Incorrect data passed from data source.")
		}
		expect++
		if status.Time.IsZero() {
			t.Errorf("Time not set.")
		}
	}
	poller.Stop()
}

func TestError(t *testing.T) {
	poller := NewDataPoller(&mockDataSource{shouldBreak: true}, 1*time.Millisecond)
	statChan := poller.C()
	status := <-statChan
	if status.Err == nil {
		t.Errorf("Error not correctly propagated.")
	}
	poller.Stop()
}
