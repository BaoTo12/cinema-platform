package services

// Note: This file creates placeholder handler functions
// In a real Connect RPC setup, these would be auto-generated from protobuf
// For now, we're creating manual stubs to make the server compile

import (
	"connectrpc.com/connect"
)

// Handler registration functions (stubs for compilation)
func NewAuthServiceHandler(svc *AuthService, opts ...connect.HandlerOption) (string, any) {
	return "/cinema.v1.AuthService/", svc
}

func NewMoviesServiceHandler(svc *MoviesService, opts ...connect.HandlerOption) (string, any) {
	return "/cinema.v1.MoviesService/", svc
}

func NewShowtimesServiceHandler(svc *ShowtimesService, opts ...connect.HandlerOption) (string, any) {
	return "/cinema.v1.ShowtimesService/", svc
}

func NewBookingsServiceHandler(svc *BookingsService, opts ...connect.HandlerOption) (string, any) {
	return "/cinema.v1.BookingsService/", svc
}

func NewPricingServiceHandler(svc *PricingService, opts ...connect.HandlerOption) (string, any) {
	return "/cinema.v1.PricingService/", svc
}
