// Package cli предоставляет консольное меню учёта транспорта.
package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"escooter-sharing/internal/application"
)

var errInputClosed = errors.New("ввод завершён")

var errInvalidInput = errors.New("некорректный ввод")

const menu = "\nВышКат\n1. Добавить велосипед\n2. Активный парк\n3. Отчёт о парке\n4. Добавить электросамокат\n5. Техника для новичков\n6. Добавить шлем\n7. Добавить зарядную станцию\n8. Инвентарная ведомость\n9. Добавить электровелосипед\n0. Выход\n> "

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
			err = s.admit(bicycleKind)
		case "2":
			err = s.showActive()
		case "3":
			err = s.showReport()
		case "4":
			err = s.admit(scooterKind)
		case "5":
			err = s.showBeginners()
		case "6":
			err = s.addItem(false)
		case "7":
			err = s.addItem(true)
		case "8":
			err = s.showInventory()
		case "9":
			err = s.admit(electricBicycleKind)
		default:
			err = s.printf("Неизвестная команда. Выберите действие от 0 до 9.\n")
		}
		if err != nil {
			return err
		}
	}
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
