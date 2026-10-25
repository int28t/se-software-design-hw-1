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
	TotalAssets          int
	TotalItems           int
	TotalTransports      int
	ActiveTransports     int
	RejectedTransports   int
	TotalDailyEnergyKWh  float64
	ActiveDailyEnergyKWh float64
}

// Report рассчитывает количество объектов и суточную энергию по одному снимку баланса.
func (f *Fleet) Report() (Report, error) {
	admissions, err := f.repository.List()
	if err != nil {
		return Report{}, fmt.Errorf("отчёт о парке: %w", err)
	}
	report := Report{TotalAssets: len(admissions)}
	for _, admission := range admissions {
		_, isTransport := admission.Item().(domain.Vehicle)
		active := isTransport && admission.Approved()
		if isTransport {
			report.TotalTransports++
			if active {
				report.ActiveTransports++
			} else {
				report.RejectedTransports++
			}
		} else {
			report.TotalItems++
		}
		if consumer, ok := admission.Item().(domain.EnergyConsumer); ok {
			energy := consumer.DailyEnergyKWh()
			report.TotalDailyEnergyKWh += energy
			if active {
				report.ActiveDailyEnergyKWh += energy
			}
		}
	}
	if math.IsInf(report.TotalDailyEnergyKWh, 0) || math.IsInf(report.ActiveDailyEnergyKWh, 0) {
		return Report{}, ErrEnergyOverflow
	}
	return report, nil
}
