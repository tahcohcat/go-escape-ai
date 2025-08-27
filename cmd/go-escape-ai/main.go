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
		scenario, jsonErr := game.ScenarioFromJSON(data)
		if jsonErr == nil && len(scenario.Actions) > 0 && len(scenario.ProgressiveHints) > 0 {
			// Only use saved scenario if it has the new action system
			return scenario, nil
		}
		fmt.Println("Existing scenario is outdated, creating new one...")
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
		Setting:   "Your late uncle's study",
		BackStory: "You've inherited your eccentric uncle's house. The study door slammed shut behind you and won't budge. Your uncle was known for his clever puzzles and hidden treasures. There must be a way out that reveals what he left for you.",
		Rooms: []game.Room{
			{
				ID:          "study",
				Name:        "Uncle's Study",
				Description: "A cozy study filled with your uncle's eclectic collections. Tall bookshelves line the walls, packed with leather-bound volumes. His mahogany desk dominates the center, cluttered with papers, a vintage brass compass, and an ornate letter opener. A grandfather clock ticks steadily in the corner, its pendulum catching the light. Above the fireplace hangs a large painting of a sailing ship. The door you entered through won't budge - there must be another way out.",
				Items:       []string{"compass", "letter_opener", "ship_painting", "loose_book", "desk_drawer"},
				Puzzles:     []string{"painting_puzzle", "clock_puzzle"},
				Exits:       []string{"hidden_passage"},
				Locked:      false,
			},
			{
				ID:          "hidden_passage",
				Name:        "Hidden Passage",
				Description: "A narrow stone passage behind the bookshelf, lit by flickering torches. The walls are carved with nautical symbols and star charts. At the end stands an ornate treasure chest with multiple locks, and beside it, a final door marked with your family's crest.",
				Items:       []string{"treasure_map", "family_letter", "gold_coins"},
				Puzzles:     []string{"treasure_chest"},
				Exits:       []string{"victory"},
				Locked:      true,
				UnlockKey:   "solved_puzzles",
			},
			{
				ID:          "victory",
				Name:        "Freedom and Fortune",
				Description: "You step into the garden, breathing fresh air and clutching your uncle's final gift - both treasure and the knowledge that you were clever enough to earn it.",
				Items:       []string{},
				Puzzles:     []string{},
				Exits:       []string{},
				Locked:      false,
			},
		},
		Items: []game.Item{
			{
				ID:          "compass",
				Name:        "brass compass",
				Description: "An antique brass compass with intricate engravings. The needle points north, but when you rotate it, you notice it has multiple directional markings and seems to be more than just a navigation tool.",
				Usable:      true,
				Hidden:      false,
			},
			{
				ID:          "letter_opener",
				Name:        "ornate letter opener",
				Description: "A beautiful silver letter opener with your family crest on the handle. It's sharp and well-crafted, and seems important to your uncle.",
				Usable:      true,
				Hidden:      false,
			},
			{
				ID:          "ship_painting",
				Name:        "ship painting",
				Description: "A large oil painting of a three-masted sailing ship on stormy seas. The frame is ornate and heavy. Something about the ship's flag seems familiar, and you notice the painting isn't quite flush against the wall.",
				Usable:      false,
				Hidden:      false,
			},
			{
				ID:          "loose_book",
				Name:        "leather journal",
				Description: "A well-worn leather journal that was sticking out from the bookshelf. It contains your uncle's notes about 'the family legacy' and sketches of compass directions, star charts, and what looks like a map of this very house.",
				Usable:      true,
				Hidden:      true,
				RevealedBy:  "examine_books",
			},
			{
				ID:          "desk_drawer",
				Name:        "mysterious key",
				Description: "An unusual key made of dark metal with astronomical symbols etched along its length. It's heavier than it looks and seems to be made for something important.",
				Usable:      true,
				Hidden:      true,
				RevealedBy:  "examine_desk",
			},
			{
				ID:          "solved_puzzles",
				Name:        "knowledge",
				Description: "Your understanding of your uncle's puzzles - this opens the way forward.",
				Usable:      true,
				Hidden:      true,
			},
			{
				ID:          "treasure_map",
				Name:        "treasure map",
				Description: "An authentic treasure map showing several islands and marked with an X. This must be worth a fortune to collectors!",
				Usable:      false,
				Hidden:      false,
			},
			{
				ID:          "family_letter",
				Name:        "family letter",
				Description: "A letter from your uncle: 'My dear heir, if you're reading this, you've proven yourself worthy of the family treasure. Use it wisely, and remember - the real treasure was the puzzles we solved along the way.'",
				Usable:      false,
				Hidden:      false,
			},
			{
				ID:          "gold_coins",
				Name:        "gold coins",
				Description: "A collection of genuine Spanish doubloons and other antique coins. Your uncle's treasure is real!",
				Usable:      false,
				Hidden:      false,
			},
		},
		Puzzles: []game.Puzzle{
			{
				ID:            "painting_puzzle",
				Name:          "The Ship's Secret",
				Description:   "The ship painting isn't quite flush with the wall. Maybe something can move it aside to reveal what's behind.",
				Solution:      "move painting",
				RequiredItems: []string{"letter_opener"},
				Reward:        "You use the letter opener to carefully pry the painting aside. Behind it is a hidden panel with a compass rose carved into the wood!",
				Solved:        false,
			},
			{
				ID:            "clock_puzzle",
				Name:          "Time and Direction",
				Description:   "The grandfather clock shows the current time, but the compass rose behind the painting has directional markings. Your uncle's journal mentions 'when the clock points north'.",
				Solution:      "twelve",
				RequiredItems: []string{"compass", "loose_book"},
				Reward:        "You realize 12 o'clock is north on a compass! You set the clock hands to 12, and hear a mechanism grinding behind the bookshelf.",
				Solved:        false,
			},
			{
				ID:            "treasure_chest",
				Name:          "Uncle's Legacy",
				Description:   "A beautiful chest with your family crest. It has three keyholes - but you only found one key. The other locks seem to respond to something else.",
				Solution:      "use key and compass",
				RequiredItems: []string{"desk_drawer", "compass"},
				Reward:        "The mysterious key fits perfectly! The compass, when placed in a depression on the lid, completes the mechanism. The chest opens to reveal maps, gold, and your uncle's final letter.",
				Solved:        false,
			},
		},
		Actions: []game.Action{
			{
				ID: "examine_desk",
				Trigger: game.ActionTrigger{
					Type:   "examine",
					Target: "desk",
				},
				Effects: []game.ActionEffect{
					{
						Type:   "reveal_item",
						Target: "desk_drawer",
					},
				},
				Message: "You search through your uncle's desk drawers. Most are locked, but one slides open to reveal an unusual key made of dark metal with strange astronomical symbols.",
				OneTimeOnly: true,
			},
			{
				ID: "examine_books",
				Trigger: game.ActionTrigger{
					Type:   "examine",
					Target: "bookshelf",
				},
				Effects: []game.ActionEffect{
					{
						Type:   "reveal_item",
						Target: "loose_book",
					},
				},
				Message: "As you scan the bookshelves, one leather journal catches your eye - it's sticking out slightly. You pull it free to find your uncle's personal notes about the 'family legacy' and detailed sketches.",
				OneTimeOnly: true,
			},
			{
				ID: "move_painting",
				Trigger: game.ActionTrigger{
					Type:   "use_with",
					Target: "letter_opener",
					With:   "painting",
				},
				Conditions: []game.ActionCondition{
					{
						Type:  "has_item",
						Value: "letter_opener",
					},
				},
				Effects: []game.ActionEffect{
					{
						Type:   "add_inventory",
						Target: "solved_puzzles",
					},
				},
				Message: "You carefully use the letter opener to pry the painting aside. Behind it, carved into the wood, is an intricate compass rose with directional markings!",
				OneTimeOnly: true,
			},
			{
				ID: "solve_clock",
				Trigger: game.ActionTrigger{
					Type:   "use_with",
					Target: "compass",
					With:   "clock",
				},
				Conditions: []game.ActionCondition{
					{
						Type:  "has_item",
						Value: "loose_book",
					},
					{
						Type:  "puzzle_solved",
						Value: "painting_puzzle",
					},
				},
				Effects: []game.ActionEffect{
					{
						Type:   "unlock_room",
						Target: "hidden_passage",
					},
				},
				Message: "Following your uncle's notes, you realize that 12 o'clock corresponds to north on the compass rose! You adjust the clock hands, and with a grinding sound, the bookshelf swings aside to reveal a hidden passage!",
				OneTimeOnly: true,
			},
		},
		WinCondition: "Discover your uncle's clues, solve his puzzles, and claim the family treasure.",
		Hints: map[string]string{
			"study": "Your uncle left clues throughout his study. The desk, bookshelves, painting, and grandfather clock all seem important. Start by examining them carefully.",
			"hidden_passage": "You've found your uncle's secret chamber! The treasure chest requires both the mysterious key and something else to complete the mechanism.",
		},
		ProgressiveHints: []game.ProgressiveHint{
			{
				Context: "study",
				Triggers: []game.HintTrigger{
					{Type: "commands_tried", Threshold: 3},
				},
				HintText: "Your uncle was methodical. Start by examining the obvious things: his desk, the bookshelves, and that interesting painting.",
				Priority: 1,
			},
			{
				Context: "study",
				Triggers: []game.HintTrigger{
					{Type: "commands_tried", Threshold: 7},
				},
				HintText: "That painting of the ship looks like it could be moved. The letter opener might be useful for more than just opening letters.",
				Priority: 2,
			},
			{
				Context: "study",
				Triggers: []game.HintTrigger{
					{Type: "commands_tried", Threshold: 12},
				},
				HintText: "The compass, journal, and clock seem connected. Your uncle's notes mention 'when the clock points north' - think about how clocks and compasses relate.",
				Priority: 3,
			},
			{
				Context: "painting_puzzle",
				Triggers: []game.HintTrigger{
					{Type: "failed_attempts", Threshold: 2},
				},
				HintText: "The painting isn't flush with the wall. Try using a tool to carefully move it aside.",
				Priority: 4,
			},
			{
				Context: "clock_puzzle",
				Triggers: []game.HintTrigger{
					{Type: "failed_attempts", Threshold: 2},
				},
				HintText: "On a compass, north is at 12 o'clock. Your uncle's journal talks about 'when the clock points north'.",
				Priority: 5,
			},
		},
	}
}

func gameLoop(engine *game.Engine, llmClient *llm.Client) {
	reader := bufio.NewReader(os.Stdin)
	
	// Initial room description - show exact factual description
	room, _ := engine.GetCurrentRoom()
	fmt.Printf("📝 %s\n", room.Description)
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
		
		// Always show the factual result first
		if state.LastResult != "" {
			fmt.Printf("📝 %s\n", state.LastResult)
		}
		
		// Then add atmospheric narration if available
		narration, err := generateNarration(llmClient, engine, currentRoom, input)
		if err == nil && narration != "" {
			fmt.Printf("🤖 %s\n", narration)
		}
		
		fmt.Println()
	}
}

func generateNarration(llmClient *llm.Client, engine *game.Engine, room *game.Room, input string) (string, error) {
	if llmClient == nil {
		return "", fmt.Errorf("no LLM client")
	}
	
	state := engine.GetState()
	hints := engine.GetProgressiveHints()
	
	ctx := llm.NarrationContext{
		CurrentRoom:      room,
		LastAction:       state.LastAction,
		LastResult:       state.LastResult,
		Inventory:        state.Inventory,
		GameState:        state,
		PlayerInput:      input,
		ProgressiveHints: hints,
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