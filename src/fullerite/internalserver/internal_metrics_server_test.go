package internalserver

import (
	"fullerite/config"
	"fullerite/handler"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"testing"

	l "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type testHandler struct {
	handler.BaseHandler
	metrics handler.InternalMetrics
	name    string
}

func (h testHandler) Run()                             {} // noop
func (h testHandler) Configure(map[string]interface{}) {} // noop
func (h testHandler) InternalMetrics() handler.InternalMetrics {
	return h.metrics
}
func (h testHandler) Name() string {
	return h.name
}

func buildTestHandler(name string, counters, gauges map[string]float64) handler.Handler {
	testMetrics := handler.NewInternalMetrics()
	for name, value := range counters {
		testMetrics.Counters[name] = value
	}
	for name, value := range gauges {
		testMetrics.Gauges[name] = value
	}

	h := new(testHandler)
	h.metrics = *testMetrics
	h.name = name
	return h
}

func TestServerConfigure(t *testing.T) {
	// TODO test config
}

func TestBuildResponse(t *testing.T) {
	testLog := l.WithField("testing", "internal_server")

	h := buildTestHandler(
		"somehandler",
		map[string]float64{"somecounter": 12.3},
		map[string]float64{"somegauge": 432.3},
	)
	testHandlers := []handler.Handler{h}

	srv := internalServer{
		log:      testLog,
		handlers: &testHandlers,
	}

	rsp := srv.buildResponse()
	assert.NotNil(t, rsp)

	rspFormat := ResponseFormat{}
	err := json.Unmarshal(*rsp, &rspFormat)
	assert.Nil(t, err)

	// in this test ignore the memory stats
	assert.NotNil(t, rspFormat.Memory)
	assert.Equal(t, 1, len(rspFormat.Handlers))

	realHandlerRsp := rspFormat.Handlers["somehandler"]
	assert.Equal(t, 1, len(realHandlerRsp.Counters))
	assert.Equal(t, 12.3, realHandlerRsp.Counters["somecounter"])
	assert.Equal(t, 1, len(realHandlerRsp.Gauges))
	assert.Equal(t, 432.3, realHandlerRsp.Gauges["somegauge"])
}

func TestBuildResponseMemory(t *testing.T) {
	testLog := l.WithField("testing", "internal_server")
	emptyHandlers := []handler.Handler{}

	srv := internalServer{
		log:      testLog,
		handlers: &emptyHandlers,
	}

	rspFormat := new(ResponseFormat)
	rsp := srv.buildResponse()
	err := json.Unmarshal(*rsp, rspFormat)
	assert.Nil(t, err)

	// only care about the memory part
	assert.NotNil(t, rspFormat.Memory)
	assert.NotNil(t, rspFormat.Handlers)

	// only check that there are enough items in the list
	assert.Equal(t, 7, len(rspFormat.Memory.Counters))
	assert.Equal(t, 19, len(rspFormat.Memory.Gauges))
}

func TestRespondToHttp(t *testing.T) {
	getPort := func() string {
		l, _ := net.Listen("tcp", "")
		defer l.Close()
		port := regexp.MustCompile("[^0-9]").ReplaceAll([]byte(l.Addr().String()), []byte(""))
		return string(port)
	}
	port := getPort()
	cfg := config.Config{}
	cfg.InternalServerConfig = map[string]interface{}{"port": port}

	h1 := buildTestHandler(
		"firsthandler",
		map[string]float64{"somecounter": 12.3},
		map[string]float64{"somegauge": 432.3},
	)
	h2 := buildTestHandler(
		"secondhandler",
		map[string]float64{"secondcounter": 456.2},
		map[string]float64{"secondgauge": 890.2},
	)
	testHandlers := []handler.Handler{h1, h2}

	go RunServer(&cfg, &testHandlers)

	rsp, err := http.Get(fmt.Sprintf("http://localhost:%s/metrics", port))
	assert.Nil(t, err)
	assert.Equal(t, 200, rsp.StatusCode)

	// get the body - make sure we can unmarshall it
	// then check the contents. The length of the memory parts and the
	// handlers that we put in

	txt, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	assert.Nil(t, err)

	var parsedResponse ResponseFormat
	err = json.Unmarshal(txt, &parsedResponse)
	assert.Nil(t, err)

	assert.Equal(t, 7, len(parsedResponse.Memory.Counters))
	assert.Equal(t, 19, len(parsedResponse.Memory.Gauges))

	// test that both handlers are present and have the right values
	assert.Equal(t, 2, len(parsedResponse.Handlers))

	handlerMetrics := parsedResponse.Handlers["firsthandler"]
	assert.Equal(t, 1, len(handlerMetrics.Counters))
	assert.Equal(t, 1, len(handlerMetrics.Gauges))
	assert.Equal(t, 12.3, handlerMetrics.Counters["somecounter"])
	assert.Equal(t, 432.3, handlerMetrics.Gauges["somegauge"])

	handlerMetrics = parsedResponse.Handlers["secondhandler"]
	assert.Equal(t, 1, len(handlerMetrics.Counters))
	assert.Equal(t, 1, len(handlerMetrics.Gauges))
	assert.Equal(t, 456.2, handlerMetrics.Counters["secondcounter"])
	assert.Equal(t, 890.2, handlerMetrics.Gauges["secondgauge"])
}
