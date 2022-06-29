package main

import (
	"context"
	"fmt"
	"github.com/cecobask/ocd-tracker-api/db/postgres"
	"log"
	"net/http"
	"os"
)

func main() {
	ctx := context.Background()
	pg, err := postgres.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer pg.Connection.Close(context.Background())
	server := &http.Server{
		Addr: fmt.Sprintf("0.0.0.0:%s", os.Getenv("SERVER_PORT")),
	}
	log.Println("starting http server at", server.Addr)
	log.Fatal(server.ListenAndServe())
}
