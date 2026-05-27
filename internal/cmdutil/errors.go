package cmdutil

import (
	"errors"
	"fmt"
	"os"

	declaw "github.com/declaw-ai/declaw-go"
)

func HandleError(err error) {
	if err == nil {
		return
	}

	var authErr *declaw.AuthenticationError
	if errors.As(err, &authErr) {
		fmt.Fprintln(os.Stderr, "Error: Authentication failed. Run `declaw auth login` to set your API key.")
		os.Exit(4)
	}

	var notFound *declaw.NotFoundError
	if errors.As(err, &notFound) {
		fmt.Fprintf(os.Stderr, "Error: %s\n", notFound.Error())
		os.Exit(1)
	}

	var balance *declaw.InsufficientBalanceError
	if errors.As(err, &balance) {
		fmt.Fprintln(os.Stderr, "Error: Insufficient balance. Visit https://console.declaw.ai to add funds.")
		os.Exit(1)
	}

	var rateLimit *declaw.RateLimitError
	if errors.As(err, &rateLimit) {
		fmt.Fprintf(os.Stderr, "Error: Rate limited. Retry after %ds.\n", int(rateLimit.RetryAfter.Seconds()))
		os.Exit(1)
	}

	var cmdExit *declaw.CommandExitError
	if errors.As(err, &cmdExit) {
		os.Exit(cmdExit.ExitCode)
	}

	var invalidArg *declaw.InvalidArgumentError
	if errors.As(err, &invalidArg) {
		fmt.Fprintf(os.Stderr, "Error: Invalid argument: %s\n", invalidArg.Error())
		os.Exit(2)
	}

	var timeout *declaw.TimeoutError
	if errors.As(err, &timeout) {
		fmt.Fprintln(os.Stderr, "Error: Request timed out.")
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	os.Exit(1)
}
