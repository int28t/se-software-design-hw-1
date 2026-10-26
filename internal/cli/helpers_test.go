package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"escooter-sharing/internal/application"
	"escooter-sharing/internal/cli"
	"escooter-sharing/internal/domain"
	"escooter-sharing/internal/infrastructure/memory"
)

func runSession(t *testing.T, input string) (*application.Fleet, string) {
	t.Helper()
	fleet := application.NewFleet(domain.ServiceCenter{}, &memory.FleetRepository{})
	var output bytes.Buffer
	if err := cli.Run(fleet, strings.NewReader(input), &output); err != nil {
		t.Fatal(err)
	}
	return fleet, output.String()
}
