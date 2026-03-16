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
├── main.go                    # Minimal entry point (delegates to gui package)
├── go.mod                     # Go module definition
├── go.sum                     # Dependency checksums
├── gui/                       # GUI package (Ebiten-based rendering)
│   ├── game.go                # Game struct and Ebiten interface implementation
│   ├── render.go              # All drawing/rendering functions
│   ├── update.go              # Input handling and game state updates
│   ├── animation.go           # Animation state management (AnimationManager)
│   └── interfaces.go          # Renderer interface for testing/mocking
├── game/                      # Core game logic package (shared between client/server)
│   ├── types.go               # Token, Player, Board, ConchShells, constants
│   ├── logic.go               # Game rules, movement, killing, winning logic
│   └── interfaces.go          # GameLogic interface for testing/mocking
├── network/                   # Network client package
│   ├── client.go              # NetworkClient struct and WebSocket methods
│   ├── types.go               # Network types (NetworkRoom, NetworkPlayer, etc.)
│   └── interfaces.go          # NetworkClientInterface for testing/mocking
├── server/                    # Backend server for multiplayer
│   ├── models.go              # All model types (Room, Player, GameState, etc.)
│   ├── constants.go           # All constants (WebSocket, game, HTTP)
│   ├── interfaces.go          # Server interfaces (RoomManager, Hub, etc.)
│   ├── room.go                # Room management and game state
│   ├── websocket.go           # WebSocket client handling
│   ├── message.go             # Message types and handlers
│   ├── server.go              # HTTP server setup
│   ├── game_logic_test.go     # Tests for server logic
│   └── cmd/
│       └── main.go            # Server entry point
├── web/                       # UNCHANGED - WebAssembly assets
│   ├── index.html             # HTML page for WebAssembly
│   ├── game.wasm              # Compiled WebAssembly (generated)
│   └── wasm_exec.js           # Go WebAssembly runtime
├── .gitignore                 # Git ignore rules
├── README.md                  # This file
└── game_rules_test.go         # Tests for game rules (uses game package)
```

### Package Descriptions

- **game/**: Core game logic package containing board rules, token movement, killing mechanics, and win conditions. This package is shared between the GUI client and the backend server to ensure consistent game behavior. Includes `GameLogic` interface for mocking in tests.

- **gui/**: Ebiten-based GUI package handling all rendering, input processing, and game state management. Implements the `ebiten.Game` interface. Includes `Renderer` interface for testing rendering logic and `AnimationManager` for handling animations.

- **network/**: Network client package for WebSocket communication with the backend server. Contains all multiplayer networking code. Includes `NetworkClientInterface` for mocking network calls in tests.

- **server/**: Backend HTTP/WebSocket server for multiplayer gameplay. Handles room management, player connections, and game state synchronization. Uses interfaces (`RoomManagerInterface`, `HubInterface`, etc.) for better testability and abstraction.

## Building and Running

### Prerequisites

- Go 1.25 or later
- A web browser with WebAssembly support

### Running the Backend Server

The multiplayer functionality requires a backend server to handle WebSocket connections and room management.

#### Start the Server

```bash
go run server/cmd/main.go
```

Or with a custom port:

```bash
go run server/cmd/main.go -port=8080
```

The server will start on port 8080 by default and expose:
- WebSocket endpoint: `ws://localhost:8080/ws`
- REST API: `http://localhost:8080/api/rooms`
- Health check: `http://localhost:8080/health`

#### Build the Server Binary

```bash
go build -o ashta-server server/cmd/main.go
```

Then run:

```bash
./ashta-server -port=8080
```

### Build WebAssembly (Game Client)

To build the WebAssembly binary for browser play:

```bash
GOOS=js GOARCH=wasm go build -o web/game.wasm main.go
```

### Run the Game Locally (Desktop)

To run the game as a desktop application:

```bash
go run main.go
```

### Run the Game in Browser

First, build the WebAssembly binary and start the server:

```bash
# Build WASM
GOOS=js GOARCH=wasm go build -o web/game.wasm main.go

# Start the backend server (in another terminal)
go run server/cmd/main.go
```

Then serve the `web` directory using any static file server:

```bash
cd web
# Using Python 3
python3 -m http.server 3000

# Or using Node.js serve package
npx serve .
```

Open your browser to `http://localhost:3000` (or the port shown).

### Full Development Setup

For full multiplayer development, run these commands in separate terminals:

**Terminal 1 - Backend Server:**
```bash
go run server/cmd/main.go
```

**Terminal 2 - Build and Watch (optional):**
```bash
# Build initial WASM
GOOS=js GOARCH=wasm go build -o web/game.wasm main.go

# Watch for changes and rebuild (using watchexec or similar)
watchexec -e go "GOOS=js GOARCH=wasm go build -o web/game.wasm main.go"
```

**Terminal 3 - Static File Server:**
```bash
cd web
python3 -m http.server 3000
```

Then open `http://localhost:3000` in your browser.

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
- [x] **Multiplayer support with WebSocket server**
- [x] **Room creation and joining**
- [x] **Player name customization**
- [x] **Shareable room links**
- [x] **Real-time game state synchronization**

## Multiplayer Guide

### How to Play Online

1. **Start the Server**: Run the backend server first:
   ```bash
   go run server/cmd/main.go
   ```

2. **Open the Game**: Build and serve the game:
   ```bash
   GOOS=js GOARCH=wasm go build -o web/game.wasm main.go
   cd web && python3 -m http.server 3000
   ```

3. **Create a Room**:
   - Open `http://localhost:3000` in your browser
   - Click "Multiplayer"
   - Enter your name
   - Click "Create Room"
   - Share the room link with friends (e.g., `http://localhost:3000?room=ABC123`)

4. **Join a Room**:
   - Open the shared link, or
   - Click "Multiplayer" → Enter your name → Enter Room ID → Click "Join Room"

5. **Start Playing**:
   - The host clicks "Start Game" when all players have joined
   - Play in real-time with synchronized game state

### Room Features

- **Room ID**: A unique 6-character code (e.g., `ABC123`)
- **Shareable Link**: Players can join directly via URL (`?room=ABC123`)
- **Host Controls**: Only the host can start the game
- **Player Names**: Each player can set their custom name
- **Max Players**: Up to 4 players per room

## Code Organization & Refactoring

This project has been refactored to improve code organization, testability, and maintainability.

### Refactoring Changes

1. **Clear Package Structure**: Separated code into distinct packages:
   - `game/` - Core game logic (shared between client and server)
   - `gui/` - Ebiten GUI code (rendering, input handling)
   - `network/` - WebSocket client code
   - `server/` - Backend server (unchanged structure)

2. **Broken into Multiple Files**: Large monolithic files split into focused modules:
   - `gui/game.go` - Game struct and Ebiten interface
   - `gui/render.go` - All drawing functions
   - `gui/update.go` - All input/update handling
   - `game/types.go` - Types and constants
   - `game/logic.go` - Game rules and mechanics
   - `network/client.go` - Network client methods
   - `network/types.go` - Network types

3. **Shared Game Logic**: The `game/` package is now used by both the GUI client and backend server, eliminating code duplication for:
   - Roll calculation (`RollFromShellStates`)
   - Board creation (`NewBoard`)
   - Token movement validation (`CanMoveToken`)
   - Kill checking (`CheckKill`)
   - Win condition (`CheckWin`)

4. **Minimal Entry Point**: `main.go` is now minimal (~20 lines) and delegates to the `gui` package.

### Future Improvements (TODO)

The following improvements can be taken up in subsequent development:

- [x] **Add GameLogic Interface**: Create `GameLogic` interface in `game/` package to allow for mocking in tests
- [x] **Add Renderer Interface**: Create `Renderer` interface in `gui/` package for testing rendering logic
- [x] **Add NetworkClient Interface**: Create `NetworkClient` interface in `network/` package for mocking network calls
- [x] **Extract Animation Logic**: Move animation state management to a separate `animation.go` file
- [ ] **Add Board Interface**: Create `Board` interface for different board configurations
- [ ] **Improve Error Handling**: Add structured error types instead of string messages
- [ ] **Add Game State History**: Implement undo/redo functionality for game states
- [x] **Extract Constants**: Move all magic numbers to a `constants.go` file (completed in server/)
- [ ] **Add Logging Package**: Create a centralized logging package
- [ ] **Add Configuration**: Support configuration via environment variables or config file
- [ ] **Improve Test Coverage**: Add more unit tests for edge cases in game logic
- [ ] **Add Benchmarks**: Add benchmark tests for performance-critical paths
- [ ] **Document APIs**: Add godoc comments to all exported functions

## License

MIT License

## Acknowledgments

- Traditional Indian board game heritage
- Ebiten game library for Go
- The ancient game designers of India
