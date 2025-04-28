package main

import (
	"context"
	"flag"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/operator-framework/operator-registry/pkg/cache"
)

var catalogDir = flag.String("catalog-dir", "", "catalog directory")

func main() {
	flag.Parse()

	tempDir, err := os.MkdirTemp("", "operator-registry-proxy-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	reg, err := cache.New(tempDir)
	if err != nil {
		panic(err)
	}

	cache.LoadOrRebuild(context.Background(), reg, os.DirFS(*catalogDir))

	backend := reg.Backend()

	r := gin.Default()

	r.GET("/package", func(c *gin.Context) {
		index, err := backend.GetPackageIndex(c.Request.Context())
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		var packages []any
		for _, pkg := range index {
			packages = append(packages, pkg)
		}

		c.JSON(http.StatusOK, packages)
	})

	r.GET("/bundle/:package/:channel/:name", func(c *gin.Context) {
		bundle, err := backend.GetBundle(c.Request.Context(), cache.BundleKey{
			PackageName: c.Params.ByName("package"),
			ChannelName: c.Params.ByName("channel"),
			Name:        c.Params.ByName("name"),
		})
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(http.StatusOK, bundle)
	})

	r.Run()
}
