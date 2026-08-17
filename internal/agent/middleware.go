package agent

import (
	"bytes"
	"compress/gzip"

	"github.com/go-resty/resty/v2"
)

func GzipRequest() resty.RequestMiddleware {
	return func(c *resty.Client, r *resty.Request) error {
		var buf bytes.Buffer

		zw := gzip.NewWriter(&buf)
		body, ok := r.Body.([]byte)
		if !ok || len(body) == 0 {
			return nil
		}

		if _, err := zw.Write(body); err != nil {
			return err
		}

		if err := zw.Close(); err != nil {
			return err
		}

		r.SetBody(buf.Bytes()).SetHeader("Content-Encoding", "gzip")

		return nil
	}
}
