package main

import (
	"context"
	"fmt"
	"log"

	"github.com/freel/backend/internal/contracts"
	"github.com/freel/backend/internal/database"
)

func strPtr(s string) *string {
	return &s
}

func main() {
	db, err := database.Connect("root@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}

	dl := contracts.NewDataLayer(db)

	ctx := context.Background()
	orgID := int64(1)

	// Create party
	res, err := db.ExecContext(ctx, "INSERT INTO contract_parties (org_id, party_name, party_type) VALUES (?, ?, ?)", orgID, "Test Carrier", "CARRIER")
	if err != nil {
		log.Fatal(err)
	}
	partyID, _ := res.LastInsertId()

	res2, _ := db.ExecContext(ctx, "INSERT INTO contract_parties (org_id, party_name, party_type) VALUES (?, ?, ?)", orgID, "Test Customer", "CUSTOMER")
	partyID2, _ := res2.LastInsertId()

	contractsToCreate := []contracts.Contract{
		{
			OrgID:             orgID,
			ContractReference: "REF-A",
			ContractName:      "Test Contract A - Draft",
			ContractType:      "CARRIER",
			PartyID:           partyID,
			PartyName:         "Test Carrier",
			Status:            "DRAFT",
			EffectiveDate:     strPtr("2026-09-01"),
			ExpiryDate:        strPtr("2027-09-01"),
			Currency:          strPtr("USD"),
		},
		{
			OrgID:             orgID,
			ContractReference: "REF-B",
			ContractName:      "Test Contract B - Active",
			ContractType:      "CUSTOMER",
			PartyID:           partyID2,
			PartyName:         "Test Customer",
			Status:            "ACTIVE",
			EffectiveDate:     strPtr("2025-01-01"),
			ExpiryDate:        strPtr("2028-01-01"),
			Currency:          strPtr("USD"),
		},
	}

	for _, c := range contractsToCreate {
		id, err := dl.CreateContract(ctx, &c)
		if err != nil {
			log.Fatalf("failed to create contract: %v", err)
		}
		fmt.Printf("Created contract ID %d: %s\n", id, c.ContractName)
	}
}
