package bootstrap_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"go.uber.org/dig"

	"escooter-sharing/internal/application"
	"escooter-sharing/internal/bootstrap"
	"escooter-sharing/internal/domain"
)

type inspectorFunc func(domain.Vehicle) (bool, error)

func (f inspectorFunc) Inspect(transport domain.Vehicle) (bool, error) {
	return f(transport)
}

func newContainer(t *testing.T) *dig.Container {
	t.Helper()
	container, err := bootstrap.NewContainer()
	if err != nil {
		t.Fatal(err)
	}
	return container
}

func bicycle(t *testing.T, serviceable bool) domain.Bicycle {
	t.Helper()
	transport, err := domain.NewBicycle("B-1", "Велосипед", serviceable, 6)
	if err != nil {
		t.Fatal(err)
	}
	return transport
}

func TestContainerSharesFleetAndRepository(t *testing.T) {
	t.Parallel()
	container := newContainer(t)
	var first *application.Fleet
	if err := container.Invoke(func(fleet *application.Fleet) error {
		first = fleet
		_, err := fleet.Admit(bicycle(t, true))
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if err := container.Invoke(func(fleet *application.Fleet, repository application.Repository) error {
		if fleet != first {
			t.Error("контейнер создал повторный экземпляр Fleet")
		}
		records, err := repository.List()
		if err != nil {
			return err
		}
		if len(records) != 1 || records[0].Item().InventoryNumber() != "B-1" || !records[0].Approved() {
			t.Errorf("Fleet и Repository не разделяют баланс: %v", records)
		}
		report, err := fleet.Report()
		if err == nil && (report.TotalAssets != 1 || report.ActiveTransports != 1) {
			t.Errorf("неверный отчёт собранного приложения: %+v", report)
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}
}

func TestContainersHaveIndependentBalances(t *testing.T) {
	t.Parallel()
	first := newContainer(t)
	if err := first.Invoke(func(fleet *application.Fleet) error {
		_, err := fleet.Admit(bicycle(t, true))
		return err
	}); err != nil {
		t.Fatal(err)
	}
	second := newContainer(t)
	if err := second.Invoke(func(fleet *application.Fleet) error {
		report, err := fleet.Report()
		if err == nil && report != (application.Report{}) {
			t.Errorf("баланс попал из другого контейнера: %+v", report)
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}
}

func TestContainerInspectionSubstitution(t *testing.T) {
	t.Parallel()
	inspectionErr := errors.New("техосмотр недоступен")
	for _, tt := range []struct {
		name     string
		approved bool
		err      error
	}{
		{name: "принять", approved: true},
		{name: "отклонить"},
		{name: "ошибка", approved: true, err: inspectionErr},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			container := newContainer(t)
			transport := bicycle(t, !tt.approved)
			calls := 0
			fake := inspectorFunc(func(got domain.Vehicle) (bool, error) {
				calls++
				if got != transport {
					t.Errorf("Inspector получил %+v, want %+v", got, transport)
				}
				return tt.approved, tt.err
			})
			if err := container.Decorate(func() application.Inspector { return fake }); err != nil {
				t.Fatal(err)
			}
			if err := container.Invoke(func(fleet *application.Fleet, repository application.Repository) error {
				record, err := fleet.Admit(transport)
				if !errors.Is(err, tt.err) {
					t.Errorf("Admit() error = %v, want %v", err, tt.err)
				}
				if calls != 1 {
					t.Errorf("Inspector вызван %d раз, want 1", calls)
				}
				records, listErr := repository.List()
				if listErr != nil {
					return listErr
				}
				if tt.err != nil {
					if len(records) != 0 || record.Approved() {
						t.Error("ошибка техосмотра создала допуск или запись на балансе")
					}
					return nil
				}
				if len(records) != 1 || record.Approved() != tt.approved {
					t.Errorf("решение подмены не сохранено: %+v, записи: %v", record, records)
				}
				active, err := fleet.ActiveTransports()
				if err == nil && ((tt.approved && len(active) != 1) || (!tt.approved && len(active) != 0)) {
					t.Errorf("допуск не соответствует решению подмены: %v", active)
				}
				return err
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestContainerPreservesConstructorError(t *testing.T) {
	t.Parallel()
	container := newContainer(t)
	storageErr := errors.New("невозможно создать хранилище")
	if err := container.Decorate(func() (application.Repository, error) {
		return nil, storageErr
	}); err != nil {
		t.Fatal(err)
	}
	called := false
	err := container.Invoke(func(*application.Fleet) { called = true })
	if !errors.Is(err, storageErr) || called {
		t.Fatalf("Invoke() error = %v, callback called = %v; want storageErr, false", err, called)
	}
}

func TestRunUsesContainerGraph(t *testing.T) {
	t.Parallel()
	input := strings.Join([]string{
		"4", "S-1", "Самокат", "да", "6", "1.25",
		"9", "E-1", "Электровелосипед", "нет", "10", "2.5",
		"6", "H-1", "Шлем",
		"3", "5", "8", "0", "",
	}, "\n")
	var output bytes.Buffer
	if err := bootstrap.Run(strings.NewReader(input), &output); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Всего объектов на балансе: 3\n",
		"Вещей на балансе: 1\n",
		"Допущено в активный парк: 1\n",
		"Отклонено техосмотром: 1\n",
		"Суточная энергия всего баланса: 3.75 кВт·ч/сутки\n",
		"Суточная энергия активного транспорта: 1.25 кВт·ч/сутки\n",
		"Техника для новичков:\nS-1 — Самокат (простота: 6/10)\n",
		"Инвентарная ведомость:\nE-1 — Электровелосипед\nH-1 — Шлем\nS-1 — Самокат\n",
	} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("нет строки %q:\n%s", want, output.String())
		}
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestRunPreservesSessionError(t *testing.T) {
	t.Parallel()
	if err := bootstrap.Run(strings.NewReader("0\n"), failingWriter{}); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("Run() error = %v, want io.ErrClosedPipe", err)
	}
}
