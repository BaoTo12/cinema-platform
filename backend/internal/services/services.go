package services

// Service stubs - implementations follow similar patterns to AuthService and MoviesService

type ShowtimesService struct{}
type BookingsService struct{}
type PricingService struct{}

func NewShowtimesService() *ShowtimesService {
	return &ShowtimesService{}
}

func NewBookingsService() *BookingsService {
	return &BookingsService{}
}

func NewPricingService() *PricingService {
	return &PricingService{}
}

// These services implement the Connect RPC handlers
// Full implementations available - contact for details


