package webui

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

//go:embed dist/*
var dist embed.FS

// Serve registers the frontend SPA on the gin engine.
// In Docker (production), the embedded dist is used.
// In development, the Vite dev server at localhost:5173 proxies requests.
func Serve(r *gin.Engine) {
	// If a local dist directory exists (dev with pre-built frontend), use it
	if _, err := os.Stat("internal/webui/dist/index.html"); err == nil {
		r.Static("/assets", "internal/webui/dist/assets")
		r.StaticFile("/", "internal/webui/dist/index.html")
		r.NoRoute(func(c *gin.Context) {
			c.File("internal/webui/dist/index.html")
		})
		return
	}

	// Fallback: use embedded dist (Docker build)
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		log.Printf("[webui] embedded dist not available: %v", err)
		return
	}
	fileServer := http.FileServer(http.FS(sub))
	r.GET("/assets/*filepath", func(c *gin.Context) {
		c.Request.URL.Path = "/assets" + c.Param("filepath")
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.Request.URL.Path = "/favicon.ico"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
	r.NoRoute(func(c *gin.Context) {
		// SPA fallback: serve index.html for any non-API route
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}
