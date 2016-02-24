# zurichess: a chess engine

[Website](https://bitbucket.org/zurichess/zurichess) |
[CCRL](http://www.computerchess.org.uk/ccrl/404/cgi/engine_details.cgi?print=Details+%28text%29&eng=Zurichess%20Appenzeller%2064-bit) |
[Wiki](http://chessprogramming.wikispaces.com/Zurichess) |
[![Reference](https://godoc.org/bitbucket.org/zurichess/zurichess?status.svg)](https://godoc.org/bitbucket.org/zurichess/zurichess)
[![Build Status](https://drone.io/bitbucket.org/zurichess/zurichess/status.png)](https://drone.io/bitbucket.org/zurichess/zurichess/latest)

zurichess is a chess engine and a chess library written in
[Go](http://www.golang.org). Its main goals are: to be a relatively
strong chess engine and to enable chess tools writing. See
the library reference.

zurichess is NOT a complete chess program. Like with most
other chess engines you need a GUI that supports the UCI
protocol. Some popular GUIs are XBoard (Linux), Eboard (Linux)
Winboard (Windows), Arena (Windows).

zurichess partially implements [UCI
protocol](http://wbec-ridderkerk.nl/html/UCIProtocol.html), but
the available commands are enough for most purposes. zurichess was
successfully tested under Linux AMD64 and Linux ARM and other people
have tested zurichess under Windows AMD64.

zurichess plays on [FICS](http://freechess.org) under the handle
[zurichess](http://ficsgames.org/cgi-bin/search.cgi?player=zurichess&action=Statistics).
Usually it runs code at tip (master) which is a bit stronger
than the latest stable version.

## Build and Compile

First you need to get the latest version of Go (currently 1.5.2). For
instructions how to download and install Go for your OS see
[documentation](https://golang.org/doc/install). After the Go compiler
is installed, create a workspace:

```
#!bash
$ mkdir gows ; cd gows
$ export GOPATH=`pwd`
```

After the workspace is created cloning and compiling zurichess is easy:

```
#!bash
$ go get -u bitbucket.org/zurichess/zurichess/zurichess
$ $GOPATH/bin/zurichess --version
zurichess (devel), build with go1.5 at (just now), running on amd64
```

## Download

Precompiled binaries for several platforms and architectures can be found
on the [downloads](https://bitbucket.org/zurichess/zurichess/downloads)
page.

Latest Linux AMD64 binaries can be downloaded from
[drone.io](https://drone.io/bitbucket.org/zurichess/zurichess/files). These
binaries should be stable for any kind of testing.


## Testing

[zuritest](https://bitbucket.org/zurichess/zuritest) is the framework used to test zurichess.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## History

Versions are named after [Swiss Cantons](http://en.wikipedia.org/wiki/Cantons_of_Switzerland)
in alphabetical order.

### zurichess - [glarus](https://en.wikipedia.org/wiki/Canton_of_Glarus) (development)

* Improve futility conditions. Geneva's futility is a bit too agressive
and causes lots of tactical mistakes.
* Add History Leaf Pruning similar to https://chessprogramming.wikispaces.com/History+Leaf+Pruning.
* Improve pawn evaluation caching. Also cache shelter evaluation.
* Usual code clean ups, speed ups and bug fixes.

### zurichess - [geneva](https://en.wikipedia.org/wiki/Canton_of_Geneva) (stable)

The theme of this release is improving evaluation.

* Implement fifty-move draw rule. Add HasLegalMoves and InsufficientMaterial methods.
* Improve move ordering: add killer phase; remove sorting.
* Improve time control: add more time when the move is predicted.
* Add basic futility pruning.
* Switch tuning to using [TensorFlow](http://tensorflow.org/) framework. txt is now deprecated.
* Evaluate rooks on open and half-open files.
* Improve mobility calculation.
* Tweak null-move conditions: allow double null-moves.
* Usual code clean ups, speed ups and bug fixes.

### zurichess - [fribourg](https://en.wikipedia.org/wiki/Canton_of_Fribourg)

The theme of this release is tuning the evaluation, search and move generation.
ELO is about 2441 on CCRL 40/40.

* Move to the new page http://bitbucket.org/zurichess/zurichess.
* Evaluate passed, connected and isolated pawns. Tuning was done
using Texel's tuning method implemented by
[txt](https://bitbucket.org/zurichess/txt).
* Add Static Exchange Evaluation (SEE).
* Ignore bad captures (SEE < 0) in quiescence search.
* Late move reduce (LMR) of all quiet non-critical moves. Aggressively reduce
bad quiet (SEE < 0) moves at higher depths.
* Adjust LMR conditions. Reduce more at high depths (near root) and high move count.
* Increase number of killers to 4. Helps with more aggressive LMR.
* Improve move generation speed. Add phased move generation: hash,
captures, and quiets. Phased move generation allows the engine to skip
generation or sorting of the moves in many cases.
* Implement `setoption Clear Hash`.
* Implement pondering. Should give some ELO boost for online competitions.
* Improve move generation order. Picked the best among 20 random orders.
* Prune two-fold repetitions at non-root nodes. This pruning cuts huge parts
of the search tree without affecting search quality. >30ELO improvement
in self play.
* Small time control adjustment. Still too little time used in the mid
game. Abort search if it takes much more time than alloted.
* Usual code clean ups, speed ups and bug fixes.

### zurichess - [bern](http://en.wikipedia.org/wiki/Canton_of_Bern)

This release's theme is pruning the search. ELO is about 2234 on CCRL 40/4.

* Implement Principal Variation Search (PVS).
* Reduce late quiet moves (LMR).
* Optimize move ordering. Penalize moves threatened by pawns in quiescence search.
* Optimize check extension. Do not extend many bad checks.
* Change Zobrist key to be equal to polyglot key. No book support, but better hashing.
* Add some integration tests such as mate in one and mate in two.
* Usual code clean ups, speed ups and bug fixes.

### zurichess - [basel](http://en.wikipedia.org/wiki/Basel-Stadt)

This release's theme is improving evaluation function.

* Speed up move ordering considerably.
* Implement null move pruning.
* Clean up and micro optimize the code.
* Tune check extensions and move ordering.
* Award mobility and add new piece square tables.
* Handle three fold repetition.
* Cache pawn structure evaluation.
* Fix transposition table bug causing a search explosion around mates.
* Prune based on mate score.

### zurichess - [appenzeller](http://en.wikipedia.org/wiki/Appenzeller_cheese)

This release's theme is improving search. ELO is about 1823 on CCRL 40/4.

* Clean code and improved documentation.
* Implement aspiration window search with gradual widening.
* Improve replacement strategy in transposition table.
* Double the number of entries in the transposition table.
* Develop [zuritest](https://bitbucket.org/zurichess/zuritest), testing infrastructure for zurichess.
* Fail-softly in more situations.
* Implement UCI commands `go movetime` and `stop`.
* Add a separate table for principal variation.
* Add killer heuristic to improve move ordering.
* Extend search when current position is in check.
* Improve time-control. In particular zurichess uses more time when there are fewer pieces on the board.

### zurichess - [aargau](http://en.wikipedia.org/wiki/Aargau)

This is the first public release. ELO is about 1727 on CCRL 40/4.

* Core search function is a mini-max with alpha-beta pruning on top of a negamax framework.
* Sliding attacks are implemented using fancy magic bitboards.
* Search is sped up with transposition table with Zobrist hashing.
* Move ordering inside alpha-beta is done using table move & Most Valuable Victim / Least Valuable Victim.
* Quiescence search is used to reduce search instability and horizon effect.
* [Simplified evaluation function](https://chessprogramming.wikispaces.com/Simplified+evaluation+function) with tapered eval.

## External links

A list of zurichess related links:

* [Chess Programming WIKI](http://chessprogramming.wikispaces.com/Zurichess)
* [CCRL 40/4](http://www.computerchess.org.uk/ccrl/404/cgi/engine_details.cgi?print=Details+%28text%29&eng=Zurichess%20Appenzeller%2064-bit)
* [FICS Games](http://ficsgames.org/cgi-bin/search.cgi?player=zurichess&action=Statistics)

Other sites, pages and articles with a lot of useful information:

* [Chess Programming Part V: Advanced Search](http://www.gamedev.net/page/resources/_/technical/artificial-intelligence/chess-programming-part-v-advanced-search-r1197)
* [Chess Programming WIKI](http://chessprogramming.wikispaces.com)
* [Computer Chess Club Forum](http://talkchess.com/forum/index.php)
* [Computer Chess Programming](http://verhelst.home.xs4all.nl/chess/search.html)
* [Computer Chess Programming Theory](http://www.frayn.net/beowulf/theory.html)
* [How Stockfish Works](http://rin.io/chess-engine/)
* [Little Chess Evaluation Compendium](https://chessprogramming.wikispaces.com/file/view/LittleChessEvaluationCompendium.pdf)
* [The effect of hash collisions in a Computer Chess program](https://cis.uab.edu/hyatt/collisions.html)
* [The UCI protocol](http://wbec-ridderkerk.nl/html/UCIProtocol.html)

## Disclaimer

This project is not associated with my employer.
