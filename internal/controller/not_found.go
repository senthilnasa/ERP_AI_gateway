package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/senthilnasa/ERP_AI_gateway/internal/models"
	"github.com/gin-gonic/gin"
)

const Generic404HTML = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>404 - Page Not Found</title>
    <link
      href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;700&display=swap"
      rel="stylesheet"
    />
    <style>
      body, html {
        height: 100%;
        margin: 0;
        padding: 0;
        background: #f8fafc;
        font-family: 'Inter', system-ui, -apple-system, sans-serif;
        display: flex;
        align-items: center;
        justify-content: center;
        text-align: center;
        color: #1e293b;
      }
      .container {
        max-width: 520px;
        padding: 40px 24px;
        background: #ffffff;
        border-radius: 16px;
        box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.05), 0 8px 10px -6px rgba(0, 0, 0, 0.01);
        border: 1px solid #e2e8f0;
      }
      h1 {
        font-size: 6rem;
        font-weight: 700;
        color: #2563eb;
        margin: 0 0 8px 0;
        line-height: 1;
      }
      p.subtitle {
        font-size: 1.25rem;
        font-weight: 600;
        margin-bottom: 12px;
        color: #0f172a;
      }
      p.description {
        font-size: 0.95rem;
        margin-bottom: 28px;
        color: #64748b;
        line-height: 1.5;
      }
      .actions {
        display: flex;
        gap: 12px;
        justify-content: center;
      }
      .btn {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        text-decoration: none;
        background-color: #2563eb;
        color: #ffffff;
        padding: 10px 20px;
        border-radius: 8px;
        font-weight: 500;
        font-size: 0.9rem;
        transition: background-color 0.2s ease;
      }
      .btn:hover {
        background-color: #1d4ed8;
      }
      .btn-secondary {
        background-color: #f1f5f9;
        color: #334155;
        border: 1px solid #cbd5e1;
      }
      .btn-secondary:hover {
        background-color: #e2e8f0;
      }
    </style>
  </head>
  <body>
    <div class="container">
      <h1>404</h1>
      <p class="subtitle">Page Not Found</p>
      <p class="description">Oops! The page you are looking for doesn't exist or has been moved.</p>
      <div class="actions">
        <a href="/docs" class="btn">View API Docs</a>
        <a href="/health" class="btn btn-secondary">Check Health</a>
      </div>
    </div>
  </body>
</html>`

func HandleNotFound(c *gin.Context) {
	if strings.HasPrefix(c.Request.URL.Path, "/api/") {
		reqID, _ := c.Get("RequestID")
		c.JSON(http.StatusNotFound, models.ApiResponse{
			Success: false,
			Error: models.ApiErrorDetail{
				Code:      "NOT_FOUND",
				Message:   "API endpoint not found",
				RequestID: fmt.Sprintf("%v", reqID),
			},
		})
		return
	}

	c.Data(http.StatusNotFound, "text/html; charset=utf-8", []byte(Generic404HTML))
}
