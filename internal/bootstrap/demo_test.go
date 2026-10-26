package bootstrap_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"escooter-sharing/internal/bootstrap"
)

func TestDemoSession(t *testing.T) {
	t.Parallel()
	input, err := os.ReadFile("../../docs/demo/session.txt")
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := bootstrap.Run(bytes.NewReader(input), &output); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Данные хранятся только до завершения программы.",
		"Активный парк пуст.",
		"Нет доступной техники для новичков.",
		"На балансе пока нет объектов.",
		"Всего объектов на балансе: 0\n",
		"Неизвестная команда.",
		"инвентарный номер уже используется: B-6",
		"простота для новичка должна быть от 1 до 10",
		"для технического состояния введите «да» или «нет»",
		"расход энергии должен быть конечным числом не меньше нуля",
		"укажите расход энергии числом, например 2.5",
		"укажите название",
		"укажите простоту целым числом от 1 до 10",
		"укажите инвентарный номер",
		"В прокат не допущен.",
		"Всего объектов на балансе: 6\nВещей на балансе: 2\nТранспорта на балансе: 4\n" +
			"Допущено в активный парк: 3\nОтклонено техосмотром: 1\n" +
			"Суточная энергия всего баланса: 7 кВт·ч/сутки\nСуточная энергия активного транспорта: 1.5 кВт·ч/сутки\n",
		"Допущенный транспорт:\nB-5 — Спортивный велосипед\nB-6 — Городской велосипед\nS-1 — Электросамокат\n\nВышКат",
		"Техника для новичков:\nB-6 — Городской велосипед (простота: 6/10)\nS-1 — Электросамокат (простота: 8/10)\n\nВышКат",
		"Инвентарная ведомость:\nB-5 — Спортивный велосипед\nB-6 — Городской велосипед\nC-1 — Зарядная станция\n" +
			"E-1 — Электровелосипед\nH-1 — Синий шлем\nS-1 — Электросамокат\n\nВышКат",
		"Работа программы завершена.",
	} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("нет ожидаемого результата %q:\n%s", want, output.String())
		}
	}
	if got := strings.Count(output.String(), "Ошибка:"); got != 8 {
		t.Errorf("ошибок добавления = %d, want 8", got)
	}
	if strings.Contains(output.String(), "Ввод завершён.") {
		t.Error("полный сценарий должен завершаться командой 0, а не концом ввода")
	}
}

func TestDemoIncompleteInput(t *testing.T) {
	t.Parallel()
	input, err := os.ReadFile("../../docs/demo/incomplete.txt")
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := bootstrap.Run(bytes.NewReader(input), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "Ввод завершён. Работа программы завершена.") {
		t.Fatalf("нет сообщения о конце ввода:\n%s", output.String())
	}
	if strings.Contains(output.String(), "добавлена на баланс") || strings.Contains(output.String(), "Ошибка:") {
		t.Fatalf("неполный ввод обработан неверно:\n%s", output.String())
	}
}
