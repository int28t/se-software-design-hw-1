package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"escooter-sharing/internal/application"
	"escooter-sharing/internal/cli"
	"escooter-sharing/internal/domain"
	"escooter-sharing/internal/infrastructure/memory"
)

func TestRunCreatesConcreteVehicleKinds(t *testing.T) {
	t.Parallel()
	repository := &memory.FleetRepository{}
	fleet := application.NewFleet(domain.ServiceCenter{}, repository)
	input := strings.Join([]string{
		"1", "B-1", "Велосипед", "да", "5",
		"4", "S-1", "Электросамокат", "да", "6", "1.25",
		"9", "E-1", "Электровелосипед", "да", "10", "2,5",
		"3", "5", "8", "0", "",
	}, "\n")
	var output bytes.Buffer
	if err := cli.Run(fleet, strings.NewReader(input), &output); err != nil {
		t.Fatal(err)
	}
	records, err := repository.List()
	if err != nil || len(records) != 3 {
		t.Fatalf("баланс = %v, %v; want три объекта", records, err)
	}
	if _, ok := records[0].Item().(domain.Bicycle); !ok {
		t.Fatalf("команда 1 создала %T, want Bicycle", records[0].Item())
	}
	if _, ok := records[1].Item().(domain.ElectricBicycle); !ok {
		t.Fatalf("команда 9 создала %T, want ElectricBicycle", records[1].Item())
	}
	if _, ok := records[2].Item().(domain.Scooter); !ok {
		t.Fatalf("команда 4 создала %T, want Scooter", records[2].Item())
	}
	report, err := fleet.Report()
	want := application.Report{
		TotalAssets: 3, TotalTransports: 3, ActiveTransports: 3,
		TotalDailyEnergyKWh: 3.75, ActiveDailyEnergyKWh: 3.75,
	}
	if err != nil || report != want {
		t.Fatalf("Report() = %+v, %v; want %+v", report, err, want)
	}
	for _, line := range []string{
		"E-1 — Электровелосипед (простота: 10/10)",
		"S-1 — Электросамокат (простота: 6/10)",
		"B-1 — Велосипед\n",
	} {
		if !strings.Contains(output.String(), line) {
			t.Errorf("нет строки %q в выводе:\n%s", line, output.String())
		}
	}
}

func TestRunElectricBicycleRejectedAndIncompleteInput(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		input string
		count int
	}{
		{name: "отказ техосмотра", input: "9\nE-1\nЭлектровелосипед\nнет\n10\n2.5\n5\n0\n", count: 1},
		{name: "EOF перед энергией", input: "9\nE-1\nЭлектровелосипед\nда\n6\n"},
		{name: "некорректная энергия", input: "9\nE-1\nЭлектровелосипед\nда\n6\nNaN\n0\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fleet := application.NewFleet(domain.ServiceCenter{}, &memory.FleetRepository{})
			var output bytes.Buffer
			if err := cli.Run(fleet, strings.NewReader(tt.input), &output); err != nil {
				t.Fatal(err)
			}
			report, err := fleet.Report()
			if err != nil || report.TotalTransports != tt.count || report.ActiveTransports != 0 {
				t.Fatalf("Report() = %+v, %v; want %d на балансе, 0 активных", report, err, tt.count)
			}
			beginner, err := fleet.BeginnerTransports()
			if err != nil || len(beginner) != 0 {
				t.Fatalf("выдача новичкам = %v, %v; want пустой список", beginner, err)
			}
		})
	}
}
