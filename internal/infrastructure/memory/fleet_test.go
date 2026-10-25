package memory_test

import (
	"errors"
	"testing"

	"escooter-sharing/internal/application"
	"escooter-sharing/internal/domain"
	"escooter-sharing/internal/infrastructure/memory"
)

func TestFleetRepositoryPreservesRecordsAndRejectsDuplicate(t *testing.T) {
	t.Parallel()
	repository := &memory.FleetRepository{}
	fleet := application.NewFleet(domain.ServiceCenter{}, repository)
	for _, tc := range []struct {
		number      string
		serviceable bool
	}{{"S-2", false}, {"S-1", true}} {
		transport, err := domain.NewTransport(tc.number, "Самокат", tc.serviceable)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fleet.Admit(transport); err != nil {
			t.Fatal(err)
		}
	}

	duplicate, err := domain.NewTransport(" S-1 ", "Другой транспорт", false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fleet.Admit(duplicate); !errors.Is(err, application.ErrDuplicateInventoryNumber) {
		t.Fatalf("дубликат: %v, want ErrDuplicateInventoryNumber", err)
	}
	admissions, err := repository.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(admissions) != 2 {
		t.Fatalf("records = %d, want 2", len(admissions))
	}
	if admissions[0].Transport().InventoryNumber() != "S-1" || !admissions[0].Approved() || admissions[0].Transport().Name() != "Самокат" {
		t.Fatalf("перезаписана или нарушена сортировка S-1: %+v", admissions[0])
	}
	if admissions[1].Transport().InventoryNumber() != "S-2" || admissions[1].Approved() {
		t.Fatalf("не сохранён отказ S-2: %+v", admissions[1])
	}
	admissions[0] = application.Admission{}
	admissions, err = repository.List()
	if err != nil || !admissions[0].Approved() {
		t.Fatalf("изменение среза повредило хранилище: %v, %v", admissions, err)
	}
}

func TestFleetRepositoryRejectsZeroAdmission(t *testing.T) {
	t.Parallel()
	repository := &memory.FleetRepository{}
	if err := repository.Add(application.Admission{}); !errors.Is(err, domain.ErrInvalidTransport) {
		t.Fatalf("Add() error = %v, want ErrInvalidTransport", err)
	}
	admissions, err := repository.List()
	if err != nil || len(admissions) != 0 {
		t.Fatalf("List() = %v, %v; want пустое хранилище", admissions, err)
	}
}
