package cli_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"escooter-sharing/internal/application"
	"escooter-sharing/internal/cli"
	"escooter-sharing/internal/domain"
	"escooter-sharing/internal/infrastructure/memory"
)

func TestRunAdmissionAndActiveFleet(t *testing.T) {
	t.Parallel()
	repository := &memory.FleetRepository{}
	fleet := application.NewFleet(domain.ServiceCenter{}, repository)
	input := "1\nS-2\nКрасный самокат\nнет\n6\n1\nS-1\nСиний самокат\nДА\n6\n1\nS-1\nПодмена\nнет\n6\n2\n0\n"
	var output bytes.Buffer
	if err := cli.Run(fleet, strings.NewReader(input), &output); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"отклонён техосмотром", "допущен в активный парк", "инвентарный номер уже используется"} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("нет сообщения %q в выводе:\n%s", want, output.String())
		}
	}
	_, activeReport, found := strings.Cut(output.String(), "Допущенный транспорт:\n")
	if !found || !strings.Contains(activeReport, "S-1 — Синий самокат") || strings.Contains(activeReport, "S-2") {
		t.Fatalf("неверный отчёт об активном парке:\n%s", output.String())
	}
	admissions, err := repository.List()
	if err != nil || len(admissions) != 2 {
		t.Fatalf("баланс: %v, %v; want две записи", admissions, err)
	}
}

func TestRunInvalidInputAndEOF(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "пустой ввод", want: "Ввод завершён. Работа программы завершена."},
		{name: "выход", input: "0\n", want: "Работа программы завершена."},
		{name: "неизвестная команда", input: "99\n0\n", want: "Неизвестная команда"},
		{name: "пустой парк", input: "2\n0\n", want: "Активный парк пуст"},
		{name: "пустой номер", input: "1\n \nСамокат\nда\n6\n0\n", want: "укажите инвентарный номер"},
		{name: "пустое название", input: "1\nS-1\n\nда\n6\n0\n", want: "укажите название"},
		{name: "неверное состояние", input: "1\nS-1\nСамокат\nможет быть\n0\n", want: "введите «да» или «нет»"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fleet, output := runSession(t, tt.input)
			if !strings.Contains(output, tt.want) {
				t.Fatalf("нет сообщения %q:\n%s", tt.want, output)
			}
			admissions, err := fleet.Inventory()
			if err != nil || len(admissions) != 0 {
				t.Fatalf("неполный/ошибочный ввод изменил баланс: %v, %v", admissions, err)
			}
		})
	}
}

func TestRunIncompleteAdditionDoesNotChangeBalance(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		lines []string
	}{
		{name: "велосипед", lines: []string{"1", "B-1", "Велосипед", "да", "6"}},
		{name: "самокат", lines: []string{"4", "S-1", "Самокат", "да", "6", "1.25"}},
		{name: "электровелосипед", lines: []string{"9", "E-1", "Электровелосипед", "да", "6", "2.5"}},
		{name: "шлем", lines: []string{"6", "H-1", "Шлем"}},
		{name: "станция", lines: []string{"7", "C-1", "Станция", "3.5"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for field := range tt.lines[1:] {
				t.Run(fmt.Sprintf("до_поля_%d", field+1), func(t *testing.T) {
					t.Parallel()
					input := strings.Join(tt.lines[:field+1], "\n") + "\n"
					fleet, output := runSession(t, input)
					if !strings.Contains(output, "Ввод завершён. Работа программы завершена.") || strings.Contains(output, "Ошибка:") {
						t.Fatalf("незавершённый ввод обработан неверно:\n%s", output)
					}
					entries, err := fleet.Inventory()
					if err != nil || len(entries) != 0 {
						t.Fatalf("незавершённый ввод изменил баланс: %v, %v", entries, err)
					}
					report, err := fleet.Report()
					if err != nil || report != (application.Report{}) {
						t.Fatalf("отчёт после незавершённого ввода: %+v, %v; want пустой", report, err)
					}
					beginners, err := fleet.BeginnerTransports()
					if err != nil || len(beginners) != 0 {
						t.Fatalf("выдача новичкам после незавершённого ввода: %v, %v; want пустой список", beginners, err)
					}
				})
			}
		})
	}
}

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

func TestRunStreamErrors(t *testing.T) {
	t.Parallel()
	streamErr := errors.New("поток недоступен")
	tests := []struct {
		name   string
		input  io.Reader
		output io.Writer
		want   error
	}{
		{name: "ошибка чтения", input: failingReader{streamErr}, output: io.Discard, want: streamErr},
		{name: "ошибка записи", input: strings.NewReader("0\n"), output: failingWriter{streamErr}, want: streamErr},
		{name: "EOF записи не означает конец ввода", input: strings.NewReader("0\n"), output: failingWriter{io.EOF}, want: io.EOF},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fleet := application.NewFleet(domain.ServiceCenter{}, &memory.FleetRepository{})
			if err := cli.Run(fleet, tt.input, tt.output); !errors.Is(err, tt.want) {
				t.Fatalf("Run() error = %v, want %v", err, tt.want)
			}
		})
	}
}
