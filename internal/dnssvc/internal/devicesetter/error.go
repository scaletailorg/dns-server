package devicesetter

import (
	"fmt"

	"github.com/AdguardTeam/AdGuardDNS/internal/errcoll"
	"github.com/AdguardTeam/golibs/errors"
)

const (
	// ErrUnknownDedicated is returned by [Interface.SetDevice] if the request
	// should be dropped, because it's a request for an unknown dedicated IP
	// address.
	ErrUnknownDedicated errors.Error = "unknown dedicated ip"
)

// deviceDataError is an error about bad device data or other issues found
// during device data checking.
type deviceDataError struct {
	err error
	typ string
}

// type check
var _ error = (*deviceDataError)(nil)

// newDeviceDataError is a helper constructor for device-data errors.
func newDeviceDataError(orig error, typ string) (err error) {
	return &deviceDataError{
		err: orig,
		typ: typ,
	}
}

// Error implements the error interface for *deviceDataError.
func (err *deviceDataError) Error() (msg string) {
	return fmt.Sprintf("%s device id check: %s", err.typ, err.err)
}

// type check
var _ errors.Wrapper = (*deviceDataError)(nil)

// Unwrap implements the [errors.Wrapper] interface for *deviceDataError.
func (err *deviceDataError) Unwrap() (unwrapped error) { return err.err }

// type check
var _ errcoll.SentryReportableError = (*deviceDataError)(nil)

// IsSentryReportable implements the [errcoll.SentryReportableError] interface
// for *deviceDataError.
func (*deviceDataError) IsSentryReportable() (ok bool) { return false }
