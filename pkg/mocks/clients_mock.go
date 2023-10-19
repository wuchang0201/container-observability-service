package mocks

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"

	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
)

type MockHttpClient struct {
	Expects map[string]string // except request and response
	URL     string            // connect url
}

func (c *MockHttpClient) RoundTrip(req *http.Request) (*http.Response, error) {
	recorder := httptest.NewRecorder()
	if c.Expects == nil || len(c.Expects) == 0 {
		return nil, errors.New("ResponseBody or RequestBody can't be null")
	}

	recorder.Header().Set("Content-Type", "application/json")
	var reqContent []byte
	var err error
	if req.Body != nil {
		reqContent, err = ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
	} else {
		reqContent = []byte(req.URL.String())
	}

	for requestStr, responseStr := range c.Expects {
		reqReg := regexp.MustCompile(requestStr)
		regRes := reqReg.Find(reqContent)
		if string(regRes) == "" {
			continue
		}
		if _, err = recorder.Write([]byte(responseStr)); err != nil {
			return nil, err
		}
		return recorder.Result(), nil
	}

	return nil, errors.Errorf(`Except query string: "%+v" But get "%s"`, c.Expects, reqContent)
}

func (c *MockHttpClient) MockElasticSearchClient() (*elastic.Client, error) {
	client, err := elastic.NewClient(
		elastic.SetURL(c.URL),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
		elastic.SetHttpClient(&http.Client{Transport: c}))

	return client, err
}

func (c *MockHttpClient) MockHttpClient() (*http.Client, error) {
	client := &http.Client{Transport: c}
	return client, nil
}

func NewMock() (httpMock *MockHttpClient) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer testServer.Close()
	var mock = &MockHttpClient{
		URL: testServer.URL,
	}
	return mock
}

func (c *MockHttpClient) ShouldRequestResponse(expects map[string]string) *MockHttpClient {
	c.Expects = expects
	return c
}
