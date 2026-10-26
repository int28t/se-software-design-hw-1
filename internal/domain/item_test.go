package domain_test

import (
	"errors"
	"math"
	"testing"

	"escooter-sharing/internal/domain"
)

func TestInventoryItemIdentity(t *testing.T) {
	t.Parallel()
	factories := []struct {
		name string
		new  func(string, string) (domain.InventoryItem, error)
	}{
		{name: "вещь", new: func(number, name string) (domain.InventoryItem, error) {
			return domain.NewItem(number, name)
		}},
		{name: "шлем", new: func(number, name string) (domain.InventoryItem, error) {
			return domain.NewHelmet(number, name)
		}},
		{name: "станция", new: func(number, name string) (domain.InventoryItem, error) {
			return domain.NewChargingStation(number, name, 1.25)
		}},
	}
	for _, factory := range factories {
		t.Run(factory.name, func(t *testing.T) {
			t.Parallel()
			for _, tt := range []struct {
				name   string
				number string
				title  string
				valid  bool
			}{
				{name: "корректно", number: " I-1 ", title: " Вещь ", valid: true},
				{name: "пустой номер", title: "Вещь"},
				{name: "номер из пробелов", number: " \t", title: "Вещь"},
				{name: "пустое название", number: "I-1"},
				{name: "название из пробелов", number: "I-1", title: " \t"},
			} {
				t.Run(tt.name, func(t *testing.T) {
					t.Parallel()
					item, err := factory.new(tt.number, tt.title)
					if !tt.valid {
						if !errors.Is(err, domain.ErrInvalidItem) {
							t.Fatalf("создание: %v, want ErrInvalidItem", err)
						}
						return
					}
					if err != nil {
						t.Fatal(err)
					}
					if item.InventoryNumber() != "I-1" || item.Name() != "Вещь" {
						t.Fatalf("неверные данные вещи: %+v", item)
					}
					if _, ok := item.(domain.Vehicle); ok {
						t.Fatal("вещь не должна быть транспортом")
					}
				})
			}
		})
	}
}

func TestChargingStationEnergy(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name   string
		energy float64
		valid  bool
	}{
		{name: "ноль", valid: true},
		{name: "дробный расход", energy: 2.5, valid: true},
		{name: "отрицательный расход", energy: -1},
		{name: "NaN", energy: math.NaN()},
		{name: "Inf", energy: math.Inf(1)},
		{name: "-Inf", energy: math.Inf(-1)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			station, err := domain.NewChargingStation("C-1", "Зарядная станция", tt.energy)
			if !tt.valid {
				if !errors.Is(err, domain.ErrInvalidItem) {
					t.Fatalf("создание: %v, want ErrInvalidItem", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var consumer domain.EnergyConsumer = station
			if consumer.DailyEnergyKWh() != tt.energy {
				t.Fatalf("расход = %v, want %v", consumer.DailyEnergyKWh(), tt.energy)
			}
		})
	}
}

func TestHelmetDoesNotConsumeElectricity(t *testing.T) {
	t.Parallel()
	helmet, err := domain.NewHelmet("H-1", "Шлем")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := any(helmet).(domain.EnergyConsumer); ok {
		t.Fatal("шлем не должен быть потребителем энергии")
	}
}

func TestInvalidInventoryItems(t *testing.T) {
	t.Parallel()
	for _, item := range []domain.InventoryItem{nil, domain.Item{}, domain.Helmet{}, domain.ChargingStation{}} {
		if err := domain.ValidateInventoryItem(item); !errors.Is(err, domain.ErrInvalidItem) {
			t.Errorf("ValidateInventoryItem(%T) = %v, want ErrInvalidItem", item, err)
		}
	}
}
