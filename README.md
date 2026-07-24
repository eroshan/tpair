# tpair

A tiny terminal typing trainer for practicing random two-letter combinations.

## Features
- Random pairs from `sadjklewcmpgh`
- Instant key-by-key input, no Enter required
- Live accuracy and WPM stats
- Final session summary on `Ctrl+C`

## Run
```bash
go run .
```

## Options
```bash
go run . --help
go run . stats
```

## Build
```bash
go build -o tpair .
./tpair
```

Note: raw terminal input is implemented with `syscall`, so this is intended for Unix-like terminals.
