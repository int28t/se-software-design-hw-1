// Package cli предоставляет консольное меню учёта транспорта.
package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"escooter-sharing/internal/application"
	"escooter-sharing/internal/domain"
)

var errInputClosed = errors.New("ввод завершён")

var errInvalidInput = errors.New("некорректный ввод")

const menu = "\nВышКат\n1. Добавить транспорт без электропривода\n2. Активный парк\n3. Отчёт о парке\n4. Добавить электротранспорт\n5. Техника для новичков\n6. Добавить шлем\n7. Добавить зарядную станцию\n8. Инвентарная ведомость\n0. Выход\n> "

type session struct {
	fleet   *application.Fleet
	scanner *bufio.Scanner
	output  io.Writer
}

// Run выполняет команды до выхода или конца ввода.
// Ошибки пользовательских данных выводятся в консоль, ошибки потоков возвращаются.
func Run(fleet *application.Fleet, input io.Reader, output io.Writer) error {
	s := session{fleet: fleet, scanner: bufio.NewScanner(input), output: output}
	err := s.run()
	if errors.Is(err, errInputClosed) {
		return nil
	}
	return err
}

func (s *session) run() error {
	for {
		command, err := s.readLine(menu)
		if err != nil {
			return err
		}
		switch command {
		case "0":
			return nil
		case "1":
			err = s.admit(false)
		case "2":
			err = s.showActive()
		case "3":
			err = s.showReport()
		case "4":
			err = s.admit(true)
		case "5":
			err = s.showBeginners()
		case "6":
			err = s.addItem(false)
		case "7":
			err = s.addItem(true)
		case "8":
			err = s.showInventory()
		default:
			err = s.printf("Неизвестная команда. Выберите действие от 0 до 8.\n")
		}
		if err != nil {
			return err
		}
	}
}

func (s *session) admit(electric bool) error {
	transport, err := s.readTransport(electric)
	if errors.Is(err, domain.ErrInvalidTransport) || errors.Is(err, errInvalidInput) {
		return s.printf("Ошибка: %v. Повторите добавление.\n", err)
	}
	if err != nil {
		return err
	}
	admission, err := s.fleet.Admit(transport)
	if err != nil {
		return s.printf("Ошибка: %v. Повторите добавление.\n", err)
	}
	if admission.Approved() {
		return s.printf("Транспорт %q (%s) допущен в активный парк.\n", transport.Name(), transport.InventoryNumber())
	}
	return s.printf("Транспорт %q (%s) отклонён техосмотром; оставлен на балансе.\n", transport.Name(), transport.InventoryNumber())
}

func (s *session) readTransport(electric bool) (domain.Vehicle, error) {
	number, name, err := s.readIdentity()
	if err != nil {
		return nil, err
	}
	condition, err := s.readLine("Технически исправен (да/нет): ")
	if err != nil {
		return nil, err
	}
	condition = strings.ToLower(condition)
	if condition != "да" && condition != "нет" {
		return nil, fmt.Errorf("%w: для технического состояния введите «да» или «нет»", errInvalidInput)
	}
	ease, err := s.readLine("Простота для новичка (1–10): ")
	if err != nil {
		return nil, err
	}
	beginnerEase, err := strconv.Atoi(ease)
	if err != nil {
		return nil, fmt.Errorf("%w: укажите простоту целым числом от 1 до 10", errInvalidInput)
	}
	transport, err := domain.NewTransport(number, name, condition == "да", beginnerEase)
	if err != nil {
		return nil, err
	}
	if !electric {
		return transport, nil
	}
	dailyEnergyKWh, err := s.readEnergy()
	if err != nil {
		return nil, err
	}
	return domain.NewElectricTransport(number, name, condition == "да", beginnerEase, dailyEnergyKWh)
}

func (s *session) showActive() error {
	transports, err := s.fleet.ActiveTransports()
	if err != nil {
		return s.printf("Ошибка: %v. Повторите запрос.\n", err)
	}
	if len(transports) == 0 {
		return s.printf("Активный парк пуст.\n")
	}
	if err := s.printf("Допущенный транспорт:\n"); err != nil {
		return err
	}
	for _, transport := range transports {
		if err := s.printf("%s — %s\n", transport.InventoryNumber(), transport.Name()); err != nil {
			return err
		}
	}
	return nil
}

func (s *session) readLine(prompt string) (string, error) {
	if err := s.printf("%s", prompt); err != nil {
		return "", err
	}
	if !s.scanner.Scan() {
		if err := s.scanner.Err(); err != nil {
			return "", fmt.Errorf("чтение ввода: %w", err)
		}
		return "", errInputClosed
	}
	return strings.TrimSpace(s.scanner.Text()), nil
}

func (s *session) printf(format string, args ...any) error {
	if _, err := fmt.Fprintf(s.output, format, args...); err != nil {
		return fmt.Errorf("вывод в консоль: %w", err)
	}
	return nil
}
