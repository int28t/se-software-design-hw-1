package domain_test

import (
	"errors"
	"testing"

	"escooter-sharing/internal/domain"
)

func TestNewTransport(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		number string
		title  string
		valid  bool
	}{
		{name: "корректные данные", number: "S-1", title: "Самокат", valid: true},
		{name: "пробелы по краям", number: " S-1 ", title: " Самокат ", valid: true},
		{name: "пустой номер", title: "Самокат"},
		{name: "номер из пробелов", number: " \t", title: "Самокат"},
		{name: "пустое название", number: "S-1"},
		{name: "название из пробелов", number: "S-1", title: " \t"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			transport, err := domain.NewTransport(tt.number, tt.title, true)
			if !tt.valid {
				if !errors.Is(err, domain.ErrInvalidTransport) {
					t.Fatalf("NewTransport() error = %v, want ErrInvalidTransport", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if transport.InventoryNumber() != "S-1" || transport.Name() != "Самокат" || !transport.IsServiceable() {
				t.Fatalf("NewTransport() = %+v", transport)
			}
		})
	}
}

func TestServiceCenterInspect(t *testing.T) {
	t.Parallel()
	for _, serviceable := range []bool{true, false} {
		t.Run(map[bool]string{true: "принят", false: "отклонён"}[serviceable], func(t *testing.T) {
			t.Parallel()
			transport, err := domain.NewTransport("S-1", "Самокат", serviceable)
			if err != nil {
				t.Fatal(err)
			}
			approved, err := (domain.ServiceCenter{}).Inspect(transport)
			if err != nil {
				t.Fatal(err)
			}
			if approved != serviceable {
				t.Fatalf("Inspect() = %v, want %v", approved, serviceable)
			}
		})
	}
}

func TestServiceCenterRejectsZeroTransport(t *testing.T) {
	t.Parallel()
	approved, err := (domain.ServiceCenter{}).Inspect(domain.Transport{})
	if approved || !errors.Is(err, domain.ErrInvalidTransport) {
		t.Fatalf("Inspect() = %v, %v; want false, ErrInvalidTransport", approved, err)
	}
}
