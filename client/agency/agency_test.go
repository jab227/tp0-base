package agency_test

import (
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/agency"
	"testing"
)

func TestSerializeBet(t *testing.T) {
	t.Run("serialize bet correctly", func(t *testing.T) {
		bettor := agency.Bettor{
			Name:      "Julio",
			Surname:   "Cortazar",
			DNI:       "52820003",
			Birthdate: "1999-03-17",
			BetNumber: "7574",
		}
		bet, err := agency.NewBet(bettor)
		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		want := "Julio,Cortazar,52820003,1999-03-17,7574"
		got := bet.MarshalBet()

		if string(got) != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("error parsing birthdate", func(t *testing.T) {
		bettor := agency.Bettor{
			Name:      "Julio",
			Surname:   "Cortazar",
			DNI:       "52820003",
			Birthdate: "not a date",
			BetNumber: "7574",
		}
		_, err := agency.NewBet(bettor)
		if err == nil {
			t.Error("Expected an error")
		}

	})

	t.Run("error parsing bet number", func(t *testing.T) {
		bettor := agency.Bettor{
			Name:      "Julio",
			Surname:   "Cortazar",
			DNI:       "52820003",
			Birthdate: "1999-03-17",
			BetNumber: "not a number",
		}
		_, err := agency.NewBet(bettor)
		if err == nil {
			t.Error("Expected an error")
		}

	})
}
