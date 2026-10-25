package cli

func (s *session) showReport() error {
	report, err := s.fleet.Report()
	if err != nil {
		return s.printf("Ошибка: %v. Повторите запрос.\n", err)
	}
	return s.printf(
		"Транспорта на балансе: %d\n"+
			"Допущено в активный парк: %d\n"+
			"Отклонено техосмотром: %d\n"+
			"Суточная энергия всего баланса: %g кВт·ч/сутки\n"+
			"Суточная энергия активного транспорта: %g кВт·ч/сутки\n",
		report.TotalTransports, report.ActiveTransports, report.RejectedTransports,
		report.TotalDailyEnergyKWh, report.ActiveDailyEnergyKWh,
	)
}
