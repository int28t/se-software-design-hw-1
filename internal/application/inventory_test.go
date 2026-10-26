package application_test

import (
	"errors"
	"slices"
	"testing"

	"escooter-sharing/internal/application"
	"escooter-sharing/internal/domain"
	"escooter-sharing/internal/infrastructure/memory"
)

func makeHelmet(t *testing.T, number string) domain.Helmet {
	t.Helper()
	helmet, err := domain.NewHelmet(number, "Шлем")
	if err != nil {
		t.Fatal(err)
	}
	return helmet
}

func makeStation(t *testing.T, number string) domain.ChargingStation {
	t.Helper()
	station, err := domain.NewChargingStation(number, "Зарядная станция", 3)
	if err != nil {
		t.Fatal(err)
	}
	return station
}

func TestItemsAreAddedWithoutInspection(t *testing.T) {
	t.Parallel()
	repository := &memory.FleetRepository{}
	inspector := inspectorFunc(func(domain.Vehicle) (bool, error) {
		t.Error("добавление вещи не должно вызывать техосмотр")
		return false, errors.New("неожиданный техосмотр")
	})
	fleet := application.NewFleet(inspector, repository)
	for _, item := range []domain.InventoryItem{makeHelmet(t, "H-1"), makeStation(t, "C-1")} {
		if err := fleet.AddItem(item); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := fleet.Inventory()
	want := []application.InventoryEntry{
		{InventoryNumber: "C-1", Name: "Зарядная станция"},
		{InventoryNumber: "H-1", Name: "Шлем"},
	}
	if err != nil || !slices.Equal(entries, want) {
		t.Fatalf("Inventory() = %v, %v; want %v", entries, err, want)
	}
	active, err := fleet.ActiveTransports()
	if err != nil || len(active) != 0 {
		t.Fatalf("вещи попали в активный парк: %v, %v", active, err)
	}
	beginner, err := fleet.BeginnerTransports()
	if err != nil || len(beginner) != 0 {
		t.Fatalf("вещи попали в выдачу новичкам: %v, %v", beginner, err)
	}
	records, err := repository.List()
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range records {
		if record.Approved() {
			t.Fatal("вещь получила допуск транспорта")
		}
	}
}

func TestInventoryAndReportIncludeTransportAndItems(t *testing.T) {
	t.Parallel()
	fleet := application.NewFleet(domain.ServiceCenter{}, &memory.FleetRepository{})
	accepted := makeElectricTransport(t, "T-1", true, 1.25)
	rejected := makeElectricTransport(t, "T-2", false, 2.5)
	for _, transport := range []domain.Vehicle{rejected, accepted} {
		if _, err := fleet.Admit(transport); err != nil {
			t.Fatal(err)
		}
	}
	for _, item := range []domain.InventoryItem{makeHelmet(t, "H-1"), makeStation(t, "C-1")} {
		if err := fleet.AddItem(item); err != nil {
			t.Fatal(err)
		}
	}
	want := []application.InventoryEntry{
		{InventoryNumber: "C-1", Name: "Зарядная станция"},
		{InventoryNumber: "H-1", Name: "Шлем"},
		{InventoryNumber: "T-1", Name: "Электросамокат"},
		{InventoryNumber: "T-2", Name: "Электросамокат"},
	}
	for range 2 {
		entries, err := fleet.Inventory()
		if err != nil || !slices.Equal(entries, want) {
			t.Fatalf("Inventory() = %v, %v; want %v", entries, err, want)
		}
		entries[0].Name = "Подмена"
	}
	report, err := fleet.Report()
	wantReport := application.Report{
		TotalAssets: 4, TotalItems: 2, TotalTransports: 2,
		ActiveTransports: 1, RejectedTransports: 1,
		TotalDailyEnergyKWh: 6.75, ActiveDailyEnergyKWh: 1.25,
	}
	if err != nil || report != wantReport {
		t.Fatalf("Report() = %+v, %v; want %+v", report, err, wantReport)
	}
	beginner, err := fleet.BeginnerTransports()
	if err != nil || !slices.Equal(beginner, []domain.Vehicle{accepted}) {
		t.Fatalf("выдача новичкам = %v, %v; want только принятый транспорт", beginner, err)
	}
}

func addInventoryObject(t *testing.T, fleet *application.Fleet, kind, number string) error {
	t.Helper()
	switch kind {
	case "transport":
		_, err := fleet.Admit(makeTransport(t, number, true))
		return err
	case "helmet":
		return fleet.AddItem(makeHelmet(t, number))
	default:
		return fleet.AddItem(makeStation(t, number))
	}
}

func TestInventoryNumbersAreUniqueAcrossObjectTypes(t *testing.T) {
	t.Parallel()
	kinds := []string{"transport", "helmet", "station"}
	for _, first := range kinds {
		for _, second := range kinds {
			t.Run(first+"/"+second, func(t *testing.T) {
				t.Parallel()
				fleet := application.NewFleet(domain.ServiceCenter{}, &memory.FleetRepository{})
				if err := addInventoryObject(t, fleet, first, "I-1"); err != nil {
					t.Fatal(err)
				}
				before, err := fleet.Inventory()
				if err != nil {
					t.Fatal(err)
				}
				if err := addInventoryObject(t, fleet, second, " I-1 "); !errors.Is(err, application.ErrDuplicateInventoryNumber) {
					t.Fatalf("повтор номера: %v, want ErrDuplicateInventoryNumber", err)
				}
				after, err := fleet.Inventory()
				if err != nil || !slices.Equal(before, after) {
					t.Fatalf("повтор номера изменил ведомость: %v, %v; было %v", after, err, before)
				}
			})
		}
	}
}

func TestCannotBypassInspectionUsingAddItem(t *testing.T) {
	t.Parallel()
	repository := &repositoryStub{}
	inspector := inspectorFunc(func(domain.Vehicle) (bool, error) {
		t.Error("неверный сценарий должен отклоняться до техосмотра")
		return true, nil
	})
	fleet := application.NewFleet(inspector, repository)
	for _, transport := range []domain.Vehicle{makeTransport(t, "T-1", true), makeElectricTransport(t, "T-2", true, 1)} {
		if err := fleet.AddItem(transport); !errors.Is(err, application.ErrTransportRequiresInspection) {
			t.Errorf("AddItem(transport) = %v, want ErrTransportRequiresInspection", err)
		}
	}
	if repository.addCalls != 0 {
		t.Fatal("обход техосмотра записал транспорт на баланс")
	}
}

func TestAddItemInvalidAndStorageErrors(t *testing.T) {
	t.Parallel()
	storageErr := errors.New("хранилище недоступно")
	repository := &repositoryStub{addErr: storageErr}
	fleet := application.NewFleet(domain.ServiceCenter{}, repository)
	for _, item := range []domain.InventoryItem{nil, domain.Helmet{}, domain.ChargingStation{}} {
		if err := fleet.AddItem(item); !errors.Is(err, domain.ErrInvalidItem) {
			t.Errorf("AddItem(%T) = %v, want ErrInvalidItem", item, err)
		}
	}
	if repository.addCalls != 0 {
		t.Fatal("невалидные вещи переданы в хранилище")
	}
	if err := fleet.AddItem(makeHelmet(t, "H-1")); !errors.Is(err, storageErr) {
		t.Fatalf("AddItem() = %v, want storageErr", err)
	}
	if len(repository.admissions) != 0 {
		t.Fatal("ошибка записи создала запись на балансе")
	}
}

func TestInventoryEmptyAndStorageError(t *testing.T) {
	t.Parallel()
	repository := &repositoryStub{}
	fleet := application.NewFleet(domain.ServiceCenter{}, repository)
	entries, err := fleet.Inventory()
	if err != nil || len(entries) != 0 {
		t.Fatalf("пустая ведомость: %v, %v", entries, err)
	}
	storageErr := errors.New("чтение недоступно")
	repository.listErr = storageErr
	entries, err = fleet.Inventory()
	if !errors.Is(err, storageErr) || entries != nil {
		t.Fatalf("Inventory() = %v, %v; want nil, storageErr", entries, err)
	}
}
