package logger

import (
	"bytes"
	"fmt"
	"log/slog"
	"math/big"
	"reflect"
	"strconv"
	"time"

	"github.com/holiman/uint256"
)

const (
	timeFormat        = "2006-01-02T15:04:05-0700"
	floatFormat       = 'f'
	termMsgJust       = 40
	termCtxMaxPadding = 40
)

// 40 spaces
var spaces = []byte("                                        ")

type TerminalStringer interface {
	TerminalString() string
}

// LevelAlignedString returns a 5-character string containing the name of a Lvl.
func LevelAlignedString(l slog.Level) string {
	switch l {
	case LevelTrace:
		return "TRACE"
	case slog.LevelDebug:
		return "DEBUG"
	case slog.LevelInfo:
		return "INFO "
	case slog.LevelWarn:
		return "WARN "
	case slog.LevelError:
		return "ERROR"
	case LevelCrit:
		return "CRIT "
	default:
		return "unknown level"
	}
}

// LevelString returns a string containing the name of a Lvl.
func LevelString(l slog.Level) string {
	switch l {
	case LevelTrace:
		return "trace"
	case slog.LevelDebug:
		return "debug"
	case slog.LevelInfo:
		return "info"
	case slog.LevelWarn:
		return "warn"
	case slog.LevelError:
		return "error"
	case LevelCrit:
		return "crit"
	default:
		return "unknown"
	}
}

func escapeMessage(s string) string {
	needsQuoting := false
	for _, r := range s {
		// Allow CR/LF/TAB. This is to make multi-line messages work.
		if r == '\r' || r == '\n' || r == '\t' {
			continue
		}
		// We quote everything below <space> (0x20) and above~ (0x7E),
		// plus equal-sign
		if r < ' ' || r > '~' || r == '=' {
			needsQuoting = true
			break
		}
	}
	if !needsQuoting {
		return s
	}
	return strconv.Quote(s)
}

// appendEscapeString writes the string s to the given writer, with
// escaping/quoting if needed.
func appendEscapeString(dst []byte, s string) []byte {
	needsQuoting := false
	needsEscaping := false
	for _, r := range s {
		// If it contains spaces or equal-sign, we need to quote it.
		if r == ' ' || r == '=' {
			needsQuoting = true
			continue
		}
		// We need to escape it, if it contains
		// - character " (0x22) and lower (except space)
		// - characters above ~ (0x7E), plus equal-sign
		if r <= '"' || r > '~' {
			needsEscaping = true
			break
		}
	}
	if needsEscaping {
		return strconv.AppendQuote(dst, s)
	}
	// No escaping needed, but we might have to place within quote-marks, in case
	// it contained a space
	if needsQuoting {
		dst = append(dst, '"')
		dst = append(dst, []byte(s)...)
		return append(dst, '"')
	}
	return append(dst, []byte(s)...)
}

// writeTimeTermFormat writes on the format "01-02|15:04:05.000"
func writeTimeTermFormat(buf *bytes.Buffer, t time.Time) {
	_, month, day := t.Date()
	writePosIntWidth(buf, int(month), 2)
	buf.WriteByte('-')
	writePosIntWidth(buf, day, 2)
	buf.WriteByte('|')
	hour, min, sec := t.Clock()
	writePosIntWidth(buf, hour, 2)
	buf.WriteByte(':')
	writePosIntWidth(buf, min, 2)
	buf.WriteByte(':')
	writePosIntWidth(buf, sec, 2)
	ns := t.Nanosecond()
	buf.WriteByte('.')
	writePosIntWidth(buf, ns/1e6, 3)
}

// writePosIntWidth writes non-negative integer i to the buffer, padded on the left
// by zeroes to the given width. Use a width of 0 to omit padding.
// Adapted from pkg.go.dev/log/slog/internal/buffer
func writePosIntWidth(b *bytes.Buffer, i, width int) {
	// Cheap integer to fixed-width decimal ASCII.
	// Copied from log/log.go.
	if i < 0 {
		panic("negative int")
	}
	// Assemble decimal in reverse order.
	var bb [20]byte
	bp := len(bb) - 1
	for i >= 10 || width > 1 {
		width--
		q := i / 10
		bb[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	bb[bp] = byte('0' + i)
	b.Write(bb[bp:])
}

// FormatSlogValue formats a slog.Value for serialization to terminal.
func FormatSlogValue(v slog.Value, tmp []byte) (result []byte) {
	var value any
	defer func() {
		if err := recover(); err != nil {
			if v := reflect.ValueOf(value); v.Kind() == reflect.Ptr && v.IsNil() {
				result = []byte("<nil>")
			} else {
				panic(err)
			}
		}
	}()

	switch v.Kind() {
	case slog.KindString:
		return appendEscapeString(tmp, v.String())
	case slog.KindInt64: // All int-types (int8, int16 etc) wind up here
		return appendInt64(tmp, v.Int64())
	case slog.KindUint64: // All uint-types (uint8, uint16 etc) wind up here
		return appendUint64(tmp, v.Uint64(), false)
	case slog.KindFloat64:
		return strconv.AppendFloat(tmp, v.Float64(), floatFormat, 3, 64)
	case slog.KindBool:
		return strconv.AppendBool(tmp, v.Bool())
	case slog.KindDuration:
		value = v.Duration()
	case slog.KindTime:
		// Performance optimization: No need for escaping since the provided
		// timeFormat doesn't have any escape characters, and escaping is
		// expensive.
		return v.Time().AppendFormat(tmp, timeFormat)
	default:
		value = v.Any()
	}
	if value == nil {
		return []byte("<nil>")
	}
	switch v := value.(type) {
	case *big.Int: // Need to be before fmt.Stringer-clause
		return appendBigInt(tmp, v)
	case *uint256.Int: // Need to be before fmt.Stringer-clause
		return appendU256(tmp, v)
	case error:
		return appendEscapeString(tmp, v.Error())
	case TerminalStringer:
		return appendEscapeString(tmp, v.TerminalString())
	case fmt.Stringer:
		return appendEscapeString(tmp, v.String())
	}

	// We can use the 'tmp' as a scratch-buffer, to first format the
	// value, and in a second step do escaping.
	internal := fmt.Appendf(tmp, "%+v", value)
	return appendEscapeString(tmp, string(internal))
}

// appendInt64 formats n with thousand separators and writes into buffer dst.
func appendInt64(dst []byte, n int64) []byte {
	if n < 0 {
		return appendUint64(dst, uint64(-n), true)
	}
	return appendUint64(dst, uint64(n), false)
}

// appendUint64 formats n with thousand separators and writes into buffer dst.
func appendUint64(dst []byte, n uint64, neg bool) []byte {
	// Small numbers are fine as is
	if n < 100000 {
		if neg {
			return strconv.AppendInt(dst, -int64(n), 10)
		} else {
			return strconv.AppendInt(dst, int64(n), 10)
		}
	}
	// Large numbers should be split
	const maxLength = 26

	var (
		out   = make([]byte, maxLength)
		i     = maxLength - 1
		comma = 0
	)
	for ; n > 0; i-- {
		if comma == 3 {
			comma = 0
			out[i] = ','
		} else {
			comma++
			out[i] = '0' + byte(n%10)
			n /= 10
		}
	}
	if neg {
		out[i] = '-'
		i--
	}
	return append(dst, out[i+1:]...)
}

// FormatLogfmtUint64 formats n with thousand separators.
func FormatLogfmtUint64(n uint64) string {
	return string(appendUint64(nil, n, false))
}

// appendBigInt formats n with thousand separators and writes to dst.
func appendBigInt(dst []byte, n *big.Int) []byte {
	if n.IsUint64() {
		return appendUint64(dst, n.Uint64(), false)
	}
	if n.IsInt64() {
		return appendInt64(dst, n.Int64())
	}

	var (
		text  = n.String()
		buf   = make([]byte, len(text)+len(text)/3)
		comma = 0
		i     = len(buf) - 1
	)
	for j := len(text) - 1; j >= 0; j, i = j-1, i-1 {
		c := text[j]

		switch {
		case c == '-':
			buf[i] = c
		case comma == 3:
			buf[i] = ','
			i--
			comma = 0
			fallthrough
		default:
			buf[i] = c
			comma++
		}
	}
	return append(dst, buf[i+1:]...)
}

// appendU256 formats n with thousand separators.
func appendU256(dst []byte, n *uint256.Int) []byte {
	if n.IsUint64() {
		return appendUint64(dst, n.Uint64(), false)
	}
	res := []byte(n.PrettyDec(','))
	return append(dst, res...)
}
