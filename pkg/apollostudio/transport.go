package apollostudio

import "net/http"

type headerTransport struct {
	APIKey string
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("x-api-key", t.APIKey)
	return http.DefaultTransport.RoundTrip(req)
}
