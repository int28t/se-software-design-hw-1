package application_test

import (
	"errors"
	"fmt"
	"math"
	"testing"

	"escooter-sharing/internal/application"
	"escooter-sharing/internal/domain"
)

// additionalConsumer проверяет, что отчёт поддерживает новый тип без switch по видам.
type additionalConsumer struct {
	domain.Transport
	energy float64
}

func (c additionalConsumer) DailyEnergyKWh() float64 { return c.energy }

func makeElectricTransport(t *testing.T, number string, serviceable bool, energy float64) domain.ElectricTransport {
	t.Helper()
	transport, err := domain.NewElectricTransport(number, "Электросамокат", serviceable, energy)
	if err != nil {
		t.Fatal(err)
	}
	return transport
}

func TestReportMixedFleet(t *testing.T) {
	t.Parallel()
	repository := &repositoryStub{}
	fleet := application.NewFleet(domain.ServiceCenter{}, repository)
	for _, transport := range []domain.Vehicle{
		makeTransport(t, "B-1", true),
		makeElectricTransport(t, "E-1", true, 1.25),
		makeElectricTransport(t, "E-2", false, 2.5),
		additionalConsumer{Transport: makeTransport(t, "X-1", true), energy: 6},
	} {
		if _, err := fleet.Admit(transport); err != nil {
			t.Fatal(err)
		}
	}
	want := application.Report{
		TotalTransports: 4, ActiveTransports: 3, RejectedTransports: 1,
		TotalDailyEnergyKWh: 9.75, ActiveDailyEnergyKWh: 7.25,
	}
	for range 2 {
		got, err := fleet.Report()
		if err != nil || got != want {
			t.Fatalf("Report() = %+v, %v; want %+v", got, err, want)
		}
	}
	if len(repository.admissions) != 4 {
		t.Fatal("отчёт изменил баланс")
	}
}

func TestReportEmptyAndNonElectricFleet(t *testing.T) {
	t.Parallel()
	repository := &repositoryStub{}
	fleet := application.NewFleet(domain.ServiceCenter{}, repository)
	got, err := fleet.Report()
	if err != nil || got != (application.Report{}) {
		t.Fatalf("пустой парк: %+v, %v", got, err)
	}
	if _, err := fleet.Admit(makeTransport(t, "B-1", false)); err != nil {
		t.Fatal(err)
	}
	got, err = fleet.Report()
	want := application.Report{TotalTransports: 1, RejectedTransports: 1}
	if err != nil || got != want {
		t.Fatalf("парк без потребления: %+v, %v; want %+v", got, err, want)
	}
}

func TestReportPreservesStorageError(t *testing.T) {
	t.Parallel()
	storageErr := errors.New("хранилище недоступно")
	repository := &repositoryStub{listErr: fmt.Errorf("чтение: %w", storageErr)}
	fleet := application.NewFleet(domain.ServiceCenter{}, repository)
	report, err := fleet.Report()
	if !errors.Is(err, storageErr) || report != (application.Report{}) {
		t.Fatalf("Report() = %+v, %v; want пустой отчёт и storageErr", report, err)
	}
}

func TestReportEnergyOverflow(t *testing.T) {
	t.Parallel()
	repository := &repositoryStub{}
	fleet := application.NewFleet(domain.ServiceCenter{}, repository)
	for _, number := range []string{"E-1", "E-2"} {
		if _, err := fleet.Admit(makeElectricTransport(t, number, true, math.MaxFloat64)); err != nil {
			t.Fatal(err)
		}
	}
	report, err := fleet.Report()
	if !errors.Is(err, application.ErrEnergyOverflow) || report != (application.Report{}) {
		t.Fatalf("Report() = %+v, %v; want пустой отчёт и ErrEnergyOverflow", report, err)
	}
}

func TestAdmitNilDoesNotInspectOrSave(t *testing.T) {
	t.Parallel()
	repository := &repositoryStub{}
	inspector := inspectorFunc(func(domain.Vehicle) (bool, error) {
		t.Error("Inspect не должен вызываться для nil")
		return true, nil
	})
	fleet := application.NewFleet(inspector, repository)
	_, err := fleet.Admit(nil)
	if !errors.Is(err, domain.ErrInvalidTransport) || repository.addCalls != 0 {
		t.Fatalf("Admit(nil) error = %v, Add calls = %d", err, repository.addCalls)
	}
}
