package cli_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"escooter-sharing/internal/application"
	"escooter-sharing/internal/cli"
	"escooter-sharing/internal/domain"
	"escooter-sharing/internal/infrastructure/memory"
)

func TestRunReadFailureDoesNotCreatePartialRecord(t *testing.T) {
	t.Parallel()
	streamErr := errors.New("ошибка чтения посреди операции")
	for _, tt := range []struct {
		name   string
		prefix string
	}{
		{name: "велосипед", prefix: "1\nB-1\nВелосипед\nда\n"},
		{name: "самокат", prefix: "4\nS-1\nСамокат\nда\n6\n"},
		{name: "шлем", prefix: "6\nH-1\n"},
		{name: "станция", prefix: "7\nC-1\nСтанция\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fleet := application.NewFleet(domain.ServiceCenter{}, &memory.FleetRepository{})
			input := io.MultiReader(strings.NewReader(tt.prefix), failingReader{streamErr})
			if err := cli.Run(fleet, input, io.Discard); !errors.Is(err, streamErr) {
				t.Fatalf("Run() error = %v, want streamErr", err)
			}
			inventory, err := fleet.Inventory()
			if err != nil || len(inventory) != 0 {
				t.Fatalf("незавершённый ввод изменил баланс: %v, %v", inventory, err)
			}
		})
	}
}

type failOnTextWriter struct {
	text   string
	failed bool
}

func (w *failOnTextWriter) Write(data []byte) (int, error) {
	if strings.Contains(string(data), w.text) {
		w.failed = true
		return 0, io.ErrClosedPipe
	}
	return len(data), nil
}

func TestRunWriteFailurePreservesCommittedStateAndStopsSession(t *testing.T) {
	t.Parallel()
	const bicycleInput = "1\nT-1\nВелосипед\nда\n6\n"
	accepted := application.Report{TotalAssets: 1, TotalTransports: 1, ActiveTransports: 1}
	for _, tt := range []struct {
		name   string
		input  string
		failOn string
		want   application.Report
	}{
		{name: "подтверждение допуска", input: bicycleInput, failOn: "допущен в активный парк", want: accepted},
		{name: "подтверждение отказа", input: "1\nT-1\nВелосипед\nнет\n10\n", failOn: "отклонён техосмотром",
			want: application.Report{TotalAssets: 1, TotalTransports: 1, RejectedTransports: 1}},
		{name: "подтверждение вещи", input: "6\nH-1\nШлем\n", failOn: "добавлена на баланс",
			want: application.Report{TotalAssets: 1, TotalItems: 1}},
		{name: "подтверждение станции", input: "7\nC-1\nСтанция\n2.5\n", failOn: "добавлена на баланс",
			want: application.Report{TotalAssets: 1, TotalItems: 1, TotalDailyEnergyKWh: 2.5}},
		{name: "отчёт", input: bicycleInput + "3\n", failOn: "Всего объектов на балансе", want: accepted},
		{name: "заголовок ведомости", input: bicycleInput + "8\n", failOn: "Инвентарная ведомость:", want: accepted},
		{name: "строка ведомости", input: bicycleInput + "8\n", failOn: "T-1 — Велосипед", want: accepted},
		{name: "заголовок активного парка", input: bicycleInput + "2\n", failOn: "Допущенный транспорт:", want: accepted},
		{name: "строка активного парка", input: bicycleInput + "2\n", failOn: "T-1 — Велосипед", want: accepted},
		{name: "заголовок выдачи новичкам", input: bicycleInput + "5\n", failOn: "Техника для новичков:", want: accepted},
		{name: "строка выдачи новичкам", input: bicycleInput + "5\n", failOn: "T-1 — Велосипед", want: accepted},
		{name: "пустой парк", input: "2\n", failOn: "Активный парк пуст."},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fleet := application.NewFleet(domain.ServiceCenter{}, &memory.FleetRepository{})
			output := &failOnTextWriter{text: tt.failOn}
			input := tt.input + "6\nH-2\nСледующая вещь\n0\n"
			if err := cli.Run(fleet, strings.NewReader(input), output); !errors.Is(err, io.ErrClosedPipe) {
				t.Fatalf("Run() error = %v, want io.ErrClosedPipe", err)
			}
			if !output.failed {
				t.Fatalf("сценарий не дошёл до ошибки вывода %q", tt.failOn)
			}
			report, err := fleet.Report()
			if err != nil || report != tt.want {
				t.Fatalf("ошибка вывода изменила баланс или сессия продолжилась: %+v, %v; want %+v", report, err, tt.want)
			}
		})
	}
}

func TestRunActiveFleetStorageErrorIsNotEmptyFleet(t *testing.T) {
	t.Parallel()
	fleet := application.NewFleet(domain.ServiceCenter{}, reportFailureRepository{errors.New("хранилище недоступно")})
	var output bytes.Buffer
	if err := cli.Run(fleet, strings.NewReader("2\n0\n"), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "Ошибка: активный парк: хранилище недоступно") {
		t.Fatalf("нет сообщения об ошибке:\n%s", output.String())
	}
	if strings.Contains(output.String(), "Активный парк пуст.") {
		t.Fatal("ошибка хранилища выдана за пустой парк")
	}
}
