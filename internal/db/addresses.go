package db

import (
	"context"
	"site-backend-go/internal/exceptions"
)

func (s *Storage) FindStreetsAlike(ctx context.Context, streetInput string) ([]StreetAlikeModel, error) {
	query := `
		SELECT DISTINCT
			a.street
		FROM addresses a
		WHERE
			a.street ILIKE $1;
	`
	formattedInput := "%" + streetInput + "%"

	rows, err := s.db.QueryContext(ctx, query, formattedInput)
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}
	defer rows.Close()

	streets := make([]StreetAlikeModel, 0)
	for rows.Next() {
		var street StreetAlikeModel
		err := rows.Scan(&street.Name)
		if err != nil {
			s.log.Error(err.Error())
			return nil, exceptions.ErrDBError
		}
		streets = append(streets, street)
	}
	if err := rows.Err(); err != nil {
		s.log.Error(err.Error())
		return nil, exceptions.ErrDBError
	}

	if len(streets) == 0 {
		return nil, exceptions.ErrNotFound
	}

	return streets, nil
}

func (s *Storage) FindSuitableHousesAlike(
	ctx context.Context, selectedStreet string, houseInput string,
) ([]HouseAlikeModel, error) {
	query := `
		SELECT DISTINCT
			a.house
		FROM addresses a
		WHERE
			a.street ILIKE $1
			AND a.house ILIKE $2
		ORDER BY a.house;
	`
	formattedInputStreet := "%" + selectedStreet + "%"
	formattedInputHouse := "%" + houseInput + "%"

	rows, err := s.db.QueryContext(ctx, query, formattedInputStreet, formattedInputHouse)
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}
	defer rows.Close()

	houses := make([]HouseAlikeModel, 0)
	for rows.Next() {
		var house HouseAlikeModel
		err := rows.Scan(&house.Number)
		if err != nil {
			s.log.Error(err.Error())
			return nil, exceptions.ErrDBError
		}
		houses = append(houses, house)
	}
	if err := rows.Err(); err != nil {
		s.log.Error(err.Error())
		return nil, exceptions.ErrDBError
	}

	if len(houses) == 0 {
		return nil, exceptions.ErrNotFound
	}

	return houses, nil
}

func (s *Storage) FindSuitableAddressesAlike(
	ctx context.Context, selectedStreet string, selectedHouse string, entranceInput string,
) ([]AddressAlikeModel, error) {
	query := `
		SELECT DISTINCT
		    CONCAT(
				'ул. ', a.street, ','
				, ' д. ', a.house, ','
				, ' под. ', a.entrance
			) AS full_address
		FROM addresses a
		WHERE
			a.street ILIKE $1
			AND a.house = $2
		    -- AND a.house ILIKE $2
			AND a.entrance ILIKE $3
	`
	formattedInputStreet := "%" + selectedStreet + "%"
	// formattedInputHouse := "%" + selectedHouse + "%"
	formattedInputHouse := selectedHouse
	formattedInputEntrance := "%" + entranceInput + "%"

	rows, err := s.db.QueryContext(ctx, query, formattedInputStreet, formattedInputHouse, formattedInputEntrance)
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}
	defer rows.Close()

	entrances := make([]AddressAlikeModel, 0)
	for rows.Next() {
		var entrance AddressAlikeModel
		err := rows.Scan(&entrance.FullAddress)
		if err != nil {
			s.log.Error(err.Error())
			return nil, exceptions.ErrDBError
		}
		entrances = append(entrances, entrance)
	}
	if err := rows.Err(); err != nil {
		s.log.Error(err.Error())
		return nil, exceptions.ErrDBError
	}

	if len(entrances) == 0 {
		return nil, exceptions.ErrNotFound
	}

	return entrances, nil
}
