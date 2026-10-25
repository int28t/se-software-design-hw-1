package domain

// EnergyConsumer описывает объект с заданным суточным расходом электроэнергии.
type EnergyConsumer interface {
	DailyEnergyKWh() float64
}
