# stdx/strings

Import path: `github.com/InsideGallery/core/stdx/strings`

## Overview

`stdx/strings` provides hashing, email cleanup, normalization, ID, masking, chunking, and password-display
helpers. Import it with an alias when the standard library `strings` package is also needed.

## Main APIs

- `CRC32`, `CRC16`, `SimHash`, `SimHashCompare`, and `HashName` provide hash helpers.
- `ABTest` maps input and salt bytes into a deterministic bucket below the sum of the supplied group weights.
- `EmailUserName`, `EmailDomain`, and `SanitizeEmail` split and normalize email-like strings.
- `NFDLowerString`, `NFKDLowerString`, `CommonString`, `SplitBetweenTokens`, and `Between` normalize or extract
  text.
- `ByteSliceToString` casts bytes to string without allocation.
- `GetUniqueID`, `GetShortID`, `GetTinyID`, and `RandStringBytes` create identifiers or random strings.
- `SafeGet`, `MaskField`, and `SplitByChunks` provide small generic/string helpers.
- `ContextKey`, `Password`, and `Str` are compatibility types for context keys and masked password values.

## Usage

```go
email := strx.SanitizeEmail("User+tag@example.com")
bucket := strx.ABTest([]byte("user-1"), []byte("experiment"), 50, 50)
masked := strx.Password("secret").String()

_ = email
_ = bucket
_ = masked
```

## Notes

`Password.String` and JSON marshaling always return `********`; use `Value` only when the real secret is needed.
`ByteSliceToString` uses `unsafe`, so callers should not mutate the byte slice after casting. `HashName` expects
a non-empty input string. `SplitByChunks` and shorthands in this package operate on byte indexes, not runes.
