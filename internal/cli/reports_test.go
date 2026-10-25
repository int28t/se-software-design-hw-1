package cli_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"escooter-sharing/internal/application"
	"escooter-sharing/internal/cli"
	"escooter-sharing/internal/domain"
	"escooter-sharing/internal/infrastructure/memory"
)

func TestRunEnergyReport(t *testing.T) {
	t.Parallel()
	repository := &memory.FleetRepository{}
	fleet := application.NewFleet(domain.ServiceCenter{}, repository)
	input := "4\nE-1\nЭлектросамокат\nда\n6\n1,25\n4\nE-2\nЭлектровелосипед\nнет\n6\n2.5\n1\nB-1\nВелосипед\nда\n6\n3\n0\n"
	var output bytes.Buffer
	if err := cli.Run(fleet, strings.NewReader(input), &output); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Транспорта на балансе: 3\n",
		"Допущено в активный парк: 2\n",
		"Отклонено техосмотром: 1\n",
		"Суточная энергия всего баланса: 3.75 кВт·ч/сутки\n",
		"Суточная энергия активного транспорта: 1.25 кВт·ч/сутки\n",
	} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("нет строки %q в выводе:\n%s", want, output.String())
		}
	}
}

func TestRunEmptyReport(t *testing.T) {
	t.Parallel()
	fleet := application.NewFleet(domain.ServiceCenter{}, &memory.FleetRepository{})
	var output bytes.Buffer
	if err := cli.Run(fleet, strings.NewReader("3\n0\n"), &output); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Транспорта на балансе: 0\n",
		"Допущено в активный парк: 0\n",
		"Отклонено техосмотром: 0\n",
		"Суточная энергия всего баланса: 0 кВт·ч/сутки\n",
		"Суточная энергия активного транспорта: 0 кВт·ч/сутки\n",
	} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("нет строки %q в выводе:\n%s", want, output.String())
		}
	}
}

func TestRunInvalidEnergy(t *testing.T) {
	t.Parallel()
	for _, consumption := range []string{"", "abc", "-1", "NaN", "+Inf", "-Inf", "1e400"} {
		t.Run(consumption, func(t *testing.T) {
			t.Parallel()
			repository := &memory.FleetRepository{}
			fleet := application.NewFleet(domain.ServiceCenter{}, repository)
			var output bytes.Buffer
			input := "4\nE-1\nЭлектросамокат\nда\n6\n" + consumption + "\n0\n"
			if err := cli.Run(fleet, strings.NewReader(input), &output); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(output.String(), "Ошибка:") {
				t.Fatalf("нет сообщения об ошибке:\n%s", output.String())
			}
			report, err := fleet.Report()
			if err != nil || report.TotalTransports != 0 {
				t.Fatalf("ошибочный расход изменил парк: %+v, %v", report, err)
			}
		})
	}
}

func TestRunEOFBeforeEnergyDoesNotAdmit(t *testing.T) {
	t.Parallel()
	fleet := application.NewFleet(domain.ServiceCenter{}, &memory.FleetRepository{})
	var output bytes.Buffer
	if err := cli.Run(fleet, strings.NewReader("4\nE-1\nСамокат\nда\n6\n"), &output); err != nil {
		t.Fatal(err)
	}
	report, err := fleet.Report()
	if err != nil || report.TotalTransports != 0 {
		t.Fatalf("неполный ввод изменил парк: %+v, %v", report, err)
	}
}

type reportFailureRepository struct{ err error }

func (r reportFailureRepository) Add(application.Admission) error { return r.err }

func (r reportFailureRepository) List() ([]application.Admission, error) { return nil, r.err }

func TestRunReportStorageError(t *testing.T) {
	t.Parallel()
	fleet := application.NewFleet(domain.ServiceCenter{}, reportFailureRepository{errors.New("хранилище недоступно")})
	var output bytes.Buffer
	if err := cli.Run(fleet, strings.NewReader("3\n0\n"), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "Ошибка: отчёт о парке: хранилище недоступно") {
		t.Fatalf("нет сообщения об ошибке:\n%s", output.String())
	}
	if strings.Contains(output.String(), "Транспорта на балансе:") {
		t.Fatal("при ошибке показан недостоверный нулевой отчёт")
	}
}
