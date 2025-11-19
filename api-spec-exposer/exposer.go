package exposer

import (
	"fmt"

	"github.com/qubership-apihub-commons-go/api-spec-exposer/config"
	"github.com/qubership-apihub-commons-go/api-spec-exposer/internal/generator"
	"github.com/qubership-apihub-commons-go/api-spec-exposer/internal/scanner"
)

// SpecExposer is the main interface for API spec exposure
type SpecExposer interface {
	// Discover scans directory and discovers specs
	Discover() (config.DiscoveryResult, error)
}

// specExposer is the default implementation
type specExposer struct {
	config          config.DiscoveryConfig
	discoveryResult config.DiscoveryResult
}

// New creates a new SpecExposer instance
func New(config config.DiscoveryConfig) SpecExposer {
	return &specExposer{
		config: config,
	}
}

// Discover scans directory and discovers specs
func (se *specExposer) Discover() (config.DiscoveryResult, error) {
	specScanner := scanner.New(se.config)

	specs, warnings, err := specScanner.Scan()
	if err != nil {
		return config.DiscoveryResult{}, fmt.Errorf("scan failed: %w", err)
	}

	se.discoveryResult.Warnings = append(se.discoveryResult.Warnings, warnings...)
	se.discoveryResult.Specs = specs

	gen := generator.New(se.discoveryResult.Specs)
	endpoints, err := gen.Generate()
	if err != nil {
		return config.DiscoveryResult{}, fmt.Errorf("endpoint generation failed: %w", err)
	}

	se.discoveryResult.Endpoints = endpoints

	return se.discoveryResult, nil
}
