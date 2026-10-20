# PPPRT — Pairing Proofs and Polynomial Ring Toolkit (paper reference)

> [!WARNING]
> **Paper reference artifact. NOT SOUND — do not use in production.**
> This fork predates two gnark security fixes that apply to the code it
> benchmarks (see [Why this is not sound](#why-this-is-not-sound)). It is
> kept, frozen, only to reproduce the numbers reported in the paper.
>
> **For anything else use [mistcash/polyring](https://github.com/mistcash/polyring)**:
> the same technique as a standalone Go module on patched `gnark`.

Reference implementation for:

> Shramee Srivastav. *Pairing Proofs and Polynomial Ring Toolkit.*
> IACR Transactions on Cryptographic Hardware and Embedded Systems (TCHES), 2027.

## Two repositories, two jobs

| | this repo (`ppprt`) | [`polyring`](https://github.com/mistcash/polyring) |
|---|---|---|
| Purpose | reproduce the paper's Table 7 | use the technique |
| Built on | gnark fork, upstream merged up to `c05c6a8` (3 Aug 2026) | gnark v0.16.4 (patched) via a pinned fork |
| Sound | **no** | yes, unaudited |
| Shape | toolkit inside gnark's `std/math/emulated` | standalone module: `std/polyring`, ring pairing, recursive Groth16 verifier, Solidity verifier |
| Maintained | frozen | yes |

polyring's [`ARTIFACT.md`](https://github.com/mistcash/polyring/blob/main/ARTIFACT.md)
puts the paper's numbers, this repo's re-run and polyring's numbers on patched
gnark side by side, and explains how they relate.

## Why this is not sound

Two gnark advisories published after this fork's last upstream merge apply to
the emulated arithmetic and pairing code this repo builds on:

- [GHSA-3mvx-pp85-pm65](https://github.com/Consensys/gnark/security/advisories/GHSA-3mvx-pp85-pm65)
  (critical, fixed in v0.16.2): the carry polynomial in emulated-field deferred
  checks came from a hint and was never range-checked, so a malicious prover
  could satisfy a false multiplication. Every `std/algebra/emulated` pairing,
  this repo's included, sits on those checks.
- [GHSA-fwf3-jmp7-gmrj](https://github.com/Consensys/gnark/security/advisories/GHSA-fwf3-jmp7-gmrj)
  (medium): the on-twist check on G2 inputs to the Miller loop was computed and
  then discarded.

The fixes add constraints to **both** columns of Table 7, gnark's own pairing
included. This repo is left as it was measured so the published numbers stay
reproducible; polyring carries the fixes.

## Where to look

| Paper | Code |
|---|---|
| Polynomial ring toolkit (Section 6) | `std/math/emulated/field_polyring.go` |
| `NewPolyRingCheck`, `MulPolyRings`, `ToCommit`, `PolyRingAccumulator` (Section 6.3.1) | same file |
| `performDeferredRingChecks`, `polyRingMulHint`, `quotientsRLCHint` (Sections 6.2, 6.3) | same file |
| `InnerProductNoReduce`, deferred reduction (Section 6.5) | same file |
| BN254 pairing with polynomial ring checks (Sections 4–5) | `std/algebra/emulated/sw_bn254/pairing.go`, `std/algebra/emulated/fields_bn254/e12.go` |
| BLS12-381 pairing | `std/algebra/emulated/sw_bls12381/pairing.go` |
| 3-pairing benchmark (Table 7) | `std/algebra/emulated/sw_bn254/g16_simulation_test.go`, `std/algebra/emulated/sw_bls12381/g16_simulation_test.go` |

## Reproducing Table 7

Requirements: Go ≥ 1.25.7 (tested with 1.26.4). No other setup. Each package
takes about 10 s to benchmark after the first build.

### "This Work" column

```sh
go test -v -run '^$' -bench 'Groth16Simulation' -benchmem \
  ./std/algebra/emulated/sw_bn254/ ./std/algebra/emulated/sw_bls12381/
```

Each package prints two lines like

```
g16_simulation_test.go:141: nb commitments: 269656, scs size: … (bytes), nb constraints 1035645, …
g16_simulation_test.go:167: nb commitments: 269656, r1cs size: … (bytes), nb constraints 316181, …
```

plus four benchmark lines (`compile_scs`, `solve_scs`, `compile_r1cs`,
`solve_r1cs`). Lines like `api.Commit …/field_polyring.go:433 commitmentIndex=0`
are a leftover debug print and can be ignored.

### Baseline column

The baseline is the same benchmark file compiled against upstream gnark. It
needs one line of instrumentation from this repo: the `nb commitments` counter
(`scs.NbCommitments`, added to `frontend/cs/scs/api.go` in `8ea4366`), which
upstream does not have. From a checkout of this repo:

```sh
git clone https://github.com/Consensys/gnark ../gnark-baseline
cd ../gnark-baseline
git checkout c05c6a8109be76360431f478ba5bc8e5969e05cf   # upstream commit this repo is merged on
git -C ../ppprt show 8ea4366 -- frontend/cs/scs/api.go | git apply
for c in sw_bn254 sw_bls12381; do
  cp ../ppprt/std/algebra/emulated/$c/g16_*_test.go std/algebra/emulated/$c/
done
go test -v -run '^$' -bench 'Groth16Simulation' -benchmem \
  ./std/algebra/emulated/sw_bn254/ ./std/algebra/emulated/sw_bls12381/
```

(Adjust `../ppprt` if your checkout of this repo has another name.)

The same steps with `git checkout v0.16.4` instead run the baseline on gnark
with the security fixes: 2,834,581 SCS and 791,907 R1CS constraints on BN254,
3,164,858 and 881,274 on BLS12-381. polyring's `paper/` benchmark reproduces
these to the digit as its gnark column.

The raw output of the baseline run reported in the paper is kept verbatim in
both test files, in the `Benchmark results BEFORE` comment at the bottom; it
matches Table 7's baseline column to the digit.

### Reading the output

| Table 7 row | Benchmark output |
|---|---|
| Nb Constraints | `nb constraints` on the `:141` (SCS) and `:167` (R1CS) lines |
| Nb Fq Committed | `nb commitments`: the number of values passed to `api.Commit`. The counter is only incremented by the SCS builder, so the R1CS line repeats the SCS value: it is the same circuit. |
| Compile Time, Solve Time | `ns/op` of `compile_*` and `solve_*` |
| Compile Memory, Solve Memory | `B/op` of the same lines: bytes allocated per operation (`-benchmem`), not peak memory. The paper's GB and MB are 2³⁰ and 2²⁰ bytes. |

Constraint and commitment counts are deterministic. Times and allocations
depend on the machine; the paper's were taken on an Apple M3 Pro (arm64).

### Expected results

Measured on an Apple M3 Pro, Go 1.26.4.

| | Paper | Re-run (this repo `HEAD`, upstream `c05c6a8`) |
|---|---|---|
| **BN254** SCS constraints | 1,592,440 → 1,035,550 (35.0%) | 1,604,279 → 1,035,645 (35.4%) |
| Fq committed | 571,125 → 269,616 (52.8%) | 606,104 → 269,656 (55.5%) |
| R1CS constraints | 491,447 → 316,167 (35.7%) | 495,810 → 316,181 (36.2%) |
| **BLS12-381** SCS constraints | 1,784,247 → 1,191,105 (33.2%) | 1,779,202 → 1,191,256 (33.0%) |
| Fq committed | 660,160 → 300,104 (54.5%) | 673,107 → 300,166 (55.4%) |
| R1CS constraints | 545,510 → 368,434 (32.5%) | 542,925 → 368,461 (32.1%) |

- **This Work** reproduces to within 0.02% (at most 151 constraints and 62
  committed elements). The exact working tree the paper's column was measured
  on was not committed. At `2f9ffd0` the committed-element counts match the
  paper exactly (269,616 and 300,104) and constraints are 212 (BN254) and 156
  (BLS12-381) below it.
- **Baseline** reproduces to within 1% on constraints and 6% on committed
  elements. The paper's baseline run was recorded on 5 July 2026 (`8ea4366`)
  against the upstream gnark of the time; `c05c6a8` is the upstream commit this
  repo's `HEAD` is merged on, so it re-runs both columns on the same base.
- Every reduction stays within half a percentage point of the paper's, or
  improves on it.

### Committed wires

Table 7's "Nb Fq Committed" sums the values passed to every `api.Commit` call,
so a wire committed twice counts twice. The circuit's commitments themselves
cover fewer, unique wires; that is what the prover commits to, and what
polyring's `paper/` benchmark reports. To count them here (test-only, the
measured code is unchanged):

```sh
go test -v -run CommittedWires ./std/algebra/emulated/sw_bn254/ ./std/algebra/emulated/sw_bls12381/
```

It runs in the baseline checkout above too, since the `cp` copies it.

| Committed wires | Baseline (upstream `c05c6a8`) | This repo `HEAD` | Reduction |
|---|---|---|---|
| BN254 SCS | 381,400 | 245,894 | 35.5% |
| BN254 R1CS | 374,989 | 273,114 | 27.2% |
| BLS12-381 SCS | 428,261 | 273,575 | 36.1% |
| BLS12-381 R1CS | 422,037 | 318,962 | 24.4% |

So the paper's "over 50% fewer committed elements" holds for values passed to
`api.Commit`. Counted as unique committed wires, the reduction is 24–36%.

## History

- The first commit is a snapshot of [Consensys/gnark](https://github.com/Consensys/gnark)
  at `6e6960808`, with upstream history squashed into it.
- All following commits are the original development history, unchanged.
  The two `Merge upstream gnark … (squashed)` commits bring in upstream changes
  (`8e4363f1a` and `c05c6a810`), again squashed into one commit each.

The Go module path is still `github.com/consensys/gnark`, so the packages build
as in upstream.

## License

Apache 2.0, as upstream `gnark` (see [LICENSE](LICENSE)).

---

# `gnark` zk-SNARK library

[![Twitter URL](https://img.shields.io/twitter/url/https/twitter.com/gnark_team.svg?style=social&label=Follow%20%40gnark_team)](https://twitter.com/gnark_team) [![License](https://img.shields.io/badge/license-Apache%202-blue)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/consensys/gnark)](https://goreportcard.com/badge/github.com/consensys/gnark)
[![PkgGoDev](https://pkg.go.dev/badge/mod/github.com/consensys/gnark)](https://pkg.go.dev/mod/github.com/consensys/gnark)
[![Documentation Status](https://readthedocs.com/projects/pegasys-gnark/badge/)][`gnark` User Documentation] [![DOI](https://zenodo.org/badge/DOI/10.5281/zenodo.5819104.svg)](https://doi.org/10.5281/zenodo.5819104)

High-performance zk-SNARKs in Go.

`gnark` provides a high-level API to define circuits, then compile, prove, and verify with production-grade proving systems. It is open-source under Apache 2.0 and uses [`gnark-crypto`] for field arithmetic and cryptographic primitives.

`gnark` powers [`Linea zk-rollup`](https://linea.build). Include your project in [known users](docs/KNOWN_USERS.md) by opening a PR.

## Why `gnark`

- Circuit development in idiomatic Go
- Fast proving and verification backends
- Reusable standard gadgets in `std/`
- Active security and regression testing culture

## Useful Links

- [`gnark` User Documentation]
- [`gnark` Playground]
- [`gnark` Issues]
- [`gnark` Benchmarks](https://docs.gnark.consensys.net/overview#gnark-is-fast) 🏁
- [`gnark-announce`] - release and security announcements

## Quick Start

### Requirements

- Go `1.25+` (module target: `go 1.25.7`)

### Install

```bash
go get github.com/consensys/gnark@latest
```

### Run an example

```bash
go run ./examples/cubic
```

To design your first circuit, follow the tutorial in [`gnark` User Documentation].

## Supported Proving Systems and Curves

`gnark` currently supports:

- Groth16
- PLONK

on the following curves:

- BN254
- BLS12-381
- BLS12-377
- BW6-761

Notes:

- Solidity verifier export support is curve-dependent (BN254 is the primary target).
- Serialized formats are not guaranteed to be stable across versions.

## GPU Acceleration (Experimental)

`gnark` includes experimental GPU acceleration through Ingonyama's ICICLE backend for Groth16 on:

- BN254
- BLS12-377
- BLS12-381
- BW6-761

See [accelerated backend documentation](backend/accelerated/icicle/doc.go) and the [ICICLE repository](https://github.com/ingonyama-zk/icicle-gnark).

## Example Circuit

The circuit below encodes `x**3 + x + 5 == y`.

```go
package main

import (
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
)

// CubicCircuit defines a simple circuit.
// x**3 + x + 5 == y
type CubicCircuit struct {
	X frontend.Variable `gnark:"x"`
	Y frontend.Variable `gnark:",public"`
}

// Define declares the circuit constraints.
func (circuit *CubicCircuit) Define(api frontend.API) error {
	x3 := api.Mul(circuit.X, circuit.X, circuit.X)
	api.AssertIsEqual(circuit.Y, api.Add(x3, circuit.X, 5))
	return nil
}

func main() {
	var circuit CubicCircuit
	ccs, _ := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)

	pk, vk, _ := groth16.Setup(ccs)

	assignment := CubicCircuit{X: 3, Y: 35}
	witness, _ := frontend.NewWitness(&assignment, ecc.BN254.ScalarField())
	publicWitness, _ := witness.Public()

	proof, _ := groth16.Prove(ccs, pk, witness)
	_ = groth16.Verify(proof, vk, publicWitness)
}
```

## Security

**`gnark` and [`gnark-crypto`] have been [extensively audited](#audits), but are provided as-is with no guarantees or warranties. In particular, `gnark` does not guarantee constant-time implementations or side-channel resistance.**

Report vulnerabilities via [Security Policy](SECURITY.md). Do **not** open public issues for security reports.

Published advisories are listed [here](https://github.com/Consensys/gnark/security/advisories?state=published).

## Testing

CI runs formatting, generated-file, lint, and test checks on pull requests and pushes.

Common local commands:

```bash
go test -short ./...
go test -tags=release_checks,solccheck .
go test -tags=prover_checks ./test/... ./examples/...
go test -run=NONE -fuzz=FuzzIntcomp -fuzztime=30s ./internal/backend/ioutils
go generate ./...
```

## Audits

- [Kudelski Security - October 2022 - gnark-crypto (contracted by Algorand Foundation)](audits/2022-10%20-%20Kudelski%20-%20gnark-crypto.pdf)
- [Sigma Prime - May 2024 - gnark-crypto KZG (contracted by Ethereum Foundation)](audits/2024-05%20-%20Sigma%20Prime%20-%20kzg.pdf)
- [Consensys Diligence - June 2023 - gnark PLONK Solidity verifier](https://consensys.io/diligence/audits/2023/06/linea-plonk-verifier/)
- [LeastAuthority - August 2023 - gnark Groth16 Solidity verifier template (contracted by Worldcoin)](https://leastauthority.com/wp-content/uploads/2023/08/Worldcoin_Groth16_Verifier_in_EVM_Smart_Contract_Final_Audit_Report.pdf)
- [OpenZeppelin - November 2023 - gnark PLONK Solidity verifier template](https://blog.openzeppelin.com/linea-verifier-audit-1)
- [ZKSecurity.xyz - May 2024 - gnark standard library](audits/2024-05%20-%20zksecurity%20-%20gnark%20std.pdf)
- [OpenZeppelin - June 2024 - gnark PLONK prover and verifier](https://blog.openzeppelin.com/linea-prover-audit)
- [LeastAuthority - September 2024 - gnark general and GKR](audits/2024-09%20-%20Least%20Authority%20-%20arithm%20and%20GKR.pdf)
- [LeastAuthority - November 2024 - Linea zkEVM](audits/2024-11%20-%20Least%20Authority%20-%20Linea%20zkEVM.pdf)

## Release Notes

See [CHANGELOG.md](CHANGELOG.md).

## Citing

If you use `gnark` in research, please cite the latest release:

```bib
@software{gnark-v0.15.0,
  author       = {Gautam Botrel and
                  Thomas Piellard and
                  Youssef El Housni and
                  Ivo Kubjas and
                  Arya Tabaie and
                  Yao J. Galteland},
  title        = {Consensys/gnark: v0.15.0},
  month        = may,
  year         = 2026,
  publisher    = {Zenodo},
  version      = {v0.15.0},
  doi          = {10.5281/zenodo.5819104},
  url          = {https://doi.org/10.5281/zenodo.5819104}
}
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) and [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).

## Versioning

`gnark` follows [SemVer](http://semver.org/). Available versions are in [tags](https://github.com/Consensys/gnark/tags).

## License

Licensed under Apache 2.0 (see [LICENSE](LICENSE)).

[`gnark` Issues]: https://github.com/Consensys/gnark/issues
[`gnark` Playground]: https://play.gnark.io
[`gnark` User Documentation]: https://docs.gnark.consensys.net/
[`gnark-announce`]: https://groups.google.com/g/gnark-announce
[`gnark-crypto`]: https://github.com/Consensys/gnark-crypto
