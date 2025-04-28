package main

import (
	"context"
	"flag"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/operator-framework/operator-registry/pkg/api"
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

	r := gin.Default()

	r.GET("/package", func(c *gin.Context) {
		resp, err := reg.ListPackages(c.Request.Context())
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		var packages []*api.PackageName
		for _, name := range resp {
			packages = append(packages, &api.PackageName{Name: name})
		}

		c.JSON(http.StatusOK, packages)
	})

	r.GET("/package/:package", func(c *gin.Context) {
		resp, err := reg.GetPackage(c.Request.Context(), c.Params.ByName("package"))
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(http.StatusOK, resp)
	})

	r.GET("/bundle/:package/:channel/:csv", func(c *gin.Context) {
		resp, err := reg.GetBundle(c.Request.Context(),
			c.Params.ByName("package"),
			c.Params.ByName("channel"),
			c.Params.ByName("csv"),
		)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(http.StatusOK, resp)
	})

	r.Run()
}
