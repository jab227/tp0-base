package agency

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

type bet struct {
	birthdate time.Time
	name      string
	surname   string
	dni       string
	number    int
}

func NewBet(name, surname, dni, birthdate, number string) (bet, error) {
	parsedBirthdate, err := time.Parse(time.DateOnly, birthdate)
	if err != nil {
		return bet{}, errors.Wrap(err, "invalid birthdate")
	}

	parsedNumber, err := strconv.ParseInt(number, 10, 32)
	if err != nil {
		return bet{}, errors.Wrap(err, "invalid bet number")
	}

	return bet{
		name:      name,
		surname:   surname,
		dni:       dni,
		number:    int(parsedNumber),
		birthdate: parsedBirthdate,
	}, nil
}

func (b bet) MarshalBet() []byte {
	return []byte(fmt.Sprintf("%s,%s,%s,%s,%d",
		b.name,
		b.surname,
		b.dni,
		b.birthdate.Format(time.DateOnly),
		b.number))
}
