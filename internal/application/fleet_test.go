package application_test

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"escooter-sharing/internal/application"
	"escooter-sharing/internal/domain"
)

type inspectorFunc func(domain.Vehicle) (bool, error)

func (f inspectorFunc) Inspect(transport domain.Vehicle) (bool, error) {
	return f(transport)
}

type repositoryStub struct {
	admissions []application.Admission
	addErr     error
	listErr    error
	addCalls   int
}

func (r *repositoryStub) Add(admission application.Admission) error {
	r.addCalls++
	if r.addErr != nil {
		return r.addErr
	}
	r.admissions = append(r.admissions, admission)
	return nil
}

func (r *repositoryStub) List() ([]application.Admission, error) {
	return slices.Clone(r.admissions), r.listErr
}

func makeTransport(t *testing.T, number string, serviceable bool) domain.Transport {
	t.Helper()
	transport, err := domain.NewTransport(number, "Самокат", serviceable)
	if err != nil {
		t.Fatal(err)
	}
	return transport
}

func TestAdmitUsesInjectedInspectionBeforeSaving(t *testing.T) {
	t.Parallel()
	for _, approved := range []bool{true, false} {
		t.Run(fmt.Sprintf("решение_%v", approved), func(t *testing.T) {
			t.Parallel()
			repository := &repositoryStub{}
			transport := makeTransport(t, "S-1", !approved)
			calls := 0
			inspector := inspectorFunc(func(got domain.Vehicle) (bool, error) {
				calls++
				if got != transport {
					t.Errorf("Inspect() transport = %+v, want %+v", got, transport)
				}
				if repository.addCalls != 0 {
					t.Error("транспорт записан до техосмотра")
				}
				return approved, nil
			})
			fleet := application.NewFleet(inspector, repository)

			admission, err := fleet.Admit(transport)
			if err != nil {
				t.Fatal(err)
			}
			if calls != 1 || repository.addCalls != 1 || len(repository.admissions) != 1 {
				t.Fatalf("Inspect calls = %d, Add calls = %d, records = %d", calls, repository.addCalls, len(repository.admissions))
			}
			if admission.Transport() != transport || admission.Approved() != approved || repository.admissions[0] != admission {
				t.Fatalf("неверный результат приёмки: %+v", admission)
			}
			active, err := fleet.ActiveTransports()
			if err != nil {
				t.Fatal(err)
			}
			if approved && !slices.Equal(active, []domain.Vehicle{transport}) {
				t.Fatalf("ActiveTransports() = %v, want принятый транспорт", active)
			}
			if !approved && len(active) != 0 {
				t.Fatalf("отклонённый транспорт попал в активный парк: %v", active)
			}
		})
	}
}

func TestAdmitInvalidTransportDoesNotInspectOrSave(t *testing.T) {
	t.Parallel()
	repository := &repositoryStub{}
	inspector := inspectorFunc(func(domain.Vehicle) (bool, error) {
		t.Error("Inspect не должен вызываться для пустого транспорта")
		return true, nil
	})
	fleet := application.NewFleet(inspector, repository)
	_, err := fleet.Admit(domain.Transport{})
	if !errors.Is(err, domain.ErrInvalidTransport) || repository.addCalls != 0 {
		t.Fatalf("Admit() error = %v, Add calls = %d", err, repository.addCalls)
	}
}

func TestAdmitInspectionErrorDoesNotSave(t *testing.T) {
	t.Parallel()
	inspectionErr := errors.New("техосмотр недоступен")
	repository := &repositoryStub{}
	inspector := inspectorFunc(func(domain.Vehicle) (bool, error) {
		return true, fmt.Errorf("проверка: %w", inspectionErr)
	})
	fleet := application.NewFleet(inspector, repository)
	admission, err := fleet.Admit(makeTransport(t, "S-1", true))
	if !errors.Is(err, inspectionErr) || admission.Approved() || repository.addCalls != 0 {
		t.Fatalf("Admit() = %+v, %v; Add calls = %d", admission, err, repository.addCalls)
	}
}

func TestRepositoryErrorsArePreserved(t *testing.T) {
	t.Parallel()
	storageErr := errors.New("хранилище недоступно")
	repository := &repositoryStub{addErr: storageErr, listErr: storageErr}
	fleet := application.NewFleet(domain.ServiceCenter{}, repository)
	admission, err := fleet.Admit(makeTransport(t, "S-1", true))
	if !errors.Is(err, storageErr) || admission.Approved() || len(repository.admissions) != 0 {
		t.Fatalf("Admit() = %+v, %v; records = %d", admission, err, len(repository.admissions))
	}
	active, err := fleet.ActiveTransports()
	if !errors.Is(err, storageErr) || active != nil {
		t.Fatalf("ActiveTransports() = %v, %v; want nil, storageErr", active, err)
	}
}

func TestActiveTransportsFiltersMixedFleet(t *testing.T) {
	t.Parallel()
	repository := &repositoryStub{}
	fleet := application.NewFleet(domain.ServiceCenter{}, repository)
	active, err := fleet.ActiveTransports()
	if err != nil || len(active) != 0 {
		t.Fatalf("пустой парк: %v, %v", active, err)
	}
	accepted := makeTransport(t, "S-1", true)
	rejected := makeTransport(t, "S-2", false)
	for _, transport := range []domain.Transport{accepted, rejected} {
		if _, err := fleet.Admit(transport); err != nil {
			t.Fatal(err)
		}
	}
	active, err = fleet.ActiveTransports()
	if err != nil || !slices.Equal(active, []domain.Vehicle{accepted}) {
		t.Fatalf("ActiveTransports() = %v, %v; want только S-1", active, err)
	}
	active[0] = rejected
	active, err = fleet.ActiveTransports()
	if err != nil || !slices.Equal(active, []domain.Vehicle{accepted}) {
		t.Fatalf("изменение результата повредило парк: %v, %v", active, err)
	}
}
