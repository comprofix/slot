# Poker Machine

A five-reel slot machine in Go, using [Ebitengine](https://ebitengine.org/).
It is a rewrite of `SLOT.PAS`, a Turbo Pascal program, which is kept here
for reference.

## Build and run

You need Go 1.27 or later.

### Windows

No C compiler is needed.

```
go build -ldflags -H=windowsgui -o slot.exe .
.\slot.exe
```

The `-H=windowsgui` flag stops a console window opening behind the game.

### Linux

Ebitengine needs a C compiler and some development libraries. On Debian or
Ubuntu (including WSL):

```
sudo apt install gcc libc6-dev libgl1-mesa-dev libxcursor-dev libxi-dev \
    libxinerama-dev libxrandr-dev libxxf86vm-dev libasound2-dev pkg-config
go build -o slot .
./slot
```

## Playing

| Key | Action |
|---|---|
| Space or Enter | Spin |
| Up / Down | Change the bet (1 to 3 credits) |
| M | Bet the maximum |
| S | Sound on or off |
| Esc | Back to the menu |

You start with 100 credits. Only the middle row counts. Three or more of a
symbol in a row, starting from the left reel, win; four or five in a row pay
much more. Cherries also pay for one or two. Every win is multiplied by the
bet.

## Layout

- `machine/` holds the rules: reel strips, paytable, spinning and credits.
  It has no graphics and is covered by tests.
- `main.go` has the screens, input and reel animation.
- `symbols.go` draws the six symbols from simple shapes.
- `draw.go` has the palette and drawing helpers.
- `sound.go` generates the sound effects.

## Odds

The odds come from the reel strips and paytable in `machine/machine.go`.
The machine pays back about 92% of what is bet over the long run. Run the tests
to recalculate it after changing either:

```
go test -v ./machine/
```
