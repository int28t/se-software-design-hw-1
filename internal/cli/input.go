package cli

import (
	"fmt"
	"strconv"
	"strings"
)

func (s *session) readIdentity() (string, string, error) {
	number, err := s.readLine("Инвентарный номер: ")
	if err != nil {
		return "", "", err
	}
	name, err := s.readLine("Название: ")
	if err != nil {
		return "", "", err
	}
	return number, name, nil
}

func (s *session) readCondition() (bool, error) {
	condition, err := s.readLine("Технически исправен (да/нет): ")
	if err != nil {
		return false, err
	}
	switch strings.ToLower(condition) {
	case "да":
		return true, nil
	case "нет":
		return false, nil
	default:
		return false, fmt.Errorf("%w: для технического состояния введите «да» или «нет»", errInvalidInput)
	}
}

func (s *session) readBeginnerEase() (int, error) {
	ease, err := s.readLine("Простота для новичка (1–10; новичкам от 6): ")
	if err != nil {
		return 0, err
	}
	beginnerEase, err := strconv.Atoi(ease)
	if err != nil {
		return 0, fmt.Errorf("%w: укажите простоту целым числом от 1 до 10", errInvalidInput)
	}
	return beginnerEase, nil
}

func (s *session) readEnergy() (float64, error) {
	consumption, err := s.readLine("Суточный расход энергии (кВт·ч, например 2.5 или 2,5): ")
	if err != nil {
		return 0, err
	}
	dailyEnergyKWh, err := strconv.ParseFloat(strings.ReplaceAll(consumption, ",", "."), 64)
	if err != nil {
		return 0, fmt.Errorf("%w: укажите расход энергии числом, например 2.5", errInvalidInput)
	}
	return dailyEnergyKWh, nil
}
