package application_test

import (
	"errors"
	"slices"
	"testing"

	"escooter-sharing/internal/application"
	"escooter-sharing/internal/domain"
	"escooter-sharing/internal/infrastructure/memory"
)

// cargoBicycle добавляет вид транспорта только в тесте, без изменения приложения.
type cargoBicycle struct {
	domain.Bicycle
}

// chargingCabinet добавляет новый вид инвентаря, потребляющий электроэнергию.
type chargingCabinet struct {
	domain.Item
}

func (chargingCabinet) DailyEnergyKWh() float64 { return 5 }

func TestFleetSupportsNewVehicleAndItemTypes(t *testing.T) {
	t.Parallel()
	bicycle, err := domain.NewBicycle("B-1", "Велосипед", true, 6)
	if err != nil {
		t.Fatal(err)
	}
	scooter, err := domain.NewScooter("S-1", "Самокат", true, 6, 1.25)
	if err != nil {
		t.Fatal(err)
	}
	electricBicycle, err := domain.NewElectricBicycle("E-1", "Электровелосипед", false, 10, 2.5)
	if err != nil {
		t.Fatal(err)
	}
	cargoBase, err := domain.NewBicycle("C-1", "Грузовой велосипед", true, 5)
	if err != nil {
		t.Fatal(err)
	}
	cabinetBase, err := domain.NewItem("P-1", "Зарядный шкаф")
	if err != nil {
		t.Fatal(err)
	}
	fleet := application.NewFleet(domain.ServiceCenter{}, &memory.FleetRepository{})
	for _, transport := range []domain.Vehicle{bicycle, scooter, electricBicycle, cargoBicycle{Bicycle: cargoBase}} {
		if _, err := fleet.Admit(transport); err != nil {
			t.Fatal(err)
		}
	}
	for _, item := range []domain.InventoryItem{
		makeHelmet(t, "H-1"), makeStation(t, "Z-1"), chargingCabinet{Item: cabinetBase},
	} {
		if err := fleet.AddItem(item); err != nil {
			t.Fatal(err)
		}
	}

	report, err := fleet.Report()
	want := application.Report{
		TotalAssets: 7, TotalItems: 3, TotalTransports: 4,
		ActiveTransports: 3, RejectedTransports: 1,
		TotalDailyEnergyKWh: 11.75, ActiveDailyEnergyKWh: 1.25,
	}
	if err != nil || report != want {
		t.Fatalf("Report() = %+v, %v; want %+v", report, err, want)
	}
	entries, err := fleet.Inventory()
	wantEntries := []application.InventoryEntry{
		{InventoryNumber: "B-1", Name: "Велосипед"},
		{InventoryNumber: "C-1", Name: "Грузовой велосипед"},
		{InventoryNumber: "E-1", Name: "Электровелосипед"},
		{InventoryNumber: "H-1", Name: "Шлем"},
		{InventoryNumber: "P-1", Name: "Зарядный шкаф"},
		{InventoryNumber: "S-1", Name: "Самокат"},
		{InventoryNumber: "Z-1", Name: "Зарядная станция"},
	}
	if err != nil || !slices.Equal(entries, wantEntries) {
		t.Fatalf("Inventory() = %v, %v; want %v", entries, err, wantEntries)
	}
	active, err := fleet.ActiveTransports()
	wantActive := []domain.Vehicle{bicycle, cargoBicycle{Bicycle: cargoBase}, scooter}
	if err != nil || !slices.Equal(active, wantActive) {
		t.Fatalf("ActiveTransports() = %v, %v; want %v", active, err, wantActive)
	}
	beginner, err := fleet.BeginnerTransports()
	if err != nil || !slices.Equal(beginner, []domain.Vehicle{bicycle, scooter}) {
		t.Fatalf("BeginnerTransports() = %v, %v; want велосипед и самокат", beginner, err)
	}
	if err := fleet.AddItem(cargoBicycle{Bicycle: cargoBase}); !errors.Is(err, application.ErrTransportRequiresInspection) {
		t.Fatalf("обход техосмотра новым видом: %v", err)
	}
}
