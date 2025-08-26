package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tahcohcat/go-escape-ai/game"
	"github.com/tahcohcat/go-escape-ai/llm"
)

const (
	SaveDir = ".escape-ai"
	SaveFile = "scenario.json"
)

func main() {
	fmt.Println("🔒 Welcome to Go Escape AI 🔒")
	fmt.Println("An AI-narrated escape room game")
	fmt.Println()

	llmClient := llm.NewClient()
	
	scenario, err := loadOrCreateScenario(llmClient)
	if err != nil {
		fmt.Printf("Error setting up game: %v\n", err)
		return
	}

	engine := game.NewEngine(scenario)
	
	fmt.Printf("📍 Theme: %s\n", scenario.Theme)
	fmt.Printf("🏛️  Setting: %s\n", scenario.Setting)
	fmt.Println()
	fmt.Printf("📖 Backstory: %s\n", scenario.BackStory)
	fmt.Println()
	
	gameLoop(engine, llmClient)
}

func loadOrCreateScenario(llmClient *llm.Client) (*game.Scenario, error) {
	saveDir := filepath.Join(os.Getenv("HOME"), SaveDir)
	saveFile := filepath.Join(saveDir, SaveFile)
	
	// Check if scenario already exists
	if data, err := ioutil.ReadFile(saveFile); err == nil {
		fmt.Println("Loading existing scenario...")
		return game.ScenarioFromJSON(data)
	}
	
	fmt.Println("Creating new escape room scenario...")
	fmt.Print("Enter a theme (or press Enter for random): ")
	
	reader := bufio.NewReader(os.Stdin)
	theme, _ := reader.ReadString('\n')
	theme = strings.TrimSpace(theme)
	
	if theme == "" {
		themes := []string{
			"Haunted Victorian Mansion",
			"Abandoned Space Station",
			"Ancient Egyptian Tomb",
			"Mad Scientist's Laboratory",
			"Pirate Ship",
			"Time Machine Malfunction",
			"Zombie Apocalypse Bunker",
			"Magic Academy",
			"Bank Heist Gone Wrong",
			"Underwater Research Base",
		}
		theme = themes[int(time.Now().Unix())%len(themes)]
		fmt.Printf("Generated theme: %s\n", theme)
	}
	
	scenario, err := generateOrUseFallback(llmClient, theme)
	if err != nil {
		return nil, err
	}
	
	// Save scenario for future use
	if err := os.MkdirAll(saveDir, 0755); err == nil {
		if data, err := scenario.ToJSON(); err == nil {
			ioutil.WriteFile(saveFile, data, 0644)
			fmt.Println("Scenario saved for future games!")
		}
	}
	
	return scenario, nil
}

func generateOrUseFallback(llmClient *llm.Client, theme string) (*game.Scenario, error) {
	if llmClient != nil {
		fmt.Println("Generating scenario with AI...")
		scenario, err := llmClient.GenerateScenario(theme)
		if err == nil {
			return scenario, nil
		}
		fmt.Printf("AI generation failed (%v), using fallback scenario...\n", err)
	} else {
		fmt.Println("Using fallback scenario (no API key)...")
	}
	
	return createFallbackScenario(theme), nil
}

func createFallbackScenario(theme string) *game.Scenario {
	return &game.Scenario{
		Theme:     theme,
		Setting:   "A mysterious locked room",
		BackStory: "You wake up in a strange room with no memory of how you got here. The door is locked, and you must find a way to escape.",
		Rooms: []game.Room{
			{
				ID:          "main_room",
				Name:        "Main Room",
				Description: "A dimly lit room with stone walls and a heavy wooden door. There's a desk in the corner and strange symbols on the walls.",
				Items:       []string{"key", "note", "candle"},
				Puzzles:     []string{"door_puzzle", "symbol_puzzle"},
				Exits:       []string{"exit"},
				Locked:      false,
			},
			{
				ID:          "exit",
				Name:        "Freedom",
				Description: "The way out!",
				Items:       []string{},
				Puzzles:     []string{},
				Exits:       []string{},
				Locked:      true,
				UnlockKey:   "final_key",
			},
		},
		Items: []game.Item{
			{
				ID:          "key",
				Name:        "old key",
				Description: "A rusty old key with strange markings.",
				Usable:      true,
				Hidden:      false,
			},
			{
				ID:          "note",
				Name:        "cryptic note",
				Description: "A note that reads: 'The answer lies in the stars above, count them well.'",
				Usable:      false,
				Hidden:      false,
			},
			{
				ID:          "candle",
				Name:        "candle",
				Description: "A half-melted candle. It might illuminate hidden things.",
				Usable:      true,
				UseWith:     "symbols",
				Hidden:      false,
			},
			{
				ID:          "final_key",
				Name:        "golden key",
				Description: "A beautiful golden key that seems to glow with inner light.",
				Usable:      true,
				Hidden:      true,
				RevealedBy:  "symbol_puzzle",
			},
		},
		Puzzles: []game.Puzzle{
			{
				ID:            "door_puzzle",
				Name:          "Locked Door",
				Description:   "The main door is locked with an old lock.",
				Solution:      "use key",
				RequiredItems: []string{"key"},
				Reward:        "The door creaks open, but there's another challenge ahead.",
				Solved:        false,
			},
			{
				ID:            "symbol_puzzle",
				Name:          "Wall Symbols",
				Description:   "Strange symbols are carved into the wall. They seem to form a pattern.",
				Solution:      "seven",
				RequiredItems: []string{"candle"},
				Reward:        "The symbols glow and reveal a hidden compartment with a golden key!",
				Solved:        false,
			},
		},
		WinCondition: "Solve all puzzles and unlock the final door.",
		Hints: map[string]string{
			"main_room": "Look carefully at everything in the room. The note might have a clue.",
		},
	}
}

func gameLoop(engine *game.Engine, llmClient *llm.Client) {
	reader := bufio.NewReader(os.Stdin)
	
	// Initial room description
	room, _ := engine.GetCurrentRoom()
	fmt.Printf("🚪 %s\n", room.Description)
	fmt.Println()
	
	for {
		if engine.IsGameWon() {
			fmt.Println("🎉 Congratulations! You've escaped! 🎉")
			fmt.Printf("📊 Final stats: %s\n", engine.GetGameStats())
			break
		}
		
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		
		if input == "" {
			continue
		}
		
		if strings.ToLower(input) == "quit" || strings.ToLower(input) == "exit" {
			fmt.Println("Thanks for playing!")
			break
		}
		
		if strings.ToLower(input) == "help" {
			printHelp()
			continue
		}
		
		if strings.ToLower(input) == "stats" {
			fmt.Printf("📊 %s\n", engine.GetGameStats())
			continue
		}
		
		// Process command
		err := engine.ProcessCommand(input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		
		// Get current state for narration
		state := engine.GetState()
		currentRoom, _ := engine.GetCurrentRoom()
		
		// Generate AI narration
		narration, err := generateNarration(llmClient, state, currentRoom, input)
		if err == nil && narration != "" {
			fmt.Printf("🤖 %s\n", narration)
		} else {
			// Fallback to basic result
			if state.LastResult != "" {
				fmt.Printf("📝 %s\n", state.LastResult)
			}
		}
		
		fmt.Println()
	}
}

func generateNarration(llmClient *llm.Client, state *game.GameState, room *game.Room, input string) (string, error) {
	if llmClient == nil {
		return "", fmt.Errorf("no LLM client")
	}
	
	ctx := llm.NarrationContext{
		CurrentRoom: room,
		LastAction:  state.LastAction,
		LastResult:  state.LastResult,
		Inventory:   state.Inventory,
		GameState:   state,
		PlayerInput: input,
	}
	
	return llmClient.GenerateNarration(ctx)
}

func printHelp() {
	fmt.Println("🆘 Available commands:")
	fmt.Println("  look [item]     - Examine your surroundings or a specific item")
	fmt.Println("  take <item>     - Pick up an item")
	fmt.Println("  use <item>      - Use an item from your inventory")
	fmt.Println("  go <direction>  - Move to a different room")
	fmt.Println("  inventory       - Check what you're carrying")
	fmt.Println("  solve <answer>  - Attempt to solve a puzzle")
	fmt.Println("  hint            - Get a hint for the current room")
	fmt.Println("  stats           - Show game statistics")
	fmt.Println("  help            - Show this help message")
	fmt.Println("  quit            - Exit the game")
	fmt.Println()
}