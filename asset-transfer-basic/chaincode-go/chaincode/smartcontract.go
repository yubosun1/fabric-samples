package chaincode

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Record
type SmartContract struct {
	contractapi.Contract
}

type Book struct {
	BookID string `json:"BookID"`
	Name   string `json:"Name"`
	Author string `json:"Author"`
	Valid  bool   `json:"Valid"`
	Price  int    `json:"Price"`
	Owner  string `json:"Owner"`
}

// Record describes basic details of what makes up a simple asset
//Insert struct field in alphabetic order => to achieve determinism across languages
// golang keeps the order when marshal to json but doesn't order automatically
type Record struct {
	RecordID  string    `json:"RecordID"`
	BookID    string    `json:"BookID"`
	StartTime time.Time `json:"StartTime"`
	EndTime   time.Time `json:"EndTime"`
	Borrower  string    `json:"Borrower"`
}

// InitLedger adds a base set of book to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	books := []Book{
		{BookID: "book1", Name: "Journey to the West", Author: "WuChengEn", Valid: true, Price: 20, Owner: "library"},
		{BookID: "book2", Name: "A Dream of Red Mansions", Author: "CaoXueQin", Valid: true, Price: 21, Owner: "library"},
		{BookID: "book3", Name: "Three Kingdoms", Author: "LuoGuanZhong", Valid: true, Price: 22, Owner: "library"},
		{BookID: "book4", Name: "Water Margin", Author: "ShiNaiAn", Valid: true, Price: 23, Owner: "library"},
	}

	for _, book := range books {
		bookJSON, err := json.Marshal(book)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(book.BookID, bookJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// AddBook issues a new book to the world state with given details.
func (s *SmartContract) AddBook(ctx contractapi.TransactionContextInterface, bookId string, name string, author string, price int) error {
	exists, err := s.BookExists(ctx, bookId)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the book %s already exists", bookId)
	}

	book := Book{
		BookID: bookId,
		Name:   name,
		Author: author,
		Valid:  true,
		Price:  price,
		Owner:  "library",
	}
	bookJSON, err := json.Marshal(book)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(bookId, bookJSON)
}

// AddRecord issues a new record to the world state with given details.
func (s *SmartContract) AddRecord(ctx contractapi.TransactionContextInterface, recordId string, bookId string, startTime time.Time, borrower string) error {
	exists, err := s.RecordExists(ctx, recordId)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the record %s already exists", recordId)
	}

	record := Record{
		RecordID:  recordId,
		BookID:    bookId,
		StartTime: startTime,
		Borrower:  borrower,
	}
	recordJSON, err := json.Marshal(record)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(recordId, recordJSON)
}

// QueryBook returns the book stored in the world state with given id.
func (s *SmartContract) QueryBook(ctx contractapi.TransactionContextInterface, bookId string) (*Book, error) {
	bookJSON, err := ctx.GetStub().GetState(bookId)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if bookJSON == nil {
		return nil, fmt.Errorf("the book %s does not exist", bookId)
	}

	var book Book
	err = json.Unmarshal(bookJSON, &book)
	if err != nil {
		return nil, err
	}

	return &book, nil
}

// QueryRecord returns the record stored in  the world state with given id.
func (s *SmartContract) QueryRecord(ctx contractapi.TransactionContextInterface, recordId string) (*Record, error) {
	recordJSON, err := ctx.GetStub().GetState(recordId)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if recordJSON == nil {
		return nil, fmt.Errorf("the record %s does not exist", recordId)
	}

	var record Record
	err = json.Unmarshal(recordJSON, &record)
	if err != nil {
		return nil, err
	}

	return &record, nil
}

// BorrowBook updates an existing book and add record in the world state with provided parameters.
func (s *SmartContract) BorrowBook(ctx contractapi.TransactionContextInterface, recordId string, bookId string, newOwner string, startTime time.Time) error {
	book, err := s.QueryBook(ctx, bookId)
	if err != nil {
		return err
	}

	book.Valid = false
	book.Owner = newOwner
	bookJSON, err := json.Marshal(book)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(bookId, bookJSON)
	if err != nil {
		return err
	}

	err = s.AddRecord(ctx, recordId, bookId, startTime, newOwner)

	return err
}
func (s *SmartContract) ReturnBook(ctx contractapi.TransactionContextInterface, recordId string, bookId string, endTime time.Time) error {
	book, err := s.QueryBook(ctx, bookId)
	if err != nil {
		return err
	}

	book.Valid = true
	book.Owner = "library"
	bookJSON, err := json.Marshal(book)
	if err != nil {
		return err
	}
	err = ctx.GetStub().PutState(bookId, bookJSON)
	if err != nil {
		return err
	}

	record, err := s.QueryRecord(ctx, recordId)
	if err != nil {
		return err
	}

	record.EndTime = endTime
	recordJSON, err := json.Marshal(record)
	if err != nil {
		return err
	}
	err = ctx.GetStub().PutState(recordId, recordJSON)
	if err != nil {
		return err
	}

	return err
}

// DeleteBook deletes an given book from the world state.
func (s *SmartContract) DeleteBook(ctx contractapi.TransactionContextInterface, bookId string) error {
	exists, err := s.BookExists(ctx, bookId)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", bookId)
	}

	return ctx.GetStub().DelState(bookId)
}

// BookExists returns true when book with given ID exists in world state
func (s *SmartContract) BookExists(ctx contractapi.TransactionContextInterface, bookId string) (bool, error) {
	bookJSON, err := ctx.GetStub().GetState(bookId)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return bookJSON != nil, nil
}

// RecordExists returns true when record with given ID exists in world state
func (s *SmartContract) RecordExists(ctx contractapi.TransactionContextInterface, recordId string) (bool, error) {
	recordJSON, err := ctx.GetStub().GetState(recordId)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return recordJSON != nil, nil
}

// GetAllBooks returns all books found in world state
func (s *SmartContract) GetAllBooks(ctx contractapi.TransactionContextInterface) (map[string]int, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all books in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	bookList := make(map[string]int)
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var book Book
		err = json.Unmarshal(queryResponse.Value, &book)
		if err != nil {
			return nil, err
		}
		bookList[book.Name]++
	}

	return bookList, nil
}

func (s *SmartContract) GetBorrowList(ctx contractapi.TransactionContextInterface, name string) ([]*Book, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var books []*Book
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var book Book
		err = json.Unmarshal(queryResponse.Value, &book)
		if err != nil {
			return nil, err
		}

		if book.Name == name {
			books = append(books, &book)
		}
	}
	return books, nil
}
