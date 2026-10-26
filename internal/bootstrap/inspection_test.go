package bootstrap_test

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"escooter-sharing/internal/application"
	"escooter-sharing/internal/domain"
)

func TestContainerInspectionFailurePreservesBalanceAndAllowsNextAdmission(t *testing.T) {
	t.Parallel()
	container := newContainer(t)
	inspectionErr := errors.New("временная ошибка техосмотра")
	calls := make(map[string]int)
	fake := inspectorFunc(func(transport domain.Vehicle) (bool, error) {
		number := transport.InventoryNumber()
		calls[number]++
		switch number {
		case "F-1":
			if calls[number] == 1 {
				return true, fmt.Errorf("сервис: %w", inspectionErr)
			}
			return true, nil
		case "R-1":
			return false, nil
		default:
			return true, nil
		}
	})
	if err := container.Decorate(func() application.Inspector { return fake }); err != nil {
		t.Fatal(err)
	}
	var fleet *application.Fleet
	if err := container.Invoke(func(resolved *application.Fleet) { fleet = resolved }); err != nil {
		t.Fatal(err)
	}
	vehicles := make(map[string]domain.Bicycle)
	for _, tc := range []struct {
		number      string
		serviceable bool
		ease        int
	}{
		{number: "A-1", serviceable: false, ease: 6},
		{number: "F-1", serviceable: true, ease: 10},
		{number: "R-1", serviceable: true, ease: 10},
	} {
		transport, err := domain.NewBicycle(tc.number, "Велосипед "+tc.number, tc.serviceable, tc.ease)
		if err != nil {
			t.Fatal(err)
		}
		vehicles[tc.number] = transport
	}
	if record, err := fleet.Admit(vehicles["A-1"]); err != nil || !record.Approved() {
		t.Fatalf("первая приёмка = %+v, %v; want допуск от подмены", record, err)
	}
	beforeInventory, err := fleet.Inventory()
	if err != nil {
		t.Fatal(err)
	}
	beforeReport, err := fleet.Report()
	if err != nil {
		t.Fatal(err)
	}

	if record, err := fleet.Admit(vehicles["F-1"]); !errors.Is(err, inspectionErr) || record.Approved() {
		t.Fatalf("ошибка техосмотра: запись = %+v, ошибка = %v", record, err)
	}
	afterInventory, err := fleet.Inventory()
	if err != nil || !slices.Equal(afterInventory, beforeInventory) {
		t.Fatalf("ошибка изменила баланс: %v, %v; было %v", afterInventory, err, beforeInventory)
	}
	afterReport, err := fleet.Report()
	if err != nil || afterReport != beforeReport {
		t.Fatalf("ошибка изменила отчёт: %+v, %v; было %+v", afterReport, err, beforeReport)
	}
	for _, number := range []string{"F-1", "R-1"} {
		record, err := fleet.Admit(vehicles[number])
		if err != nil || record.Approved() != (number == "F-1") {
			t.Fatalf("приёмка после ошибки %s: %+v, %v", number, record, err)
		}
	}
	report, err := fleet.Report()
	wantReport := application.Report{TotalAssets: 3, TotalTransports: 3, ActiveTransports: 2, RejectedTransports: 1}
	if err != nil || report != wantReport {
		t.Fatalf("итоговый отчёт = %+v, %v; want %+v", report, err, wantReport)
	}
	active, err := fleet.ActiveTransports()
	if err != nil || !slices.Equal(active, []domain.Vehicle{vehicles["A-1"], vehicles["F-1"]}) {
		t.Fatalf("активный парк = %v, %v; want A-1 и F-1", active, err)
	}
	beginner, err := fleet.BeginnerTransports()
	if err != nil || !slices.Equal(beginner, []domain.Vehicle{vehicles["A-1"], vehicles["F-1"]}) {
		t.Fatalf("выдача новичкам = %v, %v; want только допущенные A-1 и F-1", beginner, err)
	}
	for _, tc := range []struct {
		number string
		count  int
	}{{number: "A-1", count: 1}, {number: "F-1", count: 2}, {number: "R-1", count: 1}} {
		if calls[tc.number] != tc.count {
			t.Errorf("Inspector для %s вызван %d раз, want %d", tc.number, calls[tc.number], tc.count)
		}
	}
}
