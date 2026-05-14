# antibot

Import path: `github.com/InsideGallery/core/antibot`

## Overview

`antibot` provides a small SHA-256 proof-of-work helper. A challenge is solved when the hex digest of
`message + nonce` starts with the configured number of zero characters.

## Main APIs

- `ProofOfWork` stores the requested `Difficulty`.
- `NewProofOfWork(difficulty int)` creates a proof-of-work checker.
- `(*ProofOfWork).Validate(message string, nonce int)` checks whether a nonce satisfies the challenge.
- `(*ProofOfWork).FindNonce(message string)` searches from nonce `0` upward and returns the nonce and hash.

## Usage

```go
pow := antibot.NewProofOfWork(4)

nonce, hash := pow.FindNonce("Hello, world!")
valid := pow.Validate("Hello, world!", nonce)

_ = hash
_ = valid
```

## Notes

`FindNonce` is CPU-bound and has no context, timeout, or maximum nonce. Callers should choose a difficulty
appropriate for the request path and enforce their own cancellation when needed.
