package agency_test

import (
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/agency"
	"testing"
)

func TestSerializeBet(t *testing.T) {
	t.Run("serialize bet correctly", func(t *testing.T) {
		const (
			name      = "Julio"
			surname   = "Cortazar"
			dni       = "52820003"
			birthdate = "1999-03-17"
			betNumber = "7574"
		)

		bet, err := agency.NewBet(name, surname, dni, birthdate, betNumber)
		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}


		want := "Julio,Cortazar,52820003,1999-03-17,7574\n"
		got := bet.String()

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("error parsing birthdate", func(t *testing.T) {
		const (
			name      = "Julio"
			surname   = "Cortazar"
			dni       = "52820003"
			birthdate = "1999-03-17"			
			betNumber = "not a number"
		)

		_, err := agency.NewBet(name, surname, dni, birthdate, betNumber)
		if err == nil {
			t.Error("Expected an error")
		}

	})

	t.Run("error parsing bet number", func(t *testing.T) {
		const (
			name      = "Julio"
			surname   = "Cortazar"
			dni       = "52820003"
			birthdate = "not a date"
			betNumber = "not a number"
		)

		_, err := agency.NewBet(name, surname, dni, birthdate, betNumber)
		if err == nil {
			t.Error("Expected an error")
		}

	})
}
