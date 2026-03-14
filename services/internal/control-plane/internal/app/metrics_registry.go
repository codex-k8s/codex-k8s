package app

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

func registerOrReplaceCollector(registerer prometheus.Registerer, collector prometheus.Collector) error {
	if registerer == nil {
		return fmt.Errorf("prometheus registerer is required")
	}
	if collector == nil {
		return fmt.Errorf("prometheus collector is required")
	}

	if err := registerer.Register(collector); err != nil {
		var alreadyRegisteredErr prometheus.AlreadyRegisteredError
		if !errors.As(err, &alreadyRegisteredErr) {
			return err
		}
		if !registerer.Unregister(alreadyRegisteredErr.ExistingCollector) {
			return fmt.Errorf("unregister existing collector: %w", err)
		}
		if err := registerer.Register(collector); err != nil {
			return fmt.Errorf("register collector after replacement: %w", err)
		}
	}

	return nil
}
