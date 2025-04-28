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

		var packages []map[string]any
		for {
			pkg, err := resp.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}

			packages = append(packages, map[string]any{"name": pkg.Name})
		}

		c.JSON(http.StatusOK, packages)
	})

	r.Run()
}
