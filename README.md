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
6. **Killing**: Landing on an opponent's token sends it back to their starting position
7. **Bonus Turns**: Rolling a 4, 8, or killing an opponent grants an extra turn
8. **Winning**: First player to get all 4 tokens to the center wins

### Board Layout

```
+---+---+---+---+---+
|   |   | X |   |   |  <- X marks starting points
+---+---+---+---+---+
|   |   | S |   |   |  <- S marks safe zones (home columns)
+---+---+---+---+---+
| X | S | G | S | X |  <- G is the goal (center)
+---+---+---+---+---+
|   |   | S |   |   |
+---+---+---+---+---+
|   |   | X |   |   |
+---+---+---+---+---+
```

- **X**: Starting points (one for each player)
- **S**: Safe zones (player-specific paths to center)
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
- [ ] Token placement and movement
- [ ] Cowrie shell dice rolling
- [ ] Turn-based gameplay
- [ ] Token killing mechanics
- [ ] Win condition detection
- [ ] Multiplayer support

## License

MIT License

## Acknowledgments

- Traditional Indian board game heritage
- Ebiten game library for Go
- The ancient game designers of India
