package service

import (
	"context"
	"site-backend-go/internal/db"
)

type AddressesStorage interface {
	FindStreetsAlike(ctx context.Context, streetInput string) ([]db.StreetAlikeModel, error)
	FindSuitableHousesAlike(ctx context.Context, selectedStreet string, houseInput string) ([]db.HouseAlikeModel, error)
	FindSuitableAddressesAlike(ctx context.Context, selectedStreet string, selectedHouse string, entranceInput string) ([]db.AddressAlikeModel, error)
}

type AddressesService struct {
	storage AddressesStorage
}

func NewAddressesService(storage AddressesStorage) *AddressesService {
	return &AddressesService{
		storage,
	}

}

func (s *AddressesService) GetStreetsAlike(ctx context.Context, addressInput string) ([]db.StreetAlikeModel, error) {
	streetsFound, err := s.storage.FindStreetsAlike(ctx, addressInput)
	if err != nil {
		return nil, err
	}

	return streetsFound, nil
}

func (s *AddressesService) GetHousesAlike(
	ctx context.Context, selectedStreet string, addressInput string,
) ([]db.HouseAlikeModel, error) {
	housesFound, err := s.storage.FindSuitableHousesAlike(ctx, selectedStreet, addressInput)
	if err != nil {
		return nil, err
	}

	return housesFound, nil
}

func (s *AddressesService) GetAddressesAlike(
	ctx context.Context, selectedStreet string, selectedHouse string, entranceInput string,
) ([]db.AddressAlikeModel, error) {
	entrancesFound, err := s.storage.FindSuitableAddressesAlike(ctx, selectedStreet, selectedHouse, entranceInput)
	if err != nil {
		return nil, err
	}

	return entrancesFound, nil
}
