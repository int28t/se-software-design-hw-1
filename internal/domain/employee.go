package domain

import (
	"errors"
	"strings"
)

// ErrInvalidEmployee означает, что имя сотрудника не задано.
var ErrInvalidEmployee = errors.New("укажите имя сотрудника")

// Employee описывает сотрудника, который не является инвентаризируемым объектом.
type Employee struct {
	name string
}

// NewEmployee создаёт сотрудника без инвентарного номера.
func NewEmployee(name string) (Employee, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Employee{}, ErrInvalidEmployee
	}
	return Employee{name: name}, nil
}

// Name возвращает имя сотрудника.
func (e Employee) Name() string {
	return e.name
}
