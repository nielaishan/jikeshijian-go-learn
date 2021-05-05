package main

import (
	"context"
	"go-learn/errgroup"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
)

func main()  {
	ctx :=	context.Background()
	errG, _ := errgroup.WithContext(ctx)

	svr := http.Server{Addr: ":8080"}

	errG.Go(func() error {
		http.HandleFunc("/api/helloworld", func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, "hello world\n")
		})
		if err := svr.ListenAndServe(); err != nil {
			log.Printf("http server end")
			return err
		}
		return nil
	})

	errG.Go(func() error {
		c := make(chan os.Signal)
		signal.Notify(c)
		for {
			select {
			case <-c:
				return svr.Shutdown(ctx)
			}
		}
	})
	errG.Wait()
}