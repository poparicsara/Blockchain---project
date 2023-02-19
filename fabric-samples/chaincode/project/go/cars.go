/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing a car
type SmartContract struct {
	contractapi.Contract
}

type Owner struct {
	Id      int     `json:"id"`
	Name    string  `json:"name"`
	Surname string  `json:"surname"`
	Email   string  `json:"email"`
	Money   float64 `json:"money"`
}

type Malfunction struct {
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

type Car struct {
	Id           int           `json:"id"`
	Make         string        `json:"make"`
	Model        string        `json:"model"`
	Color        string        `json:"color"`
	Owner        string        `json:"owner"`
	Malfunctions []Malfunction `json:"malfunctions"`
	Price        float64       `json:"price"`
}

type QueryResult struct {
	Key    string `json:"Key"`
	Record *Car
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	cars := []Car{
		{Id: 1, Make: "Toyota", Model: "Prius", Color: "blue", Owner: "1",
			Malfunctions: []Malfunction{
				{Description: "Broken brake", Price: 200},
			},
			Price: 5000,
		},
		{Id: 2, Make: "Ford", Model: "Mustang", Color: "blue", Owner: "3", Malfunctions: []Malfunction{}, Price: 3000},
		{Id: 3, Make: "Hyundai", Model: "Tucson", Color: "green", Owner: "2",
			Malfunctions: []Malfunction{
				{Description: "Broken window", Price: 1000},
			},
			Price: 2500,
		},
		{Id: 4, Make: "Volkswagen", Model: "Passat", Color: "blue", Owner: "2", Malfunctions: []Malfunction{}, Price: 7000},
		{Id: 5, Make: "Tesla", Model: "S", Color: "blue", Owner: "3", Malfunctions: []Malfunction{}, Price: 20000},
		{Id: 6, Make: "Peugeot", Model: "205", Color: "black", Owner: "3", Malfunctions: []Malfunction{}, Price: 2000},
	}

	owners := []Owner{
		{Id: 1, Name: "Sara", Surname: "Poparic", Email: "sarapoparic@gmail.com", Money: 10000},
		{Id: 2, Name: "Mila", Surname: "Poparic", Email: "milapoparic@gmail.com", Money: 5000},
		{Id: 3, Name: "Nikola", Surname: "Nikolic", Email: "nikolanikolic@gmail.com", Money: 5000},
	}

	for _, car := range cars {
		carAsBytes, err := json.Marshal(car)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(strconv.Itoa(car.Id), carAsBytes)
		if err != nil {
			return fmt.Errorf("failed to put car to world state. %v", err)
		}

		indexName := "color~owner~id"
		colorOwnerIndexKey, err := ctx.GetStub().CreateCompositeKey(indexName, []string{car.Color, car.Owner, strconv.Itoa(car.Id)})
		if err != nil {
			return err
		}

		value := []byte{0x00}
		err = ctx.GetStub().PutState(colorOwnerIndexKey, value)
		if err != nil {
			return err
		}
	}

	for _, owner := range owners {
		ownerAsBytes, _ := json.Marshal(owner)
		err := ctx.GetStub().PutState("OWNER"+strconv.Itoa(owner.Id), ownerAsBytes)

		if err != nil {
			return fmt.Errorf("failed to put to world state. %s", err.Error())
		}
	}

	return nil
}

// CreateCar adds a new car to the world state with given details
func (s *SmartContract) CreateCar(ctx contractapi.TransactionContextInterface, carNumber string, make string, model string, color string, owner string) error {
	car := Car{
		Make:  make,
		Model: model,
		Color: color,
		Owner: owner,
	}

	carAsBytes, _ := json.Marshal(car)

	return ctx.GetStub().PutState(carNumber, carAsBytes)
}

func (s *SmartContract) GetCarById(ctx contractapi.TransactionContextInterface, carId string) (*Car, error) {

	carAsBytes, err := ctx.GetStub().GetState(carId)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state. %s", err.Error())
	}

	if carAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", carId)
	}

	car := new(Car)
	_ = json.Unmarshal(carAsBytes, car)

	return car, nil
}

func (s *SmartContract) GetOwnerById(ctx contractapi.TransactionContextInterface, ownerId string) (*Owner, error) {
	ownerAsBytes, err := ctx.GetStub().GetState(ownerId)

	if err != nil {
		return nil, fmt.Errorf("failed to read from world state. %s", err.Error())
	}

	if ownerAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", ownerId)
	}

	owner := new(Owner)
	_ = json.Unmarshal(ownerAsBytes, owner)

	return owner, nil
}

func (s *SmartContract) GetCarsByColor(ctx contractapi.TransactionContextInterface, color string) ([]*Car, error) {
	resultIter, err := ctx.GetStub().GetStateByPartialCompositeKey("color~owner~id", []string{color})
	if err != nil {
		return nil, err
	}

	defer resultIter.Close()

	cars := make([]*Car, 0)

	for i := 0; resultIter.HasNext(); i++ {
		responseRange, err := resultIter.Next()
		if err != nil {
			return nil, err
		}

		_, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(responseRange.Key)
		if err != nil {
			return nil, err
		}

		carId := compositeKeyParts[2]

		carAsset, err := s.GetCarById(ctx, carId)
		if err != nil {
			return nil, err
		}

		cars = append(cars, carAsset)
	}

	return cars, nil
}

func (s *SmartContract) GetCarsByColorAndOwner(ctx contractapi.TransactionContextInterface, color string, owner string) ([]*Car, error) {
	resultIter, err := ctx.GetStub().GetStateByPartialCompositeKey("color~owner~id", []string{color, owner})
	if err != nil {
		return nil, err
	}

	defer resultIter.Close()

	cars := make([]*Car, 0)

	for i := 0; resultIter.HasNext(); i++ {
		responseRange, err := resultIter.Next()
		if err != nil {
			return nil, err
		}

		_, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(responseRange.Key)
		if err != nil {
			return nil, err
		}

		carId := compositeKeyParts[2]

		carAsset, err := s.GetCarById(ctx, carId)
		if err != nil {
			return nil, err
		}

		cars = append(cars, carAsset)
	}

	return cars, nil
}

func (s *SmartContract) GetAllCars(ctx contractapi.TransactionContextInterface) ([]*Car, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var cars []*Car
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var car Car
		err = json.Unmarshal(queryResponse.Value, &car)
		if err != nil {
			return nil, err
		}
		cars = append(cars, &car)
	}

	return cars, nil
}

// ChangeCarOwner updates the owner field of car with given id in world state
func (s *SmartContract) ChangeCarColor(ctx contractapi.TransactionContextInterface, carId string, color string) error {
	car, err := s.GetCarById(ctx, carId)
	if err != nil {
		return err
	}

	oldKey, _ := ctx.GetStub().CreateCompositeKey("color~owner~id", []string{car.Color, car.Owner, strconv.Itoa(car.Id)})

	car.Color = color
	carAsBytes, _ := json.Marshal(car)
	err = ctx.GetStub().PutState(carId, carAsBytes)
	if err != nil {
		return err
	}

	key, err := ctx.GetStub().CreateCompositeKey("color~owner~id", []string{color, car.Owner, strconv.Itoa(car.Id)})
	if err != nil {
		return err
	}

	ctx.GetStub().DelState(oldKey)
	value := []byte{0x00}
	err = ctx.GetStub().PutState(key, value)
	if err != nil {
		return err
	}

	return nil
}

func (s *SmartContract) AddMalfunction(ctx contractapi.TransactionContextInterface, carId string, description string, price float64) error {

	car, err := s.GetCarById(ctx, carId)
	if err != nil {
		return fmt.Errorf("Car with specified id does not exist")
	}

	malfunction := Malfunction{Description: description, Price: price}

	malfunctions := car.Malfunctions
	malfunctionsPrice := 0.0
	for _, malfunction := range car.Malfunctions {
		malfunctionsPrice += malfunction.Price
	}

	if (malfunctionsPrice + price) <= car.Price {
		malfunctions = append(malfunctions, malfunction)

		car.Malfunctions = malfunctions

		carAsBytes, _ := json.Marshal(car)
		ctx.GetStub().PutState(carId, carAsBytes)
	} else {
		fmt.Println("Price of malfunctions is bigger than price of car, so car will be deleted")
		s.DeleteCar(ctx, carId)
	}

	return nil
}

func (s *SmartContract) RepairCar(ctx contractapi.TransactionContextInterface, carId string) error {

	car, err := s.GetCarById(ctx, carId)
	if err != nil {
		fmt.Println("Car with specified id does not exist")
		return err
	}

	owner, err := s.GetOwnerById(ctx, "OWNER"+car.Owner)
	if err != nil {
		return err
	}

	malfunctionsPrice := 0.0
	for _, malfunction := range car.Malfunctions {
		malfunctionsPrice += malfunction.Price
	}

	if malfunctionsPrice > owner.Money {
		fmt.Println("Owner does not have enough money to repair car")
		return nil
	} else {
		car.Malfunctions = []Malfunction{}
		carAsBytes, _ := json.Marshal(car)
		ctx.GetStub().PutState(carId, carAsBytes)

		money := owner.Money - malfunctionsPrice
		owner.Money = money
		ownerAsBytes, _ := json.Marshal(owner)
		ctx.GetStub().PutState("OWNER"+car.Owner, ownerAsBytes)
	}

	return nil
}

func (s *SmartContract) TransferOwnership(ctx contractapi.TransactionContextInterface, carId string, newOwner string, acceptsMalfunctions bool) error {

	carExists, err := s.CarExists(ctx, carId)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}

	ownerExists, err := s.OwnerExists(ctx, newOwner)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}

	if carExists && ownerExists {
		car, err := s.GetCarById(ctx, carId)
		if err != nil {
			fmt.Println("Car with specified id does not exist")
			return err
		}

		owner, err := s.GetOwnerById(ctx, newOwner)
		if err != nil {
			fmt.Println("Owner with specified id does not exist")
			return err
		}

		if len(car.Malfunctions) == 0 && owner.Money >= car.Price {

			oldOwnerId := car.Owner

			oldKey, _ := ctx.GetStub().CreateCompositeKey("color~owner~id", []string{car.Color, car.Owner, carId})
			car.Owner = strconv.Itoa(owner.Id)
			carAsBytes, _ := json.Marshal(car)
			ctx.GetStub().PutState(carId, carAsBytes)

			key, err := ctx.GetStub().CreateCompositeKey("color~owner~id", []string{car.Color, strconv.Itoa(owner.Id), strconv.Itoa(car.Id)})
			if err != nil {
				return err
			}

			ctx.GetStub().DelState(oldKey)
			value := []byte{0x00}
			err = ctx.GetStub().PutState(key, value)
			if err != nil {
				return err
			}

			ownerMoney := owner.Money - car.Price
			owner.Money = ownerMoney
			ownerAsBytes, _ := json.Marshal(owner)
			ctx.GetStub().PutState(newOwner, ownerAsBytes)

			oldOwner, err := s.GetOwnerById(ctx, "OWNER"+oldOwnerId)
			if err != nil {
				fmt.Println("Owner with specified id does not exist")
				return err
			}

			oldOwner.Money = oldOwner.Money + car.Price
			oldOwnerAsBytes, _ := json.Marshal(oldOwner)
			ctx.GetStub().PutState("OWNER"+oldOwnerId, oldOwnerAsBytes)

		} else if len(car.Malfunctions) > 0 && acceptsMalfunctions {
			malfunctionsPrice := 0.0
			for _, malfunction := range car.Malfunctions {
				malfunctionsPrice += malfunction.Price
			}

			carPrice := car.Price - malfunctionsPrice

			if owner.Money >= carPrice {
				oldOwnerId := car.Owner

				oldKey, _ := ctx.GetStub().CreateCompositeKey("color~owner~id", []string{car.Color, car.Owner, carId})
				car.Owner = strconv.Itoa(owner.Id)
				carAsBytes, _ := json.Marshal(car)
				ctx.GetStub().PutState(carId, carAsBytes)

				key, err := ctx.GetStub().CreateCompositeKey("color~owner~id", []string{car.Color, strconv.Itoa(owner.Id), strconv.Itoa(car.Id)})
				if err != nil {
					return err
				}

				ctx.GetStub().DelState(oldKey)
				value := []byte{0x00}
				err = ctx.GetStub().PutState(key, value)
				if err != nil {
					return err
				}

				ownerMoney := owner.Money - carPrice
				owner.Money = ownerMoney
				ownerAsBytes, _ := json.Marshal(owner)
				ctx.GetStub().PutState(newOwner, ownerAsBytes)

				oldOwner, err := s.GetOwnerById(ctx, "OWNER"+oldOwnerId)
				if err != nil {
					fmt.Println("Owner with specified id does not exist")
					return err
				}

				oldOwner.Money = oldOwner.Money + carPrice
				oldOwnerAsBytes, _ := json.Marshal(oldOwner)
				ctx.GetStub().PutState("OWNER"+oldOwnerId, oldOwnerAsBytes)

			} else {
				return fmt.Errorf("new owner does not have enough money to buy this car")
			}
		} else if len(car.Malfunctions) > 0 && !acceptsMalfunctions {

			return fmt.Errorf("car has malfunctions and new owner does not want them")

		} else if len(car.Malfunctions) == 0 && owner.Money < car.Price {

			return fmt.Errorf("new owner does not have enough money to buy this car")
		}

	}

	return nil
}

func (s *SmartContract) CarExists(ctx contractapi.TransactionContextInterface, carId string) (bool, error) {
	car, err := ctx.GetStub().GetState(carId)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return car != nil, nil
}

func (s *SmartContract) OwnerExists(ctx contractapi.TransactionContextInterface, ownerId string) (bool, error) {
	owner, err := ctx.GetStub().GetState(ownerId)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return owner != nil, nil
}

func (s *SmartContract) DeleteCar(ctx contractapi.TransactionContextInterface, carId string) error {
	exists, err := s.CarExists(ctx, carId)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}

	if !exists {
		return fmt.Errorf("the asset %s does not exist", carId)
	}

	return ctx.GetStub().DelState(carId)
}

func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create fabcar chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting fabcar chaincode: %s", err.Error())
	}
}
