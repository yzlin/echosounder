package version

import (
	"fmt"
	"runtime"
)

// Binary means the version of the program binary.
const Binary = "0.9.0"

// String returns binary version string of the app.
func String(app string) string {
	return fmt.Sprintf("%s v%s (built w/%s)", app, Binary, runtime.Version())
}
