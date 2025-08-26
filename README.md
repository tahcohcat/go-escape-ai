# go-escape-ai

![Build Passing](https://img.shields.io/badge/build-passing-brightgreen)
![Sanity Failing](https://img.shields.io/badge/sanity-failing-red)
![Escape Attempts](https://img.shields.io/badge/escape%20attempts-404-blue)
![AI Narration](https://img.shields.io/badge/ai--narration-unreliable-yellow)
![Go Powered](https://img.shields.io/badge/made%20with-Go-00ADD8?logo=go)
![Room Status](https://img.shields.io/badge/room-locked-black)
![Puzzle Difficulty](https://img.shields.io/badge/difficulty-impossible-lightgrey)

An AI narrator that locks you in a room and then pretends to help

## Architecture

This escape room game uses a unique architecture where:

1. **LLM Generates Complete Solution**: The AI generates the entire scenario, theme, rooms, puzzles, and solutions once at game start
2. **Game Engine Manages State**: All game state (player location, inventory, puzzle progress) is tracked by the Go engine
3. **LLM Provides Narration**: The AI only narrates the immediate context based on current state, without knowing the solutions

This ensures consistent gameplay while providing dynamic, atmospheric narration.

## Features

- **AI-Generated Scenarios**: Each playthrough can have a unique theme and puzzle set
- **Persistent State**: Game saves scenarios locally for consistent replay
- **Atmospheric Narration**: AI narrator provides immersive context without spoiling solutions  
- **Fallback Mode**: Works without API key using pre-built scenarios
- **Text Adventure Interface**: Classic command-line gameplay

## Installation

```bash
git clone https://github.com/tahcohcat/go-escape-ai
cd go-escape-ai
go mod tidy
go build -o escape-ai
```

## Setup

1. (Optional) Get an OpenAI API key from https://platform.openai.com/api-keys
2. Set environment variable: `export OPENAI_API_KEY="your-key-here"`
3. Run the game: `./escape-ai`

Without an API key, the game uses fallback scenarios and basic narration.

## Commands

- `look [item]` - Examine surroundings or specific item
- `take <item>` - Pick up an item  
- `use <item>` - Use an item from inventory
- `go <direction>` - Move to different room
- `inventory` - Check what you're carrying
- `solve <answer>` - Attempt to solve a puzzle
- `hint` - Get a hint for current room
- `stats` - Show game statistics
- `help` - Show command help
- `quit` - Exit game

## Example Gameplay

```
🔒 Welcome to Go Escape AI 🔒
Creating new escape room scenario...
Enter a theme (or press Enter for random): Haunted Victorian Mansion

📍 Theme: Haunted Victorian Mansion
🏛️ Setting: A decrepit mansion on a stormy night
📖 Backstory: You are a paranormal investigator trapped in a mansion...

🚪 You are in the Grand Foyer. Dust motes dance in the moonlight...

> look around
🤖 The shadows seem to shift as you examine the room. Something glints near the old grandfather clock...

> take key
📝 You take the ornate key.

> inventory  
📝 You have: ornate key

> solve midnight
🤖 The clock chimes ominously as the correct time reveals a hidden passage...
```

## Game Architecture

### Core Components

- **`game/scenario.go`**: Defines the complete game world structure
- **`game/engine.go`**: Manages game state and command processing  
- **`llm/client.go`**: Handles AI generation and narration
- **`main.go`**: Game loop and CLI interface

### Data Flow

1. Game starts → LLM generates complete scenario → Saved locally
2. Player commands → Engine updates state → LLM narrates context
3. Engine maintains authoritative state, LLM only provides flavor text
