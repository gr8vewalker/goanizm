package extractors

import "net/http"

type Video struct {
	Link    string
	Name    string
	Headers http.Header
	Audio   string
}
