package cli

import "os/exec"

// execCommand wraps exec.Command so it can be replaced in tests.
var execCommand = exec.Command
