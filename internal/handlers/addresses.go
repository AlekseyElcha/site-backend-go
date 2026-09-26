package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"site-backend-go/internal/dtos"
	"site-backend-go/internal/exceptions"
)

func (s *Server) FindSuitableStreetsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	targetStreetName := r.URL.Query().Get("name")

	streets, err := s.AddressesService.GetStreetsAlike(ctx, targetStreetName)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if errors.Is(err, exceptions.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message": "Подходящие улицы не найдены."}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
		return
	}

	names := make([]string, 0, len(streets))
	for _, street := range streets {
		names = append(names, street.Name)
	}

	response := dtos.StreetsAlikeResponse{
		Streets: names,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) FindSuitableHousesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	selectedStreetName := r.URL.Query().Get("street")
	targetHouseName := r.URL.Query().Get("number")

	houses, err := s.AddressesService.GetHousesAlike(ctx, selectedStreetName, targetHouseName)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if errors.Is(err, exceptions.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message": "Подходящие дома не найдены."}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
		return
	}

	names := make([]string, 0, len(houses))
	for _, house := range houses {
		names = append(names, house.Number)
	}

	response := dtos.HousesAlikeResponse{
		Houses: names,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) FindSuitableAddressesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	selectedStreetName := r.URL.Query().Get("street")
	selectedHouseName := r.URL.Query().Get("house")
	targetEntrance := r.URL.Query().Get("entrance")

	addresses, err := s.AddressesService.GetAddressesAlike(ctx, selectedStreetName, selectedHouseName, targetEntrance)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if errors.Is(err, exceptions.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message": "Подходящие адреса не найдены."}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
		return
	}

	fullAddresses := make([]string, 0, len(addresses))
	for _, address := range addresses {
		fullAddresses = append(fullAddresses, address.FullAddress)
	}

	response := dtos.AddressesAlikeResponse{
		FullAddresses: fullAddresses,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
		return
	}
	w.WriteHeader(http.StatusOK)
}
