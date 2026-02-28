// Web Application Example - Best Practices
//
// This example demonstrates a production-ready directory structure
// with separation of concerns and dependency injection.
//
// Directory Structure:
//   cmd/           - Application entry points
//   internal/      - Private application code
//     config/      - Configuration management
//     model/       - Domain models
//     repository/  - Data access layer
//     service/     - Business logic layer
//     handler/     - HTTP handlers
//     middleware/  - Request processing pipeline
//     infrastructure/ - External service connections
//     app/         - Application orchestration
//     wire/        - Dependency injection setup
//   pkg/           - Public library code
//     interfaces/  - Interface definitions

package main

import (
	"fmt"

	"github.com/Just-maple/godi/examples/09-web-app/internal/wire"
)

func main() {
	fmt.Println("=== Web Application Example ===")
	fmt.Println("Best Practices: Separation of Concerns")
	fmt.Println()

	if err := wire.Run(); err != nil {
		panic(err)
	}

	fmt.Println()
	fmt.Println("=== Demo Complete ===")
	fmt.Println()
	fmt.Println("Architecture Layers:")
	fmt.Println("  Config → Infrastructure (DB/Cache)")
	fmt.Println("             ↓")
	fmt.Println("         Repository")
	fmt.Println("             ↓")
	fmt.Println("          Service")
	fmt.Println("             ↓")
	fmt.Println("    Handler + Router")
	fmt.Println("             ↓")
	fmt.Println("       Middleware")
	fmt.Println("             ↓")
	fmt.Println("           App")
}
