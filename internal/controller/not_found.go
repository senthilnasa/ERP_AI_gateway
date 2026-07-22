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

const FaviconSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64" width="64" height="64">
  <defs>
    <linearGradient id="grad" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:#2563eb;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#7c3aed;stop-opacity:1" />
    </linearGradient>
  </defs>
  <rect width="64" height="64" rx="16" fill="url(#grad)"/>
  <path d="M32 16 L44 24 L44 40 L32 48 L20 40 L20 24 Z" fill="none" stroke="#ffffff" stroke-width="3.5" stroke-linejoin="round" stroke-linecap="round"/>
  <circle cx="32" cy="32" r="6" fill="#ffffff"/>
  <circle cx="32" cy="16" r="3.5" fill="#60a5fa"/>
  <circle cx="44" cy="24" r="3.5" fill="#60a5fa"/>
  <circle cx="44" cy="40" r="3.5" fill="#60a5fa"/>
  <circle cx="32" cy="48" r="3.5" fill="#60a5fa"/>
  <circle cx="20" cy="40" r="3.5" fill="#60a5fa"/>
  <circle cx="20" cy="24" r="3.5" fill="#60a5fa"/>
</svg>`

func HandleFavicon(c *gin.Context) {
	c.Header("Cache-Control", "public, max-age=86400")
	c.Data(http.StatusOK, "image/svg+xml", []byte(FaviconSVG))
}
