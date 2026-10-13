# Poker Machine

A five-reel slot machine in Go, using [Ebitengine](https://ebitengine.org/).
It is a rewrite of `SLOT.PAS`, a Turbo Pascal program, which is kept here
for reference. It plays in the style of modern hold-and-respin poker
machines.

This is a hobby project. It is played with pretend credits only: no real
money goes in or comes out, and the dollar amounts on screen are part of the
pretence. It is not affiliated with or endorsed by any maker of gaming
machines.

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
lines. AUTO SPIN keeps spinning at that bet until you press it again, a
feature is won or the credits run short.

| Buttons | Choices |
|---|---|
| Bet per line (top row) | 1, 2, 5, 10 or 20 credits |
| Lines (bottom row) | 1, 5, 10, 25 or 50 lines |

The keyboard works too:

| Key | Action |
|---|---|
| Space or Enter | Spin |
| Up / Down | Change the bet per line |
| Left / Right | Change the number of lines |
| M | Play all 50 lines at the highest bet the credits cover |
| A | Auto spin on or off |
| C | Insert a coin: 10,000 more credits ($100) |
| R | Reset your credits to 10,000 ($100) |
| S | Sound on or off |
| Esc | Back to the menu |

You start with 10,000 credits ($100). The INSERT COIN button adds 10,000 more
whenever you want them, and RESET puts them back to 10,000. Each payline is a path across
the five reels.
Three or more of a symbol in a row along a line, starting from the left
reel, win; four or five in a row pay more. Two dollars in a row pay too.
A win on a line is multiplied by the bet per line, and the wins on all the
lines played are added together.

There are 50 paylines, listed in `machine/machine.go`. Choosing
a number of lines shows them over the reels, and after a win each winning
line is shown in turn.

### Symbols

| Symbol | What it does |
|---|---|
| 10, J, Q, K, A | Low-paying line symbols |
| Cherry, plum, bell, dollar | Higher-paying line symbols |
| Joker | Wild: stands in for any of the symbols above and doubles the win. Reels 2 to 5 only |
| Dice | Scatter: three or more anywhere pay a prize and win free games |
| Poker chip | Six or more anywhere start the link feature |

### Free games

Three, four or five dice anywhere on the reels pay 2, 15 or 100 times the
total bet and win 6 free games. Free games cost nothing and play by
themselves at the bet that won them.

In a free game reels 2, 3 and 4 turn as one and show a single giant symbol,
which counts as that symbol in each of its nine places. The middle reels
always match, so a line wins when reel 1 shows the same symbol, and only
four or five in a row pay:

| Symbol | Four | Five |
|---|---|---|
| 10, J, Q, K | 5 | 15 |
| Ace | 5 | 20 |
| Cherry, plum | 10 | 25 |
| Bell | 10 | 50 |
| Dollar | 10 | 100 |

Three of the giant symbols are special:

| Giant symbol | Chance each game | What it does |
|---|---|---|
| Joker | 3% | Wild: every line pays for the symbol on reel 1. It does not double a win in a free game |
| Chip | 2.5% | Starts the link feature with its nine places held. It carries one prize: 5 to 100 times the total bet, or MINI, MINOR or MAJOR |
| Dice | 6% | Pays the total bet and adds 3 free games |

Reels 1 and 5 have no dice or chips in a free game, and the joker is on
reel 5 only.

### Link feature

The chips sit in stacks of three on the reels, and each one that lands
shows what it carries. Six or more anywhere lock in place and start the
link feature:

1. Everything else clears and you get three respins.
2. On each respin every empty position has the same chance of landing a
   chip, which locks too and puts the respins back to three.
3. When the respins run out, or all fifteen positions are full, every chip
   is paid.

A giant chip in a free game starts it as well, with nine positions already
held; the free games carry on afterwards.

| A chip can carry | Worth |
|---|---|
| Credits | 1 to 100 times the total bet |
| MINI | 20 times the total bet |
| MINOR | 200 times the total bet |
| MAJOR | The major pot |

The GRAND is not on a chip: it is paid for filling all fifteen positions.

The four jackpots are shown across the top in dollars, at 100 credits to
the dollar. MINI and MINOR follow the bet. MAJOR and GRAND are pots that
start at $5,000 and $20,000 and grow with every paid spin, by 1% and 0.5%
of the bet; they pay the same whatever the bet, and go back to their
starting amounts when won. To keep small bets from having the same shot at
them as large ones, a MAJOR chip and the chip that would complete the grid
both land more often the more you bet. Even at the smallest bet the GRAND
comes up at least once in a hundred million spins.

### Saved game

The game is saved after every spin, so you can close it at any time and
carry on later with the same credits, bet and jackpot pots, and with any
free games or link feature still to play. RESET only ever touches the
credits: the MAJOR and GRAND pots keep growing until someone wins them.

The save is one small file, `PokerMachine/save.json`, in your settings
folder: `%AppData%` on Windows and `~/.config` on Linux. Delete it to start
again from scratch, pots included.

## Layout

- `machine/` holds the rules: reel strips, paylines, paytable, free games, the
  link feature, credits and the odds.
  It has no graphics and is covered by tests.
- `main.go` has the screens, buttons, input and reel animation.
- `linkview.go` shows the link feature and the jackpot meters.
- `save.go` keeps the game between sessions.
- `symbols.go` draws the symbols from simple shapes.
- `draw.go` has the palette and drawing helpers.
- `sound.go` generates the sound effects.

## Odds

The odds come from the reel strips and rules in `machine/`. At a bet of 50
credits the machine pays back about 92% of what is bet over the long run:

| From | Share of what is bet |
|---|---|
| Line wins | 44% |
| Dice prizes and free games | 17% |
| Link chips, with MINI and MINOR | 27% |
| MAJOR and GRAND pots | 4% |

Free games are won about once in 68 spins, and about seven are played each
time. The link feature comes up about once in 82 spins, counting those a
giant chip starts. The return is the same at every bet apart from the
GRAND, which lifts it to about 94% at the smallest bet. Run the tests to
recalculate all of this after changing the rules:

```
go test -v ./machine/
```

Every spin is drawn from a cryptographic random number generator seeded by
the operating system, and seeded afresh every hundred spins. No spin depends
on the ones before it.

## Licence

Released under the MIT licence. See `LICENSE`.
