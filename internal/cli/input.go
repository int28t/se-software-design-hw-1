package cli

import (
	"fmt"
	"strconv"
	"strings"
)

func (s *session) readIdentity() (string, string, error) {
	number, err := s.readLine("Инвентарный номер: ")
	if err != nil {
		return "", "", err
	}
	name, err := s.readLine("Название: ")
	if err != nil {
		return "", "", err
	}
	return number, name, nil
}

func (s *session) readEnergy() (float64, error) {
	consumption, err := s.readLine("Суточный расход энергии (кВт·ч): ")
	if err != nil {
		return 0, err
	}
	dailyEnergyKWh, err := strconv.ParseFloat(strings.ReplaceAll(consumption, ",", "."), 64)
	if err != nil {
		return 0, fmt.Errorf("%w: укажите расход энергии числом, например 2.5", errInvalidInput)
	}
	return dailyEnergyKWh, nil
}
