package domain_test

import (
	"errors"
	"testing"

	"escooter-sharing/internal/domain"
)

func TestBeginnerEaseBoundsAndRule(t *testing.T) {
	t.Parallel()
	factories := []struct {
		name string
		new  func(int) (domain.Vehicle, error)
	}{
		{name: "обычный транспорт", new: func(ease int) (domain.Vehicle, error) {
			return domain.NewTransport("B-1", "Велосипед", true, ease)
		}},
		{name: "электротранспорт", new: func(ease int) (domain.Vehicle, error) {
			return domain.NewElectricTransport("E-1", "Электросамокат", true, ease, 1.25)
		}},
		{name: "велосипед", new: func(ease int) (domain.Vehicle, error) {
			return domain.NewBicycle("B-1", "Велосипед", true, ease)
		}},
		{name: "самокат", new: func(ease int) (domain.Vehicle, error) {
			return domain.NewScooter("S-1", "Самокат", true, ease, 1.25)
		}},
		{name: "электровелосипед", new: func(ease int) (domain.Vehicle, error) {
			return domain.NewElectricBicycle("E-1", "Электровелосипед", true, ease, 1.25)
		}},
	}
	tests := []struct {
		name     string
		ease     int
		valid    bool
		beginner bool
	}{
		{name: "отрицательная простота", ease: -1},
		{name: "ниже минимума", ease: 0},
		{name: "минимум", ease: 1, valid: true},
		{name: "перед порогом", ease: 5, valid: true},
		{name: "порог", ease: 6, valid: true, beginner: true},
		{name: "максимум", ease: 10, valid: true, beginner: true},
		{name: "выше максимума", ease: 11},
	}
	for _, factory := range factories {
		t.Run(factory.name, func(t *testing.T) {
			t.Parallel()
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					t.Parallel()
					transport, err := factory.new(tt.ease)
					if !tt.valid {
						if !errors.Is(err, domain.ErrInvalidTransport) {
							t.Fatalf("создание: %v, want ErrInvalidTransport", err)
						}
						return
					}
					if err != nil {
						t.Fatal(err)
					}
					if transport.BeginnerEase() != tt.ease || transport.SuitableForBeginner() != tt.beginner {
						t.Fatalf("простота = %d, пригодность = %v; want %d, %v", transport.BeginnerEase(), transport.SuitableForBeginner(), tt.ease, tt.beginner)
					}
				})
			}
		})
	}
}

func TestBeginnerRuleDoesNotReplaceInspection(t *testing.T) {
	t.Parallel()
	transport, err := domain.NewTransport("B-1", "Велосипед", false, 10)
	if err != nil {
		t.Fatal(err)
	}
	if !transport.SuitableForBeginner() {
		t.Fatal("правило простоты не должно зависеть от технического состояния")
	}
	approved, err := (domain.ServiceCenter{}).Inspect(transport)
	if err != nil || approved {
		t.Fatalf("Inspect() = %v, %v; want false, nil", approved, err)
	}
}
