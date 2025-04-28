package main

import (
	"flag"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/operator-framework/operator-registry/pkg/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var upstream = flag.String("upstream", "", "upstream grpc address")

func main() {
	flag.Parse()

	conn, err := grpc.NewClient(*upstream, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	client := api.NewRegistryClient(conn)

	r := gin.Default()

	r.GET("/package", func(c *gin.Context) {
		resp, err := client.ListPackages(c.Request.Context(), &api.ListPackageRequest{})
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		var packages []*api.PackageName
		for {
			p, err := resp.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}

			packages = append(packages, p)
		}

		c.JSON(http.StatusOK, packages)
	})

	r.GET("/package/:package", func(c *gin.Context) {
		resp, err := client.GetPackage(c.Request.Context(), &api.GetPackageRequest{Name: c.Params.ByName("package")})
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(http.StatusOK, resp)
	})

	r.GET("/bundle/:package/:channel/:csv", func(c *gin.Context) {
		resp, err := client.GetBundle(c.Request.Context(), &api.GetBundleRequest{
			PkgName:     c.Params.ByName("package"),
			ChannelName: c.Params.ByName("channel"),
			CsvName:     c.Params.ByName("csv"),
		})
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(http.StatusOK, resp)
	})

	r.Run()
}
