package swagger

import (
	"github.com/gin-gonic/gin"
)

func UI(ctx *gin.Context) {
	ctx.Data(200, "text/html; charset=utf-8", []byte(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <title>Winx API Gateway Swagger</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: '/swagger/openapi.yaml',
      dom_id: '#swagger-ui'
    });
  </script>
</body>
</html>`))
}

func Spec(ctx *gin.Context) {
	ctx.File("swagger/openapi.yaml")
}
