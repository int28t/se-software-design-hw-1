package domain_test

import (
	"errors"
	"math"
	"testing"

	"escooter-sharing/internal/domain"
)

func TestNewElectricTransportEnergy(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		energy float64
		valid  bool
	}{
		{name: "нулевой расход", energy: 0, valid: true},
		{name: "дробный расход", energy: 1.25, valid: true},
		{name: "максимальное конечное число", energy: math.MaxFloat64, valid: true},
		{name: "отрицательный расход", energy: -1},
		{name: "NaN", energy: math.NaN()},
		{name: "положительная бесконечность", energy: math.Inf(1)},
		{name: "отрицательная бесконечность", energy: math.Inf(-1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			transport, err := domain.NewElectricTransport(" E-1 ", " Электросамокат ", true, 6, tt.energy)
			if !tt.valid {
				if !errors.Is(err, domain.ErrInvalidTransport) {
					t.Fatalf("NewElectricTransport() error = %v, want ErrInvalidTransport", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var consumer domain.EnergyConsumer = transport
			if consumer.DailyEnergyKWh() != tt.energy || transport.InventoryNumber() != "E-1" || transport.Name() != "Электросамокат" {
				t.Fatalf("неверные характеристики: %+v", transport)
			}
			approved, err := (domain.ServiceCenter{}).Inspect(transport)
			if err != nil || !approved {
				t.Fatalf("Inspect() = %v, %v; want true, nil", approved, err)
			}
		})
	}
}

func TestNewElectricTransportRejectsInvalidIdentity(t *testing.T) {
	t.Parallel()
	_, err := domain.NewElectricTransport("", "Электросамокат", true, 6, 1)
	if !errors.Is(err, domain.ErrInvalidTransport) {
		t.Fatalf("NewElectricTransport() error = %v, want ErrInvalidTransport", err)
	}
}

func TestTransportWithoutElectricityIsNotEnergyConsumer(t *testing.T) {
	t.Parallel()
	transport, err := domain.NewTransport("B-1", "Велосипед", true, 6)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := any(transport).(domain.EnergyConsumer); ok {
		t.Fatal("обычный транспорт не должен реализовывать EnergyConsumer")
	}
}

func TestValidateVehicleRejectsNil(t *testing.T) {
	t.Parallel()
	approved, err := (domain.ServiceCenter{}).Inspect(nil)
	if approved || !errors.Is(err, domain.ErrInvalidTransport) {
		t.Fatalf("Inspect(nil) = %v, %v; want false, ErrInvalidTransport", approved, err)
	}
}
