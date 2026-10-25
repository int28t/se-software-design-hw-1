package application_test

import (
	"errors"
	"slices"
	"testing"

	"escooter-sharing/internal/application"
	"escooter-sharing/internal/domain"
	"escooter-sharing/internal/infrastructure/memory"
)

func TestBeginnerTransportsFiltersAndPreservesOrder(t *testing.T) {
	t.Parallel()
	repository := &memory.FleetRepository{}
	fleet := application.NewFleet(domain.ServiceCenter{}, repository)
	for _, tc := range []struct {
		number      string
		ease        int
		serviceable bool
		electric    bool
	}{
		{number: "E-6", ease: 6, serviceable: true, electric: true},
		{number: "R-10", ease: 10, serviceable: false},
		{number: "B-5", ease: 5, serviceable: true},
		{number: "B-6", ease: 6, serviceable: true},
		{number: "B-10", ease: 10, serviceable: true},
		{number: "E-10", ease: 10, serviceable: false, electric: true},
		{number: "B-1", ease: 1, serviceable: true},
	} {
		var transport domain.Vehicle
		var err error
		if tc.electric {
			transport, err = domain.NewElectricTransport(tc.number, "Электротранспорт", tc.serviceable, tc.ease, 1)
		} else {
			transport, err = domain.NewTransport(tc.number, "Велосипед", tc.serviceable, tc.ease)
		}
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fleet.Admit(transport); err != nil {
			t.Fatal(err)
		}
	}

	for range 2 {
		transports, err := fleet.BeginnerTransports()
		if err != nil {
			t.Fatal(err)
		}
		var numbers []string
		for _, transport := range transports {
			numbers = append(numbers, transport.InventoryNumber())
		}
		want := []string{"B-10", "B-6", "E-6"}
		if !slices.Equal(numbers, want) {
			t.Fatalf("BeginnerTransports() = %v, want %v", numbers, want)
		}
		transports[0] = domain.Transport{}
	}
	admissions, err := repository.List()
	if err != nil || len(admissions) != 7 {
		t.Fatalf("фильтрация изменила баланс: %v, %v", admissions, err)
	}
}

func TestBeginnerTransportsEmpty(t *testing.T) {
	t.Parallel()
	fleet := application.NewFleet(domain.ServiceCenter{}, &repositoryStub{})
	transports, err := fleet.BeginnerTransports()
	if err != nil || len(transports) != 0 {
		t.Fatalf("BeginnerTransports() = %v, %v; want пустой список", transports, err)
	}
}

func TestBeginnerTransportsPreservesStorageError(t *testing.T) {
	t.Parallel()
	storageErr := errors.New("хранилище недоступно")
	fleet := application.NewFleet(domain.ServiceCenter{}, &repositoryStub{listErr: storageErr})
	transports, err := fleet.BeginnerTransports()
	if !errors.Is(err, storageErr) || transports != nil {
		t.Fatalf("BeginnerTransports() = %v, %v; want nil, storageErr", transports, err)
	}
}
