package transport

import "github.com/gin-gonic/gin"

func ReadBody(ctx *gin.Context) ([]byte, error) {
	return ctx.GetRawData()
}

func WriteJSONError(ctx *gin.Context, status int, message string) {
	ctx.JSON(status, map[string]any{
		"success": false,
		"message": message,
		"data":    nil,
	})
}

func WriteProxyResponse(ctx *gin.Context, status int, contentType string, body []byte) {
	if contentType == "" {
		contentType = "application/json"
	}

	ctx.Data(status, contentType, body)
}

func ForwardHeaders(ctx *gin.Context, names ...string) map[string]string {
	headers := make(map[string]string, len(names))
	for _, name := range names {
		value := ctx.GetHeader(name)
		if value != "" {
			headers[name] = value
		}
	}

	return headers
}
