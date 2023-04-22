# Varnam Desktop App

Cross-platform desktop app for Varnam. Uses [varnam-editor](https://github.com/varnamproject/editor) as GUI frontend.

## Installation

- Download from [Releases](https://github.com/varnamproject/desktop/releases)

## Development

- Clone the repo
- Fetch submodules:

```bash
git submodule update --init --recursive
```

- Build govarnam:

```bash
make govarnam
```

- Build the main desktop app :

```bash
make deps
make ui
make build
```

- Run :

```bash
./varnam
```

Thanks to [stuffbin](https://github.com/knadh/stuffbin), HTML+CSS+JS files are all merged into one big binary (`varnam`).

## Usage

`varnam` bundles with `varnamd`, the HTTP server to interface with varnam library. By default, the server runs on `127.0.0.1:8123`. See [varnamd](https://github.com/varnamproject/varnamd) for more details.
