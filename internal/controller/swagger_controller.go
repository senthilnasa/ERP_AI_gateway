package controller

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type SwaggerController struct {
	swaggerJsonPath string
}

func NewSwaggerController(swaggerJsonPath string) *SwaggerController {
	if swaggerJsonPath == "" {
		swaggerJsonPath = "docs/swagger.json"
	}
	return &SwaggerController{
		swaggerJsonPath: swaggerJsonPath,
	}
}

func (s *SwaggerController) ServeJSON(c *gin.Context) {
	absPath, err := filepath.Abs(s.swaggerJsonPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve swagger.json path"})
		return
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "swagger.json not found"})
		return
	}
	c.Data(http.StatusOK, "application/json", data)
}

func (s *SwaggerController) ServeUI(c *gin.Context) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>ERP AI Gateway - Swagger API Docs</title>
  <link rel="stylesheet" type="text/css" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css" />
  <style>
    html { box-sizing: border-box; overflow-y: scroll; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin: 0; background: #fafafa; }
    .topbar { display: none !important; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js" charset="UTF-8"></script>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-standalone-preset.js" charset="UTF-8"></script>
  <script>
    window.onload = function() {
      const ui = SwaggerUIBundle({
        url: "/docs/swagger.json",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout"
      });
      window.ui = ui;
    };
  </script>
</body>
</html>`

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
