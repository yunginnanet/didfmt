# didfmt

package `didfmt` implements `io.ReadWriteCloser` and allows you to write streams of unformatted 11 digit (USA) phone numbers to it, and read formatted results separated by newline.
It is unnecessarily fast, but only supports 11 digit USA numbers. This makes it nearly a toy, but I've also made use of it in code that made it to staging environments, hilariously enough.

This was something I made a while ago for a very specific purpose. Mostly fo fun, I decided to extract the package and do some interesting things with Go's stdlib `testing` package.

For example, the tests in this package will actually fail if using this package is slower than using `fmt.Sprintf` to do the same thing. i.e: this package, while doing the same thing as the following snippet, is faster and allocates less memory; guaranteed. ***otherwise the unit tests will literally fail :^)***

```golang
fmt.Sprintf("+%s (%s) %s-%s", string(inbuf[0]), inbuf[1:4], inbuf[5:8], inbuf[6:10])
```


I'll let the output of `go test -v` speak for itself:

```
=== RUN   TestNumberFormatter
=== RUN   TestNumberFormatter/single
=== RUN   TestNumberFormatter/multiple
=== RUN   TestNumberFormatter/error
=== RUN   TestNumberFormatter/FormatNumberString
=== RUN   TestNumberFormatter/FormatNumberBytes
=== RUN   TestNumberFormatter/reader
=== RUN   TestNumberFormatter/writer
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
        "1 2 8 1))) MIKE JONES 33 oh (0 rather) eight 8 zero 0 zero 0 fo' 4"" hell yeah
        
        1555 
        
        how's that one song go lmao its like 867 ... uh, 5309? i think?
        
        idfk
    numbers_test.go:207: output:
        +1 (281) 330-8004
        +1 (555) 867-5309
--- PASS: TestNumberFormatter (0.00s)
    --- PASS: TestNumberFormatter/single (0.00s)
    --- PASS: TestNumberFormatter/multiple (0.00s)
    --- PASS: TestNumberFormatter/error (0.00s)
    --- PASS: TestNumberFormatter/FormatNumberString (0.00s)
    --- PASS: TestNumberFormatter/FormatNumberBytes (0.00s)
    --- PASS: TestNumberFormatter/reader (0.00s)
    --- PASS: TestNumberFormatter/writer (0.00s)
        --- PASS: TestNumberFormatter/writer/ridiculous (0.00s)
=== RUN   TestBenchmark
    numbers_test.go:273: 
        
        --------------------------------------
        NumberFormatter.Next() benchmark 
        --------------------------------------
        312 nanoseconds per phone number
        allocated memory 3 times per phone number
        allocated bytes 72 bytes of memory per number
        it took 1.201518743s to format 3842572 phone numbers
        --------------------------------------
        
    numbers_test.go:273: 
        
        --------------------------------------
        NumberFormatter.Read() benchmark 
        --------------------------------------
        323 nanoseconds per phone number
        allocated memory 3 times per phone number
        allocated bytes 72 bytes of memory per number
        it took 1.186606168s to format 3665048 phone numbers
        --------------------------------------
        
    numbers_test.go:273: 
        
        --------------------------------------
        fmt.Sprintf() benchmark 
        --------------------------------------
        457 nanoseconds per phone number
        allocated memory 6 times per phone number
        allocated bytes 116 bytes of memory per number
        it took 1.200540423s to format 2622765 phone numbers
        --------------------------------------
        
    numbers_test.go:380: Read() is 29% faster than fmt.Sprintf(), allocates 50% less memory, and allocates 37% less bytes
    numbers_test.go:387: Next() is 31% faster than fmt.Sprintf(), allocates 50% less memory, and allocates 37% less bytes
    numbers_test.go:394: 
        
--- PASS: TestBenchmark (4.70s)
PASS
ok  	git.tcp.direct/kayos/didfmt	4.705s
```





