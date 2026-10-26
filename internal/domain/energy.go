package domain

import (
	"fmt"
	"math"
)

// EnergyConsumer описывает объект с заданным суточным расходом электроэнергии.
// Расход должен быть конечным неотрицательным числом в кВт·ч.
type EnergyConsumer interface {
	DailyEnergyKWh() float64
}

func validateDailyEnergyKWh(energy float64, invalid error) error {
	if energy < 0 || math.IsNaN(energy) || math.IsInf(energy, 0) {
		return fmt.Errorf("%w: расход энергии должен быть конечным числом не меньше нуля", invalid)
	}
	return nil
}
