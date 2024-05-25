package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

const timeOut = 5

const apiTimeOut = 150000

func Init(env string) {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	ap := path.Join(basepath, "../../config", env)

	if err := godotenv.Load(ap); err != nil {
		log.Fatalf("%s", err)
	}

	r := NewRouter(env)

	srv := &http.Server{
		Addr:    os.Getenv("SERVER_PORT"),
		Handler: http.TimeoutHandler(r, apiTimeOut*time.Millisecond, "Timeout!\n"),
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server...")

	ctx, cancel := context.WithTimeout(context.Background(), timeOut*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown: ", err)
	}

	log.Println("Server exiting")
}
