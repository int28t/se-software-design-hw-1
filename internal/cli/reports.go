package cli

func (s *session) showReport() error {
	report, err := s.fleet.Report()
	if err != nil {
		return s.printf("Ошибка: %v. Повторите запрос.\n", err)
	}
	return s.printf(
		"Всего объектов на балансе: %d\n"+
			"Вещей на балансе: %d\n"+
			"Транспорта на балансе: %d\n"+
			"Допущено в активный парк: %d\n"+
			"Отклонено техосмотром: %d\n"+
			"Суточная энергия всего баланса: %g кВт·ч/сутки\n"+
			"Суточная энергия активного транспорта: %g кВт·ч/сутки\n",
		report.TotalAssets, report.TotalItems, report.TotalTransports, report.ActiveTransports, report.RejectedTransports,
		report.TotalDailyEnergyKWh, report.ActiveDailyEnergyKWh,
	)
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

func (s *session) showInventory() error {
	entries, err := s.fleet.Inventory()
	if err != nil {
		return s.printf("Ошибка: %v. Повторите запрос.\n", err)
	}
	if len(entries) == 0 {
		return s.printf("На балансе пока нет объектов.\n")
	}
	if err := s.printf("Инвентарная ведомость:\n"); err != nil {
		return err
	}
	for _, entry := range entries {
		if err := s.printf("%s — %s\n", entry.InventoryNumber, entry.Name); err != nil {
			return err
		}
	}
	return nil
}

func (s *session) showBeginners() error {
	transports, err := s.fleet.BeginnerTransports()
	if err != nil {
		return s.printf("Ошибка: %v. Повторите запрос.\n", err)
	}
	if len(transports) == 0 {
		return s.printf("Нет доступной техники для новичков.\n")
	}
	if err := s.printf("Техника для новичков:\n"); err != nil {
		return err
	}
	for _, transport := range transports {
		if err := s.printf("%s — %s (простота: %d/10)\n", transport.InventoryNumber(), transport.Name(), transport.BeginnerEase()); err != nil {
			return err
		}
	}
	return nil
}
