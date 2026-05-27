package gapi

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func unauthenticatedError(err error) error {
	return status.Error(codes.Unauthenticated, fmt.Sprintf("unauthenticated: %s", err))
}
