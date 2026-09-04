package agent

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

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

func HashBody(key string) resty.RequestMiddleware {
	return func(c *resty.Client, r *resty.Request) error {
		if key == "" {
			return nil
		}

		h := hmac.New(sha256.New, []byte(key))

		body, ok := r.Body.([]byte)
		if !ok || len(body) == 0 {
			return nil
		}

		h.Write(body)

		r.SetHeader("HashSHA256", hex.EncodeToString(h.Sum(nil)))

		return nil
	}
}
