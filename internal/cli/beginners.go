package cli

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
