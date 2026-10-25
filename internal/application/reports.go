package application

import (
	"errors"
	"fmt"
	"math"

	"escooter-sharing/internal/domain"
)

// ErrEnergyOverflow означает, что суммарный расход не помещается в float64.
var ErrEnergyOverflow = errors.New("суммарный расход энергии слишком велик")

// Report содержит показатели всего баланса и активного транспорта.
// Энергия означает сумму заданных суточных расходов, а не фактический замер.
type Report struct {
	TotalTransports      int
	ActiveTransports     int
	RejectedTransports   int
	TotalDailyEnergyKWh  float64
	ActiveDailyEnergyKWh float64
}

// Report рассчитывает количество транспорта и суточную энергию по одному снимку баланса.
func (f *Fleet) Report() (Report, error) {
	admissions, err := f.repository.List()
	if err != nil {
		return Report{}, fmt.Errorf("отчёт о парке: %w", err)
	}
	report := Report{TotalTransports: len(admissions)}
	for _, admission := range admissions {
		if admission.Approved() {
			report.ActiveTransports++
		} else {
			report.RejectedTransports++
		}
		if consumer, ok := admission.Transport().(domain.EnergyConsumer); ok {
			energy := consumer.DailyEnergyKWh()
			report.TotalDailyEnergyKWh += energy
			if admission.Approved() {
				report.ActiveDailyEnergyKWh += energy
			}
		}
	}
	if math.IsInf(report.TotalDailyEnergyKWh, 0) || math.IsInf(report.ActiveDailyEnergyKWh, 0) {
		return Report{}, ErrEnergyOverflow
	}
	return report, nil
}
