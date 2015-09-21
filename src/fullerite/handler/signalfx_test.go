package handler

import (
	"fullerite/metric"

	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	l "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func getTestSignalfxHandler(interval, buffsize, timeoutsec int) *SignalFx {
	testChannel := make(chan metric.Metric)
	testLog := l.WithField("testing", "signalfx_handler")
	timeout := time.Duration(timeoutsec) * time.Second

	return NewSignalFx(testChannel, interval, buffsize, timeout, testLog)
}

func TestSignalfxConfigureEmptyConfig(t *testing.T) {
	config := make(map[string]interface{})

	s := getTestSignalfxHandler(12, 13, 14)
	s.Configure(config)

	assert.Equal(t, 12, s.Interval())
	assert.Equal(t, 13, s.MaxBufferSize())
}

func TestSignalfxConfigure(t *testing.T) {
	config := map[string]interface{}{
		"interval":        "10",
		"timeout":         "10",
		"max_buffer_size": "100",
		"authToken":       "secret",
		"endpoint":        "signalfx.server",
	}

	s := getTestSignalfxHandler(40, 50, 60)
	s.Configure(config)

	assert.Equal(t, 10, s.Interval())
	assert.Equal(t, 100, s.MaxBufferSize())
	assert.Equal(t, "signalfx.server", s.Endpoint())
}

func TestSignalFxRun(t *testing.T) {
	assert := assert.New(t)

	wait := make(chan bool)
	// Mock SignalFx server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		message := &DataPointUploadMessage{}
		proto.Unmarshal(body, message)
		datapoint := message.Datapoints[0]

		assert.Nil(err)
		assert.Equal(*datapoint.Metric, "Test")
		assert.Equal(r.Header["Content-Type"], []string{"application/x-protobuf"})
		assert.Equal(r.Header["X-Sf-Token"], []string{"secret"})

		wait <- true
	}))
	defer ts.Close()

	config := map[string]interface{}{
		"interval":        "1",
		"timeout":         "1",
		"max_buffer_size": "1",
		"authToken":       "secret",
		"endpoint":        ts.URL,
	}

	s := getTestSignalfxHandler(12, 12, 12)
	s.Configure(config)

	go s.Run()

	m := metric.New("Test")
	s.Channel() <- m

	<-wait
}
