package domain_test

import (
	"errors"
	"testing"

	"escooter-sharing/internal/domain"
)

func TestEmployeeIsNotInventoried(t *testing.T) {
	t.Parallel()
	employee, err := domain.NewEmployee(" Алексей ")
	if err != nil {
		t.Fatal(err)
	}
	if employee.Name() != "Алексей" {
		t.Fatalf("имя = %q, want Алексей", employee.Name())
	}
	if _, ok := any(employee).(domain.InventoryItem); ok {
		t.Fatal("сотрудник не должен быть инвентаризируемым объектом")
	}
}

func TestEmployeeRequiresName(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"", " \t"} {
		if _, err := domain.NewEmployee(name); !errors.Is(err, domain.ErrInvalidEmployee) {
			t.Errorf("NewEmployee(%q) error = %v, want ErrInvalidEmployee", name, err)
		}
	}
}
