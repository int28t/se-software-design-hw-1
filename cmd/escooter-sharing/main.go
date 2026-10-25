// Приложение escooter-sharing запускает консольный учёт транспорта «ВышКат».
package main

import (
	"fmt"
	"os"

	"escooter-sharing/internal/bootstrap"
)

func main() {
	if err := bootstrap.Run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка:", err)
		os.Exit(1)
	}
}
