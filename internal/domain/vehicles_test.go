package domain_test

import (
	"errors"
	"math"
	"testing"

	"escooter-sharing/internal/domain"
)

func TestVehicleKindsPreserveCharacteristicsAndRoles(t *testing.T) {
	t.Parallel()
	factories := []struct {
		name     string
		electric bool
		new      func() (domain.Vehicle, error)
	}{
		{name: "велосипед", new: func() (domain.Vehicle, error) {
			return domain.NewBicycle(" T-1 ", " Транспорт ", false, 6)
		}},
		{name: "самокат", electric: true, new: func() (domain.Vehicle, error) {
			return domain.NewScooter(" T-1 ", " Транспорт ", false, 6, 1.25)
		}},
		{name: "электровелосипед", electric: true, new: func() (domain.Vehicle, error) {
			return domain.NewElectricBicycle(" T-1 ", " Транспорт ", false, 6, 1.25)
		}},
	}
	for _, factory := range factories {
		t.Run(factory.name, func(t *testing.T) {
			t.Parallel()
			transport, err := factory.new()
			if err != nil {
				t.Fatal(err)
			}
			var item domain.InventoryItem = transport
			if item.InventoryNumber() != "T-1" || item.Name() != "Транспорт" {
				t.Fatalf("неверные данные инвентаризации: %+v", item)
			}
			if transport.BeginnerEase() != 6 || !transport.SuitableForBeginner() || transport.IsServiceable() {
				t.Fatalf("неверные характеристики транспорта: %+v", transport)
			}
			approved, err := (domain.ServiceCenter{}).Inspect(transport)
			if err != nil || approved {
				t.Fatalf("неисправная техника: допуск = %v, ошибка = %v", approved, err)
			}
			consumer, ok := transport.(domain.EnergyConsumer)
			if ok != factory.electric {
				t.Fatalf("EnergyConsumer = %v, want %v", ok, factory.electric)
			}
			if ok && consumer.DailyEnergyKWh() != 1.25 {
				t.Fatalf("расход энергии = %v, want 1.25", consumer.DailyEnergyKWh())
			}
		})
	}
}

func TestConcreteVehicleConstructorsRejectInvalidData(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		new  func() (domain.Vehicle, error)
	}{
		{name: "велосипед без номера", new: func() (domain.Vehicle, error) {
			return domain.NewBicycle("", "Велосипед", true, 6)
		}},
		{name: "самокат с отрицательным расходом", new: func() (domain.Vehicle, error) {
			return domain.NewScooter("S-1", "Самокат", true, 6, -1)
		}},
		{name: "электровелосипед с NaN", new: func() (domain.Vehicle, error) {
			return domain.NewElectricBicycle("E-1", "Электровелосипед", true, 6, math.NaN())
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if _, err := tt.new(); !errors.Is(err, domain.ErrInvalidTransport) {
				t.Fatalf("создание: %v, want ErrInvalidTransport", err)
			}
		})
	}
}
