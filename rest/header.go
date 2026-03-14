package rest

import (
	"net/http"
	"slices"
)

func CopyHeader(dst *http.Header, src http.Header) {
	for k, headers := range src {
		for _, header := range headers {
			alreadyStored := dst.Values(k)
			if slices.Contains(alreadyStored, header) {
				continue
			}

			dst.Add(k, header)
		}
	}
}
