package main

import (
	"org-service/handler"
	pb "org-service/proto"

	"github.com/micro/micro/v3/service"
	"github.com/micro/micro/v3/service/logger"
)

func main() {
	// Create service
	srv := service.New(
		service.Name("org-service"),
		service.Version("0.1"),
	)

	// Register handler
	pb.RegisterOrgServiceHandler(srv.Server(), handler.NewOrgService())

	// Run service
	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
}
