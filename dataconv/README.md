# dataconv

Import path: `github.com/InsideGallery/core/dataconv`

## Overview

`dataconv` contains IP conversion helpers and a struct/map merge wrapper used by shared services.

## Main APIs

- `IPV4ToIPV6`, `IPV6ToString`, `IP2Int`, `IPv4ToInt`, `IPv6ToInt`, and `IPv6ToBigInt` convert IP values to
  string, integer, or high/low forms.
- `IntToIPv4`, `IntToIPv6`, and `BigIntToIPv6` convert integer forms back to `net.IP`.
- `ParseIP` wraps `net.ParseIP` and also returns the IP byte length (`net.IPv4len` or `net.IPv6len`).
- `CutIP` masks the last byte of a parsed IP address and returns an empty string for empty or invalid input.
- `MergeStruct` wraps `mergo.Merge` to fill zero values in a destination from a source.
- `ErrInvalidIPAddress`, `ErrNotIPv4Address`, and `ErrNotIPv6Address` identify parse and version failures.

## Usage

```go
ip := net.ParseIP("192.168.1.1")

value, err := dataconv.IPv4ToInt(ip)
if err != nil {
	return err
}

sameIP := dataconv.IntToIPv4(value)
masked := dataconv.CutIP(ip.String())

_ = sameIP
_ = masked
```

## Notes

`IPv4ToInt` returns `ErrNotIPv4Address` when `To4` fails. `ParseIP` returns `ErrInvalidIPAddress` with a nil IP
and zero length for invalid input. `MergeStruct` requires a pointer destination; existing non-zero destination
fields are preserved by `mergo`.
