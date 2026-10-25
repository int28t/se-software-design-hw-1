// Приложение escooter-sharing запускает консольный учёт транспорта «ВышКат».
package main

import (
	"fmt"
	"os"

	"escooter-sharing/internal/application"
	"escooter-sharing/internal/cli"
	"escooter-sharing/internal/domain"
	"escooter-sharing/internal/infrastructure/memory"
)

func main() {
	repository := &memory.FleetRepository{}
	fleet := application.NewFleet(domain.ServiceCenter{}, repository)
	if err := cli.Run(fleet, os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка:", err)
		os.Exit(1)
	}
}
