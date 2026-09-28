# zuginbox

Local batch audit for German e-invoices (**XRechnung** / **ZUGFeRD**). Offline-first: lightweight XML checks run without Java; optional [Mustangproject](https://github.com/ZUGFeRD/mustangproject) JAR enables PDF embedded-XML validation.

## Install

```bash
go install github.com/sysrqio/zuginbox/cmd/zuginbox@latest
# or
make build && ./bin/zuginbox version
```

## Usage

```bash
# Audit a folder of .xml / .pdf files
zuginbox audit --input ./invoices

# JSON stdout + reports on disk (default paths)
zuginbox audit --input invoice.xml --format json --output receipt-audit.json --report-md receipt-audit.md

# Treat warnings (e.g. PDF without Mustang) as failure
zuginbox audit --input ./invoices --fail-on-warn

# Demo bundled fixtures (no Java)
zuginbox audit --offline-fixture

# Optional full PDF/ZUGFeRD validation
zuginbox audit --input ./invoices --mustang-jar ./mustang-validator.jar

zuginbox doctor
zuginbox doctor --mustang-jar ./mustang-validator.jar
zuginbox version
```

### Exit codes

| Code | Meaning |
|-----:|---------|
| 0 | All receipts valid |
| 1 | I/O error or Java/Mustang failure |
| 2 | Invalid receipt, or warn when `--fail-on-warn` |

## Offline vs Mustang

- **XML:** Built-in checks for BT-1 (invoice number), BT-24 (specification ID), and document totals. No network calls during `audit`.
- **PDF:** Without `--mustang-jar`, PDFs are listed with `validation=warn` and a note to install Java + Mustang. Use `scripts/fetch-mustang.sh` to download the JAR (not vendored in git).

## Development

```bash
go test ./...
make build
```

## Docker

Image includes **Eclipse Temurin 21 JRE** for optional Mustang use:

```bash
docker build -t zuginbox .
docker run --rm zuginbox version
```

## License

Apache-2.0 — see [LICENSE](LICENSE).
