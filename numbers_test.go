package didfmt

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"testing"

	"git.tcp.direct/kayos/common/entropy"
)

const (
	mikejones  = "12813308004"
	mikejonesf = "+1 (281) 330-8004"
)

var (
	numbers      = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	numbersBytes []byte
)

func init() {
	for _, n := range numbers {
		numbersBytes = append(numbersBytes, byte(n)+'0')
	}
	if string(numbersBytes) != "0123456789" {
		panic("bad numbers: " + string(numbersBytes))
	}
}

func TestNumberFormatter(t *testing.T) { //nolint:funlen //nolint:gocognit
	t.Run("single", func(t *testing.T) {
		nf := NewNumberFormatter(strings.NewReader(mikejones))
		res := nf.Next()
		if nf.Err() != nil {
			t.Fatal(nf.Err())
		}
		if res != mikejonesf {
			t.Fatalf("got: %s", res)
		}
	})
	t.Run("multiple", func(t *testing.T) {
		nums := &bytes.Buffer{}
		for n := 0; n != 11000; n++ {
			nums.WriteByte(numbersBytes[entropy.GetSharedRand().Int()%len(numbers)])
		}
		if len(nums.String()) != 11000 {
			t.Fatalf("bad length: %d", len(nums.String()))
		}
		nf := NewNumberFormatter(nums)
		var next = nf.Next()
		for next != "" {
			if nf.Err() != nil {
				t.Fatal(nf.Err())
			}
			if len(next) != 17 {
				t.Fatalf("bad length: %d", len(next))
			}
			next = nf.Next()
		}
	})
	t.Run("error", func(t *testing.T) {
		nf := NewNumberFormatter(strings.NewReader("123456789"))
		nf.Next()
		if nf.Err() == nil {
			t.Fatal("expected error")
		}
		nf.src = strings.NewReader(mikejones)
		nf.Next()
		if nf.Err() == nil {
			t.Fatal("expected error without reset")
		}
		nf.Reset(strings.NewReader(mikejones))
		res := nf.Next()
		if nf.Err() != nil {
			t.Fatal(nf.Err())
		}
		if res != mikejonesf {
			t.Fatalf("got: %s", res)
		}
		nf.Reset(strings.NewReader("1234a6789"))
		nf.Next()
		if nf.Err() == nil {
			t.Fatal("expected error after reset with invalid character")
		}
	})
	t.Run("FormatNumberString", func(t *testing.T) {
		res, err := FormatNumberString(mikejones)
		if err != nil {
			t.Fatal(err)
		}
		if res != mikejonesf {
			t.Fatalf("got: %s", res)
		}
	})
	t.Run("FormatNumberBytes", func(t *testing.T) {
		res, err := FormatNumberBytes([]byte(mikejones))
		if err != nil {
			t.Fatal(err)
		}
		if res != mikejonesf {
			t.Fatalf("got: %s", res)
		}
	})
	t.Run("reader", func(t *testing.T) {
		nf := NewNumberFormatter(strings.NewReader(mikejones))

		out := make([]byte, 17)
		n, err := nf.Read(out)
		if err != nil {
			t.Fatal(err)
		}
		if n != 17 {
			t.Fatalf("expected 17 bytes, got: %d", n)
		}
		if string(out) != mikejonesf {
			t.Fatalf("got: %s", string(out))
		}
	})
	t.Run("writer", func(t *testing.T) {
		nf := NewNumberFormatter(nil)
		if n, err := nf.Write([]byte("hello")); err == nil {
			t.Fatalf("expected error reading letters, got: %d", n)
		}

		n, err := nf.Write([]byte(mikejones))
		if err != nil {
			t.Fatal(err)
		}
		if n != 11 {
			t.Fatalf("expected 11 bytes, got: %d", n)
		}

		t.Logf("input:\n%s", mikejones)

		out2 := make([]byte, 17)
		n, err = nf.Read(out2)
		if err != nil {
			t.Fatal(err)
		}
		if n != 17 {
			t.Errorf("expected 17 bytes, got: %d", n)
		}
		if string(out2) != mikejonesf {
			t.Fatalf("got: %s", string(out2))
		}

		t.Logf("output:\n%s", string(out2))

		if n, err = nf.Write([]byte(mikejones)); err != nil {
			t.Fatal(err)
		}
		if n != 11 {
			t.Fatalf("expected 11 bytes, got: %d", n)
		}
		if n, err = nf.Write([]byte(mikejones)); err != nil {
			t.Fatal(err)
		}
		if n != 11 {
			t.Fatalf("expected 11 bytes, got: %d", n)
		}
		t.Logf("input:\n%s", mikejones+mikejones)

		out3 := make([]byte, 35)
		if n, err = nf.Read(out3); err != nil {
			t.Fatal(err)
		}
		if n != 35 {
			t.Errorf("expected 35 bytes, got: %d", n)
		}
		if string(out3) != mikejonesf+"\n"+mikejonesf {
			t.Fatalf("expected:%s\n\ngot:\n%s\n\n", mikejonesf+"\n"+mikejonesf, string(out3))
		}
		t.Logf("output:\n%s", string(out3))

		t.Run("ridiculous", func(t *testing.T) {
			ridiculous := []byte(`"1 2 8 1))) MIKE JONES 33 oh (0 rather) eight 8 zero 0 zero 0 fo' 4"" :^)

1555 

how's that one song go lmao its like 867 ... uh, 5309? i think?

idfk`)

			if n, err = nf.Write(ridiculous); err != nil {
				t.Fatal(err)
			}
			if n != 22 {
				t.Fatalf("expected 11 bytes, got: %d", n)
			}
			t.Logf("input:\n%s", ridiculous)
			out4 := make([]byte, 35)
			if n, err = nf.Read(out4); err != nil {
				t.Fatal(err)
			}
			if n != 35 {
				t.Errorf("expected 35 bytes, got: %d", n)
			}

			if string(out4) != "+1 (281) 330-8004\n+1 (555) 867-5309" {
				t.Errorf("expected:\n+1 (281) 330-8004\n+1 (555) 867-5309\ngot:\n%s", string(out4))
			}

			t.Logf("output:\n%s", string(out4))

			for out := nf.Next(); out != ""; out = nf.Next() {
				t.Logf("%s", out)
			}
		})
	})
}

var benchCorpus []byte

type loopingReader struct {
	r     *bytes.Reader
	i     int64
	bench *testing.B
}

func (l *loopingReader) Read(p []byte) (n int, err error) {
	l.bench.Helper()
	if l.r == nil {
		panic("nil reader")
	}
	if l.i >= int64(len(benchCorpus)) {
		_, _ = l.r.Seek(0, io.SeekStart)
		l.i = 0
	}

	n, err = l.r.Read(p)
	if err != nil && !errors.Is(err, io.EOF) {
		l.bench.Fatal(err)
	}
	l.i += int64(n)
	if n == len(p) {
		return
	}
	l.bench.StopTimer()
	if n < len(p) {
		// handle short reads by looping back around
		n2, err2 := io.LimitReader(l.r, int64(len(p)-n)).Read(p)
		if err2 != nil && !errors.Is(err2, io.EOF) {
			l.bench.Fatal(err2)
		}
		n += n2
	}
	l.bench.StartTimer()
	return
}

var lrd *loopingReader

func init() {
	rando := entropy.GetSharedRand()
	benchCorpusBuilder := new(bytes.Buffer)
	for i := 0; i < 11; i++ {
		benchCorpusBuilder.WriteString(strconv.Itoa(numbers[rando.Int()%len(numbers)]))
	}
	benchCorpus = bytes.Repeat(benchCorpusBuilder.Bytes(), 500)
	lrd = &loopingReader{r: bytes.NewReader(benchCorpus), i: 0}
}

func TestBenchmark(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping benchmark comparison in short mode")
	}

	printResults := func(name string, bres testing.BenchmarkResult) {
		t.Logf("\n\n"+
			"--------------------------------------\n"+
			name+" benchmark \n"+
			"--------------------------------------\n"+
			"took %d nanoseconds per phone number\n"+
			"allocated memory %d times per phone number\n"+
			"allocated %d bytes of memory per number\n"+
			"it took %s to format %d phone numbers\n"+
			"--------------------------------------\n\n",
			bres.NsPerOp(),
			bres.AllocsPerOp(),
			bres.AllocedBytesPerOp(),
			bres.T, bres.N,
		)
	}

	nextResults := testing.Benchmark(func(b *testing.B) {
		lrd.bench = b
		b.ReportAllocs()
		nf := NewNumberFormatter(lrd)
		for i := 0; i < b.N; i++ {
			if nf.Err() != nil {
				panic(nf.Err())
			}
			asdf := nf.Next()
			if asdf == "" {
				panic("empty corpus")
			}
			if len(asdf) != 17 {
				panic("bad length")
			}
		}
	})
	printResults("NumberFormatter.Next()", nextResults)

	readResults := testing.Benchmark(func(b *testing.B) {
		lrd.bench = b
		b.ReportAllocs()
		out := make([]byte, 17)
		nf := NewNumberFormatter(lrd)
		for i := 0; i < b.N; i++ {
			if i > 0 {
				out = out[:0]
				out = out[:17]
			}
			n, err := nf.Read(out)
			if errors.Is(err, io.EOF) {
				panic("empty corpus")
			}
			if err != nil {
				panic(err)
			}
			if n != 17 && string(out) != "" {
				panic("bad length")
			}
		}
	})
	printResults("NumberFormatter.Read()", readResults)

	fmtResults := testing.Benchmark(func(b *testing.B) {
		lrd.bench = b
		b.ReportAllocs()
		inbuf := make([]byte, 11)
		for i := 0; i < b.N; i++ {
			if i > 0 {
				inbuf = inbuf[:0]
				inbuf = inbuf[:11]
			}
			_, err := lrd.Read(inbuf)
			if errors.Is(err, io.EOF) {
				panic("empty corpus")
			}
			if err != nil {
				panic(err)
			}
			outbuf := fmt.Sprintf("+%s (%s) %s-%s", string(inbuf[0]), inbuf[1:4], inbuf[5:8], inbuf[6:10])
			if len(outbuf) != 17 {
				panic("bad length")
			}
		}
	})
	printResults("fmt.Sprintf()", fmtResults)

	if readResults.NsPerOp() > fmtResults.NsPerOp() {
		t.Errorf("our Read() operation is slower than fmt.Sprintf(), this package is pointless")
	}

	if readResults.AllocsPerOp() > fmtResults.AllocsPerOp() {
		t.Errorf("our Read() operation allocates more memory than fmt.Sprintf(), this package is pointless")
	}

	if readResults.AllocedBytesPerOp() > fmtResults.AllocedBytesPerOp() {
		t.Errorf("our Read() operation allocates more bytes than fmt.Sprintf(), this package is pointless")
	}

	if nextResults.NsPerOp() > fmtResults.NsPerOp() {
		t.Errorf("our Next() operation is slower than fmt.Sprintf(), this package is pointless")
	}

	if nextResults.AllocsPerOp() > fmtResults.AllocsPerOp() {
		t.Errorf("our Next() operation allocates more memory than fmt.Sprintf(), this package is pointless")
	}

	if nextResults.AllocedBytesPerOp() > fmtResults.AllocedBytesPerOp() {
		t.Errorf("our Next() operation allocates more bytes than fmt.Sprintf(), this package is pointless")
	}

	t.Logf("\n\nRead() is %d%% faster than fmt.Sprintf(),\n"+
		"makes %d%% less allocations to allocate %d%% less bytes.\n\n",
		(fmtResults.NsPerOp()-readResults.NsPerOp())*100/fmtResults.NsPerOp(),
		(fmtResults.AllocsPerOp()-readResults.AllocsPerOp())*100/fmtResults.AllocsPerOp(),
		(fmtResults.AllocedBytesPerOp()-readResults.AllocedBytesPerOp())*100/fmtResults.AllocedBytesPerOp(),
	)

	t.Logf("\n\nNext() is %d%% faster than fmt.Sprintf(),\n"+
		"makes %d%% less allocations to allocate %d%% less bytes.\n\n",
		(fmtResults.NsPerOp()-nextResults.NsPerOp())*100/fmtResults.NsPerOp(),
		(fmtResults.AllocsPerOp()-nextResults.AllocsPerOp())*100/fmtResults.AllocsPerOp(),
		(fmtResults.AllocedBytesPerOp()-nextResults.AllocedBytesPerOp())*100/fmtResults.AllocedBytesPerOp(),
	)
}
