# Ashta Changa - Traditional Indian Board Game

A web-based implementation of **Ashta Changa** (also known as Ashta Chamma), a traditional Indian board game similar to Ludo. This game is built using Go and the Ebiten game library, compiled to WebAssembly for browser-based play.

## About the Game

Ashta Changa is an ancient Indian race game that has been played for centuries. It is the precursor to modern Ludo and Pachisi. The game is traditionally played on a 5x5 board with cowrie shells as dice.

### Game Rules

1. **Players**: 2-4 players can participate, each with 4 tokens
2. **Board**: A 5x5 grid with a central goal area
3. **Dice**: 4 cowrie shells are used, resulting in values 1-8
4. **Starting**: Players must roll a 1, 4, or 8 to move a token out of their starting position
5. **Movement**: Tokens move anti-clockwise around the outer edge, then enter the inner path to reach the center
6. **Safe Zones**: Only the starting points (marked with X) are safe zones - tokens cannot be killed there
7. **Killing**: Landing on an opponent's token sends it back to their home position (not possible on safe zones)
8. **Bonus Turns**: Rolling a 4, 8, or killing an opponent grants an extra turn
9. **Winning**: First player to get all 4 tokens to the center wins

### Board Layout

```
+---+---+---+---+---+
|   |   | X |   |   |  <- X marks starting points (safe zones)
+---+---+---+---+---+
|   |   | I |   |   |  <- I marks inner paths (not safe)
+---+---+---+---+---+
| X | I | G | I | X |  <- G is the goal (center)
+---+---+---+---+---+
|   |   | I |   |   |
+---+---+---+---+---+
|   |   | X |   |   |
+---+---+---+---+---+
```

- **X**: Starting points (one for each player) - ONLY safe zones on the board
- **I**: Inner paths leading to center (not safe - tokens can be killed here)
- **G**: Goal (golden center block)
- Outer edge: Movement path

## Technology Stack

- **Language**: Go 1.25
- **Game Engine**: Ebiten v2
- **Platform**: WebAssembly for browser deployment

## Project Structure

```
.
├── main.go          # Game logic and board rendering
├── go.mod           # Go module definition
├── go.sum           # Dependency checksums
├── web/
│   ├── index.html   # HTML page for WebAssembly
│   ├── game.wasm    # Compiled WebAssembly (generated)
│   └── wasm_exec.js # Go WebAssembly runtime
├── wasm_exec.js     # Go WebAssembly runtime (source)
├── .gitignore       # Git ignore rules
└── README.md        # This file
```

## Building and Running

### Prerequisites

- Go 1.25 or later
- A web browser with WebAssembly support

### Build WebAssembly

To build the WebAssembly binary:

```bash
GOOS=js GOARCH=wasm go build -o web/game.wasm main.go
```

### Run Locally

To serve the game locally, use the `serve` package:

```bash
cd web
serve .
```

Then open your browser to `http://localhost:3000` (or the port shown).

## Development Status

This project is under active development. Current features:

- [x] Board rendering with traditional styling
- [x] Token placement and movement
- [x] Cowrie shell dice rolling
- [x] Turn-based gameplay with 4 players
- [x] Token killing mechanics
- [x] Safe zones (starting points only)
- [x] Extra turn on 4, 8, or killing
- [x] Win condition detection

## License

MIT License

## Acknowledgments

- Traditional Indian board game heritage
- Ebiten game library for Go
- The ancient game designers of India
