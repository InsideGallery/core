# dataconv

Helpers to convert data to and from binary format.
Example:
```go
// we want to encode some data into binary format
b := NewBinaryEncoder()
testutils.Equal(t, b.Encode(uint8(1)), nil) // convert uint8: 1 to binary, with offset 0, return new offset
testutils.Equal(t, b.Encode(uint16(1245)), nil) // convert uin16: 1235 to binary by given offset, and return new offset
testutils.Equal(t, b.Encode(uint32(1245678)), nil)
testutils.Equal(t, b.Encode(uint64(124567891011)), nil)
testutils.Equal(t, b.Encode([]byte("test")), nil)
testutils.Equal(t, b.Encode(12.56), nil)
data := b.Bytes() // here we should have filled data with different values

// that how we can take data back:
d := NewBinaryDecoder(data)
var tuint8 uint8
err := d.Decode(&tuint8)
testutils.Equal(t, err, nil)
testutils.Equal(t, tuint8, uint8(1))
var tuint16 uint16
err = d.Decode(&tuint16)
testutils.Equal(t, err, nil)
testutils.Equal(t, tuint16, uint16(1245))
var tuint32 uint32
err = d.Decode(&tuint32)
testutils.Equal(t, err, nil)
testutils.Equal(t, tuint32, uint32(1245678))
var tuint64 uint64
err = d.Decode(&tuint64)
testutils.Equal(t, err, nil)
testutils.Equal(t, tuint64, uint64(124567891011))
var tbytes []byte
err = d.Decode(&tbytes)
testutils.Equal(t, err, nil)
testutils.Equal(t, tbytes, []byte("test"))
var tfloat64 float64
err = d.Decode(&tfloat64)
testutils.Equal(t, err, nil)
testutils.Equal(t, tfloat64, float64(12.56))
```