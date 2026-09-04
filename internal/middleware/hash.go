package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

const HashHeader = "HashSHA256"

func HashRequest(key string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if key == "" {
			ctx.Next()
			return
		}

		hash := ctx.GetHeader("HashSHA256")
		if hash == "" {
			ctx.Next()
			return
		}

		decodedHash, err := hex.DecodeString(hash)
		if err != nil {
			ctx.Abort()
			ctx.String(http.StatusBadRequest, "некорректный формат хеша")
			return
		}

		body, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			ctx.Abort()
			ctx.String(http.StatusInternalServerError, "не удалось прочитать тело запроса")
			return
		}

		ctx.Request.Body = io.NopCloser(bytes.NewReader(body))

		h := hmac.New(sha256.New, []byte(key))
		h.Write(body)

		if !hmac.Equal(decodedHash, h.Sum(nil)) {
			ctx.Abort()
			ctx.String(http.StatusBadRequest, "хеш не совпадает")
			return
		}

		ctx.Next()
	}
}
