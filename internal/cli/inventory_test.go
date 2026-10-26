package cli_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"escooter-sharing/internal/application"
	"escooter-sharing/internal/cli"
	"escooter-sharing/internal/domain"
)

func TestRunMixedInventory(t *testing.T) {
	t.Parallel()
	input := strings.Join([]string{
		"4", "T-1", "Электросамокат", "да", "6", "1.25",
		"1", "T-2", "Неисправный велосипед", "нет", "10",
		"6", "H-1", "Синий шлем",
		"7", "C-1", "Зарядная станция", "3,5",
		"8", "3", "5", "0", "",
	}, "\n")
	_, output := runSession(t, input)
	_, inventory, found := strings.Cut(output, "Инвентарная ведомость:\n")
	if !found {
		t.Fatalf("нет ведомости:\n%s", output)
	}
	inventory, _, _ = strings.Cut(inventory, "\nВышКат")
	wantInventory := "C-1 — Зарядная станция\nH-1 — Синий шлем\nT-1 — Электросамокат\nT-2 — Неисправный велосипед\n"
	if inventory != wantInventory {
		t.Fatalf("ведомость:\n%s\nwant:\n%s", inventory, wantInventory)
	}
	for _, want := range []string{
		"Всего объектов на балансе: 4\n",
		"Вещей на балансе: 2\n",
		"Транспорта на балансе: 2\n",
		"Допущено в активный парк: 1\n",
		"Отклонено техосмотром: 1\n",
		"Суточная энергия всего баланса: 4.75 кВт·ч/сутки\n",
		"Суточная энергия активного транспорта: 1.25 кВт·ч/сутки\n",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("нет строки %q:\n%s", want, output)
		}
	}
	_, beginners, found := strings.Cut(output, "Техника для новичков:\n")
	if !found || !strings.HasPrefix(beginners, "T-1 — Электросамокат (простота: 6/10)\n\nВышКат") {
		t.Fatalf("неверная выдача новичкам:\n%s", output)
	}
}

func TestRunInventoryDuplicateNumber(t *testing.T) {
	t.Parallel()
	input := "6\nI-1\nШлем\n1\nI-1\nВелосипед\nда\n6\n7\nI-1\nСтанция\n2\n8\n0\n"
	fleet, output := runSession(t, input)
	if strings.Count(output, "инвентарный номер уже используется") != 2 {
		t.Fatalf("нет двух отказов по номеру:\n%s", output)
	}
	entries, err := fleet.Inventory()
	if err != nil || len(entries) != 1 || entries[0].Name != "Шлем" {
		t.Fatalf("дубликаты изменили баланс: %v, %v", entries, err)
	}
}

func TestRunInvalidItemInput(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		input string
	}{
		{name: "шлем без номера", input: "6\n \nШлем\n0\n"},
		{name: "шлем без названия", input: "6\nH-1\n\n0\n"},
		{name: "станция без номера", input: "7\n \nСтанция\n0\n"},
		{name: "станция без названия", input: "7\nC-1\n\n0\n"},
		{name: "отрицательная энергия", input: "7\nC-1\nСтанция\n-1\n0\n"},
		{name: "NaN", input: "7\nC-1\nСтанция\nNaN\n0\n"},
		{name: "бесконечность", input: "7\nC-1\nСтанция\nInf\n0\n"},
		{name: "не число", input: "7\nC-1\nСтанция\nabc\n0\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fleet, output := runSession(t, tt.input)
			if !strings.Contains(output, "Ошибка:") {
				t.Fatalf("нет сообщения об ошибке:\n%s", output)
			}
			entries, err := fleet.Inventory()
			if err != nil || len(entries) != 0 {
				t.Fatalf("некорректный ввод изменил баланс: %v, %v", entries, err)
			}
		})
	}
}

func TestRunEmptyInventory(t *testing.T) {
	t.Parallel()
	_, output := runSession(t, "8\n0\n")
	if !strings.Contains(output, "На балансе пока нет объектов.") {
		t.Fatalf("нет сообщения о пустой ведомости:\n%s", output)
	}
}

func TestRunInventoryStorageErrors(t *testing.T) {
	t.Parallel()
	for _, input := range []string{"8\n0\n", "6\nH-1\nШлем\n0\n"} {
		fleet := application.NewFleet(domain.ServiceCenter{}, reportFailureRepository{errors.New("хранилище недоступно")})
		var output bytes.Buffer
		if err := cli.Run(fleet, strings.NewReader(input), &output); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(output.String(), "Ошибка:") || !strings.Contains(output.String(), "хранилище недоступно") {
			t.Fatalf("нет сообщения об ошибке:\n%s", output.String())
		}
		if strings.Contains(output.String(), "добавлена на баланс") || strings.Contains(output.String(), "На балансе пока нет объектов.") {
			t.Fatalf("ошибка выдана за успешный результат:\n%s", output.String())
		}
	}
}
