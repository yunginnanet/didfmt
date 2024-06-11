# didfmt

`import "git.tcp.direct/kayos/didfmt"`

## Overview

package `didfmt` implements `io.ReadWriteCloser` and allows you to write streams of unformatted 11 digit (USA) phone numbers to it, and read formatted results separated by newline.

It is unnecessarily fast, but only supports 11 digit USA numbers. This makes it nearly a toy, but I've also made use of it in code that made it to staging environments, hilariously enough.

## ...Why?

This was something I made a while ago for a very specific purpose, and was a small package in a large module. 
Mostly for fun, I decided to extract the package and do some interesting things with Go's stdlib `testing` package.

## Faster, or your green checkmark back.

This package, while doing the same thing as the following snippet, is:
  - faster
  - allocates less memory as a whole
  - has less memory allocation operations

```golang
fmt.Sprintf("+%s (%s) %s-%s", string(inbuf[0]), inbuf[1:4], inbuf[5:8], inbuf[6:10])
```

#### guaranteed. ***otherwise the unit tests will literally fail :^)***

## Yes, I'm serious.

I'll let the output of `go test -v` speak for itself:

```
=== RUN   TestNumberFormatter
[...]
    numbers_test.go:137: input:
        12813308004
    numbers_test.go:151: output:
        +1 (281) 330-8004
    numbers_test.go:165: input:
        1281330800412813308004
    numbers_test.go:177: output:
        +1 (281) 330-8004
        +1 (281) 330-8004
=== RUN   TestNumberFormatter/writer/ridiculous
    numbers_test.go:194: input:
        "1 2 8 1))) MIKE JONES 33 oh (0 rather) eight 8 zero 0 zero 0 fo' 4"" :^)
        
        1555 
        
        how's that one song go lmao its like 867 ... uh, 5309? i think?
        
        idfk
    numbers_test.go:207: output:
        +1 (281) 330-8004
        +1 (555) 867-5309
--- PASS: TestNumberFormatter (0.00s)
[...]
=== RUN   TestBenchmark
    numbers_test.go:273: 
        
        --------------------------------------
        NumberFormatter.Next() benchmark 
        --------------------------------------
        took 312 nanoseconds per phone number
        allocated memory 3 times per phone number
        allocated 72 bytes of memory per number
        it took 1.198201239s to format 3836050 phone numbers
        --------------------------------------
        
    numbers_test.go:273: 
        
        --------------------------------------
        NumberFormatter.Read() benchmark 
        --------------------------------------
        took 324 nanoseconds per phone number
        allocated memory 3 times per phone number
        allocated 72 bytes of memory per number
        it took 1.202981817s to format 3704997 phone numbers
        --------------------------------------
        
    numbers_test.go:273: 
        
        --------------------------------------
        fmt.Sprintf() benchmark 
        --------------------------------------
        took 465 nanoseconds per phone number
        allocated memory 6 times per phone number
        allocated 116 bytes of memory per number
        it took 1.212666898s to format 2607760 phone numbers
        --------------------------------------
        
    numbers_test.go:380: 
        
        Read() is 30% faster than fmt.Sprintf(),
        makes 50% less allocations to allocate 37% less bytes.
        
    numbers_test.go:387: 
        
        Next() is 32% faster than fmt.Sprintf(),
        makes 50% less allocations to allocate 37% less bytes.
        
--- PASS: TestBenchmark (4.73s)
PASS
ok  	git.tcp.direct/kayos/didfmt	4.730s
```

## Documentation

#### func  FormatNumberBytes

```go
func FormatNumberBytes(in []byte) (string, error)
```
FormatNumberBytes formats an input phone number byte slice in the format of +1
(123) 456-7890.

#### func  FormatNumberString

```go
func FormatNumberString(in string) (string, error)
```
FormatNumberString formats an input phone number string in the format of +1
(123) 456-7890.

#### type NumberFormatter

```go
type NumberFormatter struct {
	*sync.Mutex
}
```

NumberFormatter is an optimized DID formatter that formats phone numbers.

#### func  NewNumberFormatter

```go
func NewNumberFormatter(source io.Reader) *NumberFormatter
```
NewNumberFormatter returns a new NumberFormatter.

#### func (*NumberFormatter) Close

```go
func (nf *NumberFormatter) Close() error
```
Close closes the NumberFormatter and (if we can cast it to an [io.Closer]) the
underlying source.

#### func (*NumberFormatter) Err

```go
func (nf *NumberFormatter) Err() error
```
Err returns the error, if any, that occurred during the last call to Next.

#### func (*NumberFormatter) Next

```go
func (nf *NumberFormatter) Next() string
```
Next returns the next formatted phone number or an empty string if there are no
more numbers to format.

#### func (*NumberFormatter) Read

```go
func (nf *NumberFormatter) Read(p []byte) (n int, err error)
```
Read implements io.Reader. It formats numbers from the data previously written,
and then reads formatted phone numbers into the provided buffer.

#### func (*NumberFormatter) Reset

```go
func (nf *NumberFormatter) Reset(source io.Reader)
```
Reset resets the NumberFormatter to use the provided io.Reader. Reset clears any
errors, as well as any existing buffered data.

#### func (*NumberFormatter) Write

```go
func (nf *NumberFormatter) Write(p []byte) (n int, err error)
```
Write implements io.Writer. It writes the provided bytes to the internal buffer.

---
