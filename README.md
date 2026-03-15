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
├── main.go              # Game logic and board rendering with multiplayer support
├── go.mod               # Go module definition
├── go.sum               # Dependency checksums
├── server/              # Backend server for multiplayer
│   ├── room.go          # Room management and game state
│   ├── websocket.go     # WebSocket client handling
│   ├── message.go       # Message types and handlers
│   ├── server.go        # HTTP server setup
│   └── cmd/
│       └── main.go      # Server entry point
├── web/
│   ├── index.html       # HTML page for WebAssembly
│   ├── game.wasm        # Compiled WebAssembly (generated)
│   └── wasm_exec.js     # Go WebAssembly runtime
├── wasm_exec.js         # Go WebAssembly runtime (source)
├── .gitignore           # Git ignore rules
└── README.md            # This file
```

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

## License

MIT License

## Acknowledgments

- Traditional Indian board game heritage
- Ebiten game library for Go
- The ancient game designers of India
