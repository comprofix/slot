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

Choose a bet per line and a number of lines with the buttons under the
reels, then press SPIN. A spin costs the bet per line times the number of
lines.

| Buttons | Choices |
|---|---|
| Bet per line (top row) | 1, 2, 5, 10 or 20 credits |
| Lines (bottom row) | 1, 5, 10, 15 or 25 lines |

The keyboard works too:

| Key | Action |
|---|---|
| Space or Enter | Spin |
| Up / Down | Change the bet per line |
| Left / Right | Change the number of lines |
| M | Play all 25 lines at the highest bet the credits cover |
| S | Sound on or off |
| Esc | Back to the menu |

You start with 1000 credits. Each payline is a path across the five reels.
Three or more of a symbol in a row along a line, starting from the left
reel, win; four or five in a row pay much more. Cherries also pay for two.
A line's win is multiplied by the bet per line, and the wins on all the
lines played are added together.

There are 25 paylines, listed in `machine/machine.go`. Choosing
a number of lines shows them over the reels, and after a win each winning
line is shown in turn.

## Layout

- `machine/` holds the rules: reel strips, paylines, paytable, spinning and
  credits.
  It has no graphics and is covered by tests.
- `main.go` has the screens, buttons, input and reel animation.
- `symbols.go` draws the six symbols from simple shapes.
- `draw.go` has the palette and drawing helpers.
- `sound.go` generates the sound effects.

## Odds

The odds come from the reel strips and paytable in `machine/machine.go`.
The machine pays back about 72% of what is bet over the long run, however
many lines are played. Run the tests
to recalculate it after changing either:

```
go test -v ./machine/
```
