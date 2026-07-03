package main

import (
	"errors"
	"fmt"
)

type Entry struct {
	ChargeID  string // identifier of the charge this entry belongs to
	Direction string // "debit" or "credit"
	Account   string // customer or shop
	Amount    int    // money store as the smallest currency unit(cents) e.g. $23.89 -> 2389; avoid fload, which break the debit == credit balance
}

var ledger []Entry
var refundable int

func main() {
	// 1. customer pay for 100
	err := appendLedger("123", "customer", "shop", 100)
	if err != nil {
		panic(err)
	}

	// 2. customer refund 30 at the first time
	err = appendLedger("123", "shop", "customer", 30)
	if err != nil {
		panic(err)
	}

	// 3. customer refund 60 at the first time
	err = appendLedger("123", "shop", "customer", 60)
	if err != nil {
		panic(err)
	}

	var debit int
	var credit int
	for _, entry := range ledger {
		direction := entry.Direction
		amount := entry.Amount
		switch direction {
		case "debit":
			debit += amount
		case "credit":
			credit += amount
		}
	}

	if debit == credit {
		fmt.Println("Success: debit == credit")
	} else {
		fmt.Println("Failed: debit != credit")
	}

	// 4. edge case: refund exceeds refundable should be rejected and not record to ledger
	lenBefore := len(ledger)
	err = appendLedger("123", "shop", "customer", 80) // refundable only 10 left
	if err == nil {
	    fmt.Println("Failed: over-refund should have been rejected")
	} else if len(ledger) != lenBefore {
	    fmt.Println("Failed: rejected refund still be recored to the ledger")
	} else {
	    fmt.Println("Success: over-refund rejected, ledger untouched")
	}
}

func appendLedger(chargeID, from, to string, amount int) error {
	// 1. check if it's payment or refund
	if from == "customer" && to == "shop" {
		refundable += amount
	}
	// 2. refund should not exceed refundable
	if from == "shop" && to == "customer" {
		if refundable < amount {
			return errors.New("refund exceeds refundable")
		} else {
			refundable -= amount
		}
	}

	fromEntry := Entry{
		ChargeID:  chargeID,
		Direction: "debit",
		Account:   from,
		Amount:    amount,
	}

	toEntry := Entry{
		ChargeID:  chargeID,
		Direction: "credit",
		Account:   to,
		Amount:    amount,
	}
	ledger = append(ledger, fromEntry)
	ledger = append(ledger, toEntry)

	return nil
}
