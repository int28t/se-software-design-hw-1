// Package cli предоставляет консольное меню учёта транспорта.
package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"escooter-sharing/internal/application"
	"escooter-sharing/internal/domain"
)

var errInputClosed = errors.New("ввод завершён")

const menu = "\nВышКат\n1. Добавить транспорт\n2. Активный парк\n0. Выход\n> "

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
			err = s.admit()
		case "2":
			err = s.showActive()
		default:
			err = s.printf("Неизвестная команда. Выберите 0, 1 или 2.\n")
		}
		if err != nil {
			return err
		}
	}
}

func (s *session) admit() error {
	number, err := s.readLine("Инвентарный номер: ")
	if err != nil {
		return err
	}
	name, err := s.readLine("Название: ")
	if err != nil {
		return err
	}
	condition, err := s.readLine("Технически исправен (да/нет): ")
	if err != nil {
		return err
	}
	condition = strings.ToLower(condition)
	if condition != "да" && condition != "нет" {
		return s.printf("Ошибка: для технического состояния введите «да» или «нет».\n")
	}
	transport, err := domain.NewTransport(number, name, condition == "да")
	if err != nil {
		return s.printf("Ошибка: %v. Повторите добавление.\n", err)
	}
	admission, err := s.fleet.Admit(transport)
	if err != nil {
		return s.printf("Ошибка: %v. Повторите добавление.\n", err)
	}
	if admission.Approved() {
		return s.printf("Транспорт %q (%s) допущен в активный парк.\n", transport.Name(), number)
	}
	return s.printf("Транспорт %q (%s) отклонён техосмотром; оставлен на балансе.\n", transport.Name(), number)
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
