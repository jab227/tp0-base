package agency

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

type Bettor struct {
	Name      string
	Surname   string
	DNI       string
	Birthdate string
	BetNumber string
}

type Bet struct {
	birthdate time.Time
	name      string
	surname   string
	dni       string
	number    uint32
}

const timeLayout = "2006-01-2"

func NewBet(b Bettor) (Bet, error) {
	parsedBirthdate, err := time.Parse(timeLayout, b.Birthdate)
	if err != nil {
		return Bet{}, errors.Wrap(err, "invalid birthdate")
	}

	parsedNumber, err := strconv.ParseUint(b.BetNumber, 10, 32)
	if err != nil {
		return Bet{}, errors.Wrap(err, "invalid bet number")
	}

	return Bet{
		name:      b.Name,
		surname:   b.Surname,
		dni:       b.DNI,
		number:    uint32(parsedNumber),
		birthdate: parsedBirthdate,
	}, nil
}

func (b Bet) MarshalBet() []byte {
	return []byte(fmt.Sprintf("%s,%s,%s,%s,%d",
		b.name,
		b.surname,
		b.dni,
		b.birthdate.Format(timeLayout),
		b.number))
}
