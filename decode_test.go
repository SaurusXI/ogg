// © 2016 Steve McCoy under the MIT license. See LICENSE for details.

package ogg

import (
	"bytes"
	"io"
	"math/rand"
	"strings"
	"testing"
	"time"
)

func TestBasicDecode(t *testing.T) {
	var b bytes.Buffer
	e := NewEncoder(1, &b)

	err := e.EncodeBOS(2, [][]byte{[]byte("hello")})
	if err != nil {
		t.Fatal("unexpected EncodeBOS error:", err)
	}

	d := NewDecoder(&b)

	p, _, err := d.Decode()
	if err != nil {
		t.Fatal("unexpected Decode error:", err)
	}

	if p.Type != BOS {
		t.Fatal("expected BOS, got", p.Type)
	}

	if p.Serial != 1 {
		t.Fatal("expected serial 1, got", p.Serial)
	}

	if p.Granule != 2 {
		t.Fatal("expected granule 2, got", p.Granule)
	}

	expect := []byte{
		'h', 'e', 'l', 'l', 'o',
	}

	if len(p.Packets) != 1 {
		t.Fatalf("len(p.Packets) = %d", len(p.Packets))
	}

	if !bytes.Equal(p.Packets[0], expect) {
		t.Fatalf("bytes != expected:\n%x\n%x", p.Packets[0], expect)
	}
}

func TestBasicMultiDecode(t *testing.T) {
	var b bytes.Buffer
	e := NewEncoder(1, &b)

	err := e.EncodeBOS(2, [][]byte{[]byte("hello")})
	if err != nil {
		t.Fatal("unexpected EncodeBOS error:", err)
	}
	err = e.Encode(7, [][]byte{[]byte("there")})
	if err != nil {
		t.Fatal("unexpected Encode error:", err)
	}

	d := NewDecoder(&b)

	p, _, err := d.Decode()
	if err != nil {
		t.Fatal("unexpected Decode error:", err)
	}

	if p.Type != BOS {
		t.Fatal("expected BOS, got", p.Type)
	}

	if p.Serial != 1 {
		t.Fatal("expected serial 1, got", p.Serial)
	}

	if p.Granule != 2 {
		t.Fatal("expected granule 2, got", p.Granule)
	}

	expect := []byte{
		'h', 'e', 'l', 'l', 'o',
	}

	if len(p.Packets) != 1 {
		t.Fatalf("len(p.Packets) = %d", len(p.Packets))
	}

	if !bytes.Equal(p.Packets[0], expect) {
		t.Fatalf("bytes != expected:\n%x\n%x", p.Packets[0], expect)
	}

	p, _, err = d.Decode()
	if err != nil {
		t.Fatal("unexpected Decode error:", err)
	}

	if p.Type != 0 {
		t.Fatal("expected normal type, got", p.Type)
	}

	if p.Serial != 1 {
		t.Fatal("expected serial 1, got", p.Serial)
	}

	if p.Granule != 7 {
		t.Fatal("expected granule 7, got", p.Granule)
	}

	expect = []byte{
		't', 'h', 'e', 'r', 'e',
	}

	if len(p.Packets) != 1 {
		t.Fatalf("len(p.Packets) = %d", len(p.Packets))
	}

	if !bytes.Equal(p.Packets[0], expect) {
		t.Fatalf("bytes != expected:\n%x\n%x", p.Packets[0], expect)
	}
}

func TestMultipacketDecode(t *testing.T) {
	var b bytes.Buffer
	e := NewEncoder(1, &b)

	err := e.EncodeBOS(2, [][]byte{[]byte("hello"), []byte("there")})
	if err != nil {
		t.Fatal("unexpected EncodeBOS error:", err)
	}

	d := NewDecoder(&b)

	p, _, err := d.Decode()
	if err != nil {
		t.Fatal("unexpected Decode error:", err)
	}

	if p.Type != BOS {
		t.Fatal("expected BOS, got", p.Type)
	}

	if p.Serial != 1 {
		t.Fatal("expected serial 1, got", p.Serial)
	}

	if p.Granule != 2 {
		t.Fatal("expected granule 2, got", p.Granule)
	}

	expect := []byte{
		'h', 'e', 'l', 'l', 'o',
	}

	if len(p.Packets) != 2 {
		t.Fatalf("len(p.Packets) = %d", len(p.Packets))
	}

	if !bytes.Equal(p.Packets[0], expect) {
		t.Fatalf("bytes != expected:\n%x\n%x", p.Packets[0], expect)
	}

	expect = []byte{
		't', 'h', 'e', 'r', 'e',
	}

	if !bytes.Equal(p.Packets[1], expect) {
		t.Fatalf("bytes != expected:\n%x\n%x", p.Packets[0], expect)
	}
}

func TestBadCrc(t *testing.T) {
	var b bytes.Buffer
	e := NewEncoder(1, &b)

	err := e.EncodeBOS(2, [][]byte{[]byte("hello")})
	if err != nil {
		t.Fatal("unexpected EncodeBOS error:", err)
	}

	b.Bytes()[22] = 0

	d := NewDecoder(&b)

	_, _, err = d.Decode()
	if err == nil {
		t.Fatal("unexpected lack of Decode error")
	}
	if bs, ok := err.(ErrBadCrc); !ok {
		t.Fatal("exected ErrBadCrc, got:", err)
	} else if !strings.HasPrefix(bs.Error(), "invalid crc in packet") {
		t.Fatalf("the error message looks wrong: %q", err.Error())
	}
}

func TestShortDecode(t *testing.T) {
	var b bytes.Buffer
	d := NewDecoder(&b)
	_, _, err := d.Decode()
	if err != io.EOF {
		t.Fatal("expected EOF, got:", err)
	}

	e := NewEncoder(1, &b)
	err = e.Encode(2, [][]byte{[]byte("hello")})
	if err != nil {
		t.Fatal("unexpected Encode error:", err)
	}
	d = NewDecoder(&io.LimitedReader{R: &b, N: headsz})
	_, _, err = d.Decode()
	if err != io.EOF {
		t.Fatal("expected EOF, got:", err)
	}

	b.Reset()
	e = NewEncoder(1, &b)
	err = e.Encode(2, [][]byte{[]byte("hello")})
	if err != nil {
		t.Fatal("unexpected Encode error:", err)
	}
	d = NewDecoder(&io.LimitedReader{R: &b, N: int64(b.Len()) - 1})
	_, _, err = d.Decode()
	if err != io.ErrUnexpectedEOF {
		t.Fatal("expected ErrUnexpectedEOF, got:", err)
	}
}

func TestBadSegs(t *testing.T) {
	var b bytes.Buffer
	e := NewEncoder(1, &b)

	err := e.EncodeBOS(2, [][]byte{[]byte("hello")})
	if err != nil {
		t.Fatal("unexpected EncodeBOS error:", err)
	}

	b.Bytes()[26] = 0

	d := NewDecoder(&b)
	_, _, err = d.Decode()
	if err != ErrBadSegs {
		t.Fatal("expected ErrBadSegs, got:", err)
	}
}

func TestSyncDecode(t *testing.T) {
	var b bytes.Buffer
	for i := 0; i < headsz-1; i++ {
		b.Write([]byte("x"))
	}
	b.Write([]byte("O"))

	for i := 0; i < headsz-3; i++ {
		b.Write([]byte("x"))
	}
	b.Write([]byte("Og"))

	for i := 0; i < headsz-5; i++ {
		b.Write([]byte("x"))
	}
	b.Write([]byte("Ogg"))

	e := NewEncoder(1, &b)

	err := e.EncodeBOS(2, [][]byte{[]byte("hello")})
	if err != nil {
		t.Fatal("unexpected EncodeBOS error:", err)
	}

	d := NewDecoder(&b)

	p, _, err := d.Decode()
	if err != nil {
		t.Fatal("unexpected Decode error:", err)
	}

	if p.Type != BOS {
		t.Fatal("expected BOS, got", p.Type)
	}

	if p.Serial != 1 {
		t.Fatal("expected serial 1, got", p.Serial)
	}

	if p.Granule != 2 {
		t.Fatal("expected granule 2, got", p.Granule)
	}

	expect := []byte{
		'h', 'e', 'l', 'l', 'o',
	}

	if len(p.Packets) != 1 {
		t.Fatalf("len(p.Packets) = %d", len(p.Packets))
	}

	if !bytes.Equal(p.Packets[0], expect) {
		t.Fatalf("bytes != expected:\n%x\n%x", p.Packets[0], expect)
	}
}

func TestLongDecode(t *testing.T) {
	var b bytes.Buffer
	e := NewEncoder(1, &b)

	var junk bytes.Buffer
	for i := 0; i < maxPageSize*2; i++ {
		c := byte(rand.Intn(26)) + 'a'
		junk.WriteByte(c)
	}

	err := e.Encode(2, [][]byte{junk.Bytes()})
	if err != nil {
		t.Fatal("unexpected Encode error:", err)
	}

	d := NewDecoder(&b)
	p1, _, err := d.Decode()
	if err != nil {
		t.Fatal("unexpected Decode error:", err)
	}
	if p1.Type != 0 {
		t.Fatal("unexpected page type:", p1.Type)
	}
	if len(p1.Packets) != 1 {
		t.Fatalf("len(p1.Packets) = %d", len(p1.Packets))
	}
	if !bytes.Equal(p1.Packets[0], junk.Bytes()[:mps]) {
		t.Fatalf("packet is wrong:\n\t%x\nvs\n\t%x\n", p1.Packets[0], junk.Bytes()[:mps])
	}

	p2, _, err := d.Decode()
	if err != nil {
		t.Fatal("unexpected Decode error:", err)
	}
	if p2.Type != COP {
		t.Fatal("unexpected page type:", p1.Type)
	}
	if len(p2.Packets) != 1 {
		t.Fatalf("len(p2.Packets) = %d", len(p2.Packets))
	}
	if !bytes.Equal(p2.Packets[0], junk.Bytes()[mps:mps+mps]) {
		t.Fatalf("packet is wrong:\n\t%x\nvs\n\t%x\n", p2.Packets[0], junk.Bytes()[mps:mps+mps])
	}

	p3, _, err := d.Decode()
	if err != nil {
		t.Fatal("unexpected Decode error:", err)
	}
	if p3.Type != COP {
		t.Fatal("unexpected page type:", p1.Type)
	}
	if len(p3.Packets) != 1 {
		t.Fatalf("len(p3.Packets) = %d", len(p3.Packets))
	}
	rem := (maxPageSize * 2) - mps*2
	if !bytes.Equal(p3.Packets[0], junk.Bytes()[mps*2:mps*2+rem]) {
		t.Fatalf("packet is wrong:\n\t%x\nvs\n\t%x\n", p3.Packets[0], junk.Bytes()[mps*2:mps*2+rem])
	}
}

func TestLongMultipacketDecode(t *testing.T) {
	var b bytes.Buffer
	e := NewEncoder(1, &b)

	var junk bytes.Buffer
	for i := 0; i < maxPageSize*2; i++ {
		c := byte(rand.Intn(26)) + 'a'
		junk.WriteByte(c)
	}

	err := e.Encode(2, [][]byte{junk.Bytes()[:50], junk.Bytes()[50:]})
	if err != nil {
		t.Fatal("unexpected Encode error:", err)
	}

	d := NewDecoder(&b)
	p1, _, err := d.Decode()
	if err != nil {
		t.Fatal("unexpected Decode error:", err)
	}
	if p1.Type != 0 {
		t.Fatal("unexpected page type:", p1.Type)
	}
	if len(p1.Packets) != 2 {
		t.Fatalf("len(p1.Packets) = %d", len(p1.Packets))
	}
	if !bytes.Equal(p1.Packets[0], junk.Bytes()[:50]) {
		t.Fatalf("packet is wrong:\n\t%x\nvs\n\t%x\n", p1.Packets[0], junk.Bytes()[:50])
	}
	if len(p1.Packets[1]) != mps-mss {
		t.Fatalf("packet is wrong size: %d vs. %d", len(p1.Packets[1]), mps-mss)
	}
	if !bytes.Equal(p1.Packets[1], junk.Bytes()[50:50+mps-mss]) {
		t.Fatalf("packet is wrong:\n\t%x\nvs\n\t%x\n", p1.Packets[1], junk.Bytes()[50:50+mps-mss])
	}

	p2, _, err := d.Decode()
	if err != nil {
		t.Fatal("unexpected Decode error:", err)
	}
	if p2.Type != COP {
		t.Fatal("unexpected page type:", p1.Type)
	}
	if len(p2.Packets) != 1 {
		t.Fatalf("len(p2.Packets) = %d", len(p2.Packets))
	}
	if len(p2.Packets[0]) != mps {
		t.Fatalf("packet is wrong size: %d vs. %d", len(p2.Packets[0]), mps)
	}

	start := 50 + mps - mss
	if !bytes.Equal(p2.Packets[0], junk.Bytes()[start:start+mps]) {
		t.Fatalf("packet is wrong:\n\t%x\nvs\n\t%x\n", p2.Packets[0], junk.Bytes()[start:start+mps])
	}

	p3, _, err := d.Decode()
	if err != nil {
		t.Fatal("unexpected Decode error:", err)
	}
	if p3.Type != COP {
		t.Fatal("unexpected page type:", p1.Type)
	}
	if len(p3.Packets) != 1 {
		t.Fatalf("len(p3.Packets) = %d", len(p3.Packets))
	}
	start += mps
	if !bytes.Equal(p3.Packets[0], junk.Bytes()[start:]) {
		t.Fatalf("packet is wrong:\n\t%x\nvs\n\t%x\n", p3.Packets[0], junk.Bytes()[start:])
	}
}

func TestEvenLongerMultipacketDecode(t *testing.T) {
	var b bytes.Buffer
	e := NewEncoder(1, &b)

	var junk bytes.Buffer
	for i := 0; i < maxPageSize*2; i++ {
		c := byte(rand.Intn(26)) + 'a'
		junk.WriteByte(c)
	}

	err := e.Encode(2, [][]byte{
		junk.Bytes()[:50],
		junk.Bytes()[50 : junk.Len()-13],
		junk.Bytes()[junk.Len()-13:],
	})
	if err != nil {
		t.Fatal("unexpected Encode error:", err)
	}

	d := NewDecoder(&b)
	p1, _, err := d.Decode()
	if err != nil {
		t.Fatal("unexpected Decode error:", err)
	}
	if p1.Type != 0 {
		t.Fatal("unexpected page type:", p1.Type)
	}
	if len(p1.Packets) != 2 {
		t.Fatalf("len(p1.Packets) = %d", len(p1.Packets))
	}
	if !bytes.Equal(p1.Packets[0], junk.Bytes()[:50]) {
		t.Fatalf("packet is wrong:\n\t%x\nvs\n\t%x\n", p1.Packets[0], junk.Bytes()[:50])
	}
	if len(p1.Packets[1]) != mps-mss {
		t.Fatalf("packet is wrong size: %d vs. %d", len(p1.Packets[1]), mps-mss)
	}
	if !bytes.Equal(p1.Packets[1], junk.Bytes()[50:50+mps-mss]) {
		t.Fatalf("packet is wrong:\n\t%x\nvs\n\t%x\n", p1.Packets[1], junk.Bytes()[50:50+mps-mss])
	}

	p2, _, err := d.Decode()
	if err != nil {
		t.Fatal("unexpected Decode error:", err)
	}
	if p2.Type != COP {
		t.Fatal("unexpected page type:", p1.Type)
	}
	if len(p2.Packets) != 1 {
		t.Fatalf("len(p2.Packets) = %d", len(p2.Packets))
	}
	if len(p2.Packets[0]) != mps {
		t.Fatalf("packet is wrong size: %d vs. %d", len(p2.Packets[0]), mps)
	}

	start := 50 + mps - mss
	if !bytes.Equal(p2.Packets[0], junk.Bytes()[start:start+mps]) {
		t.Fatalf("packet is wrong:\n\t%x\nvs\n\t%x\n", p2.Packets[0], junk.Bytes()[start:start+mps])
	}

	p3, _, err := d.Decode()
	if err != nil {
		t.Fatal("unexpected Decode error:", err)
	}
	if p3.Type != COP {
		t.Fatal("unexpected page type:", p3.Type)
	}
	if len(p3.Packets) != 2 {
		t.Fatalf("len(p3.Packets) = %d", len(p3.Packets))
	}
	start += mps
	if !bytes.Equal(p3.Packets[0], junk.Bytes()[start:junk.Len()-13]) {
		t.Fatalf("packet is wrong:\n\t%x\nvs\n\t%x\n", p3.Packets[0], junk.Bytes()[start:start+mps-13])
	}
	start = junk.Len() - 13
	if !bytes.Equal(p3.Packets[1], junk.Bytes()[start:]) {
		t.Fatalf("packet is wrong:\n\t%x\nvs\n\t%x\n", p3.Packets[0], junk.Bytes()[start:])
	}
}

func TestGetPacketDuration(t *testing.T) {
    tests := []struct {
        name     string
        packet   []byte
        want     time.Duration
        wantErr  bool
        errMsg   string
    }{
        {
            name:    "empty packet",
            packet:  []byte{},
            wantErr: true,
            errMsg:  "empty opus packet",
        },
        {
            name:    "single 10ms frame",
            packet:  []byte{0x00}, // config 0, frame count code 0
            want:    10 * time.Millisecond,
            wantErr: false,
        },
        {
            name:    "single 20ms frame",
            packet:  []byte{0x20}, // config 4, frame count code 0
            want:    20 * time.Millisecond,
            wantErr: false,
        },
        {
            name:    "single 40ms frame",
            packet:  []byte{0x40}, // config 8, frame count code 0
            want:    40 * time.Millisecond,
            wantErr: false,
        },
        {
            name:    "single 60ms frame",
            packet:  []byte{0x60}, // config 12, frame count code 0
            want:    60 * time.Millisecond,
            wantErr: false,
        },
        {
            name:    "two 20ms frames (code 1)",
            packet:  []byte{0x21}, // config 4, frame count code 1
            want:    40 * time.Millisecond,
            wantErr: false,
        },
        {
            name:    "two 20ms frames (code 2)",
            packet:  []byte{0x22}, // config 4, frame count code 2
            want:    40 * time.Millisecond,
            wantErr: false,
        },
        {
            name:    "variable frame count (3 frames)",
            packet:  []byte{0x23, 0x02}, // config 4, frame count code 3, count=2+1
            want:    60 * time.Millisecond,
            wantErr: false,
        },
        {
            name:    "code 3 but packet too short",
            packet:  []byte{0x23}, // config 4, frame count code 3, but no second byte
            wantErr: true,
            errMsg:  "invalid opus packet: frame count code 3 but packet is too short",
        },
    }

    d := NewDecoder(nil) // reader not needed for this test
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := d.GetPacketDuration(tt.packet)
            if tt.wantErr {
                if err == nil {
                    t.Errorf("GetPacketDuration() error = nil, wantErr %v", tt.wantErr)
                }
                if err != nil && err.Error() != tt.errMsg {
                    t.Errorf("GetPacketDuration() error = %v, want %v", err, tt.errMsg)
                }
                return
            }
            if err != nil {
                t.Errorf("GetPacketDuration() unexpected error = %v", err)
                return
            }
            if got != tt.want {
                t.Errorf("GetPacketDuration() = %v, want %v", got, tt.want)
            }
        })
    }
}
