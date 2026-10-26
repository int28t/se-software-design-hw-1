package cli_test

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"escooter-sharing/internal/application"
	"escooter-sharing/internal/cli"
	"escooter-sharing/internal/domain"
)

func TestRunBeginners(t *testing.T) {
	t.Parallel()
	input := strings.Join([]string{
		"1", "B-5", "Сложный велосипед", "да", "5",
		"1", "B-6", "Простой велосипед", "да", "6",
		"4", "E-10", "Простой электросамокат", "да", "10", "1.25",
		"4", "R-10", "Неисправный электросамокат", "нет", "10", "2.5",
		"5", "0", "",
	}, "\n")
	_, output := runSession(t, input)
	_, listing, found := strings.Cut(output, "Техника для новичков:\n")
	if !found {
		t.Fatalf("список не выведен:\n%s", output)
	}
	for _, want := range []string{
		"B-6 — Простой велосипед (простота: 6/10)",
		"E-10 — Простой электросамокат (простота: 10/10)",
	} {
		if !strings.Contains(listing, want) {
			t.Errorf("нет строки %q в списке:\n%s", want, listing)
		}
	}
	if strings.Contains(listing, "B-5 —") || strings.Contains(listing, "R-10 —") {
		t.Fatalf("в список попала неподходящая техника:\n%s", listing)
	}
}

func TestRunNoBeginnerTransports(t *testing.T) {
	t.Parallel()
	for _, input := range []string{
		"5\n0\n",
		"1\nB-5\nВелосипед\nда\n5\n1\nR-10\nНеисправный\nнет\n10\n5\n0\n",
	} {
		_, output := runSession(t, input)
		if !strings.Contains(output, "Нет доступной техники для новичков.") {
			t.Fatalf("нет сообщения о пустом списке:\n%s", output)
		}
	}
}

func TestRunInvalidBeginnerEase(t *testing.T) {
	t.Parallel()
	for _, command := range []string{"1", "4", "9"} {
		for _, ease := range []string{"", "abc", "1.5", "-1", "0", "11", "999999999999999999999999"} {
			t.Run(command+"/"+ease, func(t *testing.T) {
				t.Parallel()
				input := fmt.Sprintf("%s\nT-1\nТранспорт\nда\n%s\n0\n", command, ease)
				fleet, output := runSession(t, input)
				if !strings.Contains(output, "Ошибка:") {
					t.Fatalf("нет ошибки для простоты %q:\n%s", ease, output)
				}
				report, err := fleet.Report()
				if err != nil || report.TotalTransports != 0 {
					t.Fatalf("ошибочный ввод изменил баланс: %+v, %v", report, err)
				}
			})
		}
	}
}

func TestRunBeginnerListStorageError(t *testing.T) {
	t.Parallel()
	fleet := application.NewFleet(domain.ServiceCenter{}, reportFailureRepository{errors.New("хранилище недоступно")})
	var output bytes.Buffer
	if err := cli.Run(fleet, strings.NewReader("5\n0\n"), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "Ошибка:") || !strings.Contains(output.String(), "хранилище недоступно") {
		t.Fatalf("нет сообщения об ошибке:\n%s", output.String())
	}
	if strings.Contains(output.String(), "Нет доступной техники для новичков.") {
		t.Fatal("ошибка хранилища выдана за пустой список")
	}
}
