package main

import (
	"log"
	"net/http"

	"golang.org/x/sync/errgroup"

	"lft/download"
	"lft/global"
	_ "lft/statik"
	"lft/upload"
)

func main() {
	var g errgroup.Group

	g.Go(func() error {
		err := upload.StartServer(global.UploadPort)
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
		return err
	})

	g.Go(func() error {
		err := download.StartServer(global.DownloadPort)
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
		return err
	})

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
