package didfmt

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"unicode"

	"git.tcp.direct/kayos/common/pool"
)

/*
NumberFormatter is an optimized DID formatter that formats phone numbers.
*/
type NumberFormatter struct {
	src    io.Reader
	err    error
	buf    *pool.Buffer
	closed bool
	*sync.Mutex
}

var (
	BufferPool               = pool.NewBufferFactory()
	PhoneNumberPool          = &sync.Pool{New: func() interface{} { return make([]byte, 11) }}
	FormattedPhoneNumberPool = &sync.Pool{New: func() interface{} { return make([]byte, 17) }}
)

// NewNumberFormatter returns a new NumberFormatter.
func NewNumberFormatter(source io.Reader) *NumberFormatter {
	nf := &NumberFormatter{Mutex: &sync.Mutex{}}
	nf.buf = BufferPool.Get()
	nf.src = source
	return nf
}

// Reset resets the NumberFormatter to use the provided io.Reader.
// Reset clears any errors, as well as any existing buffered data.
func (nf *NumberFormatter) Reset(source io.Reader) {
	nf.Lock()
	nf.src = source
	nf.err = nil
	nf.buf.MustReset()
	nf.Unlock()
}

// Close closes the NumberFormatter and (if we can cast it to an [io.Closer]) the underlying source.
func (nf *NumberFormatter) Close() error {
	nf.Lock()
	BufferPool.MustPut(nf.buf)
	nf.buf = nil
	nf.err = io.EOF
	nf.src = nil
	nf.closed = true
	if c, ok := nf.src.(io.Closer); ok {
		if err := c.Close(); err != nil {
			nf.Unlock()
			return err
		}
	}
	nf.Unlock()
	return nil
}

// Next returns the next formatted phone number or an empty string if there are no more numbers to format.
func (nf *NumberFormatter) Next() string { //nolint:funlen
	nf.Lock()
	if nf.closed {
		nf.Unlock()
		return ""
	}

	if nf.err != nil {
		nf.Unlock()
		return ""
	}

	inBuf := PhoneNumberPool.Get().([]byte)
	inBuf = inBuf[:0]
	inBuf = inBuf[:11]

	src := nf.src
	if src == nil {
		src = nf.buf
	}

	n, err := io.ReadFull(src, inBuf)
	switch {
	case err != nil && !errors.Is(err, io.EOF):
		nf.err = err
		fallthrough
	case errors.Is(err, io.EOF):
		PhoneNumberPool.Put(inBuf)
		nf.Unlock()
		return ""
	case n != 11:
		nf.err = fmt.Errorf("got length: %d (%w)", n, io.ErrShortBuffer)
		PhoneNumberPool.Put(inBuf)
		nf.Unlock()
		return ""
	default:
	}

	outBuf := FormattedPhoneNumberPool.Get().([]byte)
	outBuf = outBuf[:0]
	outBuf = outBuf[:17]

	// iterate over the input buffer of size 11 bytes and populate output buffer of size 17 bytes
	outBuf[0] = '+' // first byte is always a plus sign
	for i := range inBuf {
		if inBuf[i] < '0' || inBuf[i] > '9' {
			/*nf.err = fmt.Errorf("invalid character: %s", string(inBuf[i]))
			FormattedPhoneNumberPool.Put(outBuf)
			PhoneNumberPool.Put(inBuf)
			nf.Unlock()
			return ""*/
			panic("how'd these invalid characters get in here?")
		}
		switch i {
		case 0:
			outBuf[1] = inBuf[0]
		case 1:
			outBuf[2] = ' '
			outBuf[3] = '(' // open parenthesis
			outBuf[4] = inBuf[1]
		case 2: //nolint:gomnd
			outBuf[5] = inBuf[2]
		case 3: //nolint:gomnd
			outBuf[6] = inBuf[3]
			outBuf[7] = ')' // close parenthesis
			outBuf[8] = ' ' // space
		case 4, 5: //nolint:gomnd
			outBuf[i+5] = inBuf[i]
		case 6: //nolint:gomnd
			outBuf[11] = inBuf[6]
			outBuf[12] = '-' // hyphen
		case 7, 8, 9, 10: //nolint:gomnd
			outBuf[i+6] = inBuf[i]
		}
	}
	out := string(outBuf)
	FormattedPhoneNumberPool.Put(outBuf)
	PhoneNumberPool.Put(inBuf)
	nf.Unlock()
	return out

}

// Write implements io.Writer. It writes the provided bytes to the internal buffer.
func (nf *NumberFormatter) Write(p []byte) (n int, err error) {
	nf.Lock()
	ogLen := nf.buf.Len()

	switch {
	case len(p) == 0:
		nf.Unlock()
		return 0, fmt.Errorf("%w: len(p) == %d", io.ErrShortWrite, len(p))
		// ---------------------------
		//	case nf.err != nil:
		//		nf.Unlock()
		//		return 0, fmt.Errorf("error from previous write: %w" nf.err
		// ---------------------------
		// because nf.err was incurred during reading and isn't a write error,
		// the caller just needs to use the Err method, read errors shouldn't
		// be causing writes to fail
		// ---------------------------
	case nf.closed:
		nf.Unlock()
		return 0, io.ErrClosedPipe
	}

	for _, b := range p {
		switch {
		case !unicode.IsNumber(rune(b)), b < '0', b > '9':
			continue
		default:
			nf.buf.MustWriteByte(b)
		}
	}
	nb := nf.buf.Len()
	switch {
	case nb == 0:
		nf.Unlock()
		return 0, fmt.Errorf("%w: invalid buffer length: %d", io.ErrUnexpectedEOF, n)
	default:
		//
	}

	n = nf.buf.Len() - ogLen
	nf.Unlock()
	return n, nil
}

// Read implements io.Reader. It formats numbers from the data previously written, and then
// reads formatted phone numbers into the provided buffer.
func (nf *NumberFormatter) Read(p []byte) (n int, err error) {
	switch {
	case nf.err != nil:
		return 0, nf.err
	case cap(p) < 17:
		return 0, fmt.Errorf("invalid buffer size: %d", len(p))
	}
	for {
		if n >= cap(p) {
			break
		}
		num := nf.Next()
		if num == "" {
			err = io.EOF
			if nf.err != nil {
				err = nf.err
			}
			return n, err
		}
		n += copy(p[n:], num)
		if n >= 17 && n+1 < cap(p) {
			n += copy(p[n:], "\n")
		}
	}
	return n, nf.err
}

// Err returns the error, if any, that occurred during the last call to Next.
func (nf *NumberFormatter) Err() error {
	return nf.err
}

// FormatNumberString formats an input phone number string in the format of +1 (123) 456-7890.
func FormatNumberString(in string) (string, error) {
	nf := NewNumberFormatter(strings.NewReader(in))
	return nf.Next(), nf.Err()
}

// FormatNumberBytes formats an input phone number byte slice in the format of +1 (123) 456-7890.
func FormatNumberBytes(in []byte) (string, error) {
	return FormatNumberString(string(in))
}
