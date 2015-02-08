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
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var sampleJSON string = `{"outCurrent": 1.19,
"leds": [0, 0, 0, 0, 1, 0, 0, 1],
"batVoltage": 26.63,
"inCurrent": 1.39,
"outVoltage": 235.3,
"inVoltage": 235.3,
"inFreq": 51.3,
"batCurrent": 0.0,
"outFreq": 735.3}`

func returnJson(resp http.ResponseWriter, req *http.Request) {
	resp.Write([]byte(sampleJSON))
}

func TestFetchStatus(t *testing.T) {
	//setup test server
	testServer := httptest.NewServer(http.HandlerFunc(returnJson))

	var status MultiplusStatus
	source := NewJSONSource(testServer.URL)
	err := source.GetData(&status)
	if err != nil {
		t.Errorf("Unmarshal gave: %v", err)
	}
	expected := MultiplusStatus{1.19, 1.39, 235.3, 235.3, 26.63, 0, 51.3, 735.3, []int{0, 0, 0, 0, 1, 0, 0, 1}}
	if !reflect.DeepEqual(status, expected) {
		t.Errorf("JSON string did not decode as expected.")
	}
	testServer.Close()
}
