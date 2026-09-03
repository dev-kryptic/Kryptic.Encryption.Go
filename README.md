# Kryptic.Encryption (Go)

The Go implementation of Kryptic's open-source (Apache-2.0) encryption engine. This is
the module the **daemon, CLI, and Kubernetes operator** use to unwrap machine
keys, open sealed-box grants, and decrypt secret envelopes locally. The
platform never sees those values in plaintext.

**Go module:** `github.com/dev-kryptic/Kryptic.Encryption.Go`. This GitHub
repository is named `Kryptic.Encryption.Go` so auditors can tell the three
runtimes apart.

Sibling implementations of the same wire formats:

| Repository | Runtime | Consumed by |
| --- | --- | --- |
| [Kryptic.Encryption.Dotnet](https://github.com/dev-kryptic/Kryptic.Encryption.Dotnet) | .NET (`Kryptic.Encryption` on nuget.org) | Kryptic Platform |
| [Kryptic.Encryption.NPM](https://github.com/dev-kryptic/Kryptic.Encryption.NPM) | TypeScript / WebCrypto (`@krypticdev/encryption`) | Management dashboard |
| [Kryptic.Encryption.Go](https://github.com/dev-kryptic/Kryptic.Encryption.Go) | Go (this module) | Daemon, CLI, Kubernetes operator |

A format change (envelope, sealed box, Argon2id parameters) must land in all three
repositories in the same release. The committed files in `interop-vectors/` are the
contract: every runtime must open and, where the test is deterministic, reproduce
those bytes.

**No custom primitives.** AES-256-GCM and P-256 ECDH via the Go standard library;
Argon2id via `golang.org/x/crypto/argon2`. Read [SECURITY.md](SECURITY.md) before
reading code.

## Install

```
go get github.com/dev-kryptic/Kryptic.Encryption.Go@v1.0.0
```

Requires Go 1.25.

## What's in the box

| Package | Purpose |
| --- | --- |
| `sealedbox` | P-256 ECDH sealed box (`sbx.v1...`) for delivering the org key |
| `envelope` | AES-256-GCM secret envelope (`v1.<keyId>...`), including `SecretContext` |
| `kdf` | Argon2id passphrase / client-secret -> 256-bit key (parameter set v1) |

## Usage

### Open a secrets bundle (daemon, CLI, operator)

```go
import (
    "github.com/dev-kryptic/Kryptic.Encryption.Go/envelope"
    "github.com/dev-kryptic/Kryptic.Encryption.Go/kdf"
    "github.com/dev-kryptic/Kryptic.Encryption.Go/sealedbox"
)

wrapKey, err := kdf.ForVersion(keys.KdfParametersVersion, clientSecret, salt)
privateKey, err := envelope.Open(wrapKey, keys.WrappedPrivateKey, nil)

box, err := sealedbox.Parse(bundle.WrappedOrgKey)
orgKey, err := sealedbox.Open(sealedbox.KeyPair{Public: publicKey, Private: privateKey}, box)

plaintext, err := envelope.Open(
    orgKey,
    entry.Envelope,
    envelope.SecretContext(entry.DefinitionId, entry.EnvironmentId),
)
```

The associated data is `secret:{definitionId}:env:{environmentId}` (lowercase
GUIDs). Moving a ciphertext to another row fails decryption.

### Seal a grant (tests, future write paths)

```go
device, err := sealedbox.GenerateKeyPair()
grant, err := sealedbox.Seal(device.Public, "device-key-1", orgKey)
opened, err := sealedbox.Open(device, grant)
```

## Build & test

```
go test ./...
```

## Publishing (maintainers)

Go modules are published by **git tags**, not a package registry token.

CI lives in [`.github/workflows/publish.yml`](.github/workflows/publish.yml). Pull
requests only run tests. A release on `main` (or `workflow_dispatch`) commits
`VERSION`, pushes it as the Kryptic Release Bot, tags `vX.Y.Z`, and opens a
GitHub Release. The workflow pings `proxy.golang.org` so pkg.go.dev can index
the tag.

### GitHub Actions secrets

Add these at the org or on the GitHub repo (`Settings` > `Secrets and variables` > `Actions`):

| Secret | What it is | Where to get it |
| --- | --- | --- |
| `RELEASE_BOT_APP_ID` | App ID of **Kryptic Release Bot** | GitHub App settings |
| `RELEASE_BOT_PRIVATE_KEY` | Private key `.pem` of that app | GitHub App settings > Generate a private key |

Do not add npm or NuGet tokens to this repo.

### First publish

1. Create the public GitHub repository `dev-kryptic/Kryptic.Encryption.Go`.
2. Add the Release Bot secrets (org-wide is enough).
3. Push `main`. The workflow tags `v1.0.0` from `VERSION` on the first run.

Keep major.minor aligned with the .NET and npm packages when the wire format
changes. To ship `1.1.0` or `2.0.0`, set `VERSION` (or pass it to
`workflow_dispatch`).

## Reporting vulnerabilities

Please report security issues to **security@kryptic.dev**. See
[SECURITY.md](SECURITY.md). Do not open public issues for vulnerabilities.

## License

[Apache-2.0](LICENSE)
