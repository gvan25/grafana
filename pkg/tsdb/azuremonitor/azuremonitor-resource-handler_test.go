package azuremonitor

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_parseResourcePath(t *testing.T) {
	tests := []struct {
		name           string
		original       string
		expectedDS     string
		expectedTarget string
		Err            require.ErrorAssertionFunc
	}{
		{
			"Path with a subscription",
			"/azuremonitor/subscriptions/44693801",
			azureMonitor,
			"/subscriptions/44693801",
			require.NoError,
		},
		{
			"Malformed path",
			"/subscriptions/44693801",
			"",
			"",
			require.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsName, target, err := parseResourcePath(tt.original)
			if dsName != tt.expectedDS {
				t.Errorf("Unexpected ds name %s expecting %s", dsName, tt.expectedDS)
			}
			if target != tt.expectedTarget {
				t.Errorf("Unexpected target %s expecting %s", target, tt.expectedTarget)
			}
			tt.Err(t, err)
		})
	}
}

func Test_proxyRequest(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"forwards headers and body"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("foo", "bar")
				_, err := w.Write([]byte("result"))
				if err != nil {
					t.Fatal(err)
				}
			}))
			req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
			if err != nil {
				t.Error(err)
			}
			rw := httptest.NewRecorder()
			proxy := httpServiceProxy{}
			res := proxy.Do(rw, req, srv.Client())
			if res.Header().Get("foo") != "bar" {
				t.Errorf("Unexpected headers: %v", res.Header())
			}
			result := rw.Result()
			body, err := ioutil.ReadAll(result.Body)
			if err != nil {
				t.Error(err)
			}
			err = result.Body.Close()
			if err != nil {
				t.Error(err)
			}
			if string(body) != "result" {
				t.Errorf("Unexpected body: %v", string(body))
			}
		})
	}
}

type fakeProxy struct {
	requestedURL string
}

func (s *fakeProxy) Do(rw http.ResponseWriter, req *http.Request, cli *http.Client) http.ResponseWriter {
	s.requestedURL = req.URL.String()
	return nil
}

func Test_resourceHandler(t *testing.T) {
	proxy := &fakeProxy{}
	s := Service{
		Services: map[string]datasourceService{
			azureMonitor: {
				URL:        "https://test",
				HTTPClient: &http.Client{},
			},
		},
		proxy: proxy,
		im:    &fakeInstance{},
	}
	rw := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "http://foo/azuremonitor/subscriptions/44693801", nil)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	s.resourceHandler(rw, req)
	expectedURL := "https://test/subscriptions/44693801"
	if proxy.requestedURL != expectedURL {
		t.Errorf("Unexpected result URL. Got %s, expecting %s", proxy.requestedURL, expectedURL)
	}
}
