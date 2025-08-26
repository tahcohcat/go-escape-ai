package game

import (
	"fmt"
	"strings"
	"time"
)

type GameState struct {
	Scenario       *Scenario         `json:"scenario"`
	CurrentRoom    string            `json:"current_room"`
	Inventory      []string          `json:"inventory"`
	SolvedPuzzles  []string          `json:"solved_puzzles"`
	DiscoveredItems []string         `json:"discovered_items"`
	Moves          int               `json:"moves"`
	StartTime      time.Time         `json:"start_time"`
	GameWon        bool              `json:"game_won"`
	LastAction     string            `json:"last_action"`
	LastResult     string            `json:"last_result"`
}

type Engine struct {
	state *GameState
}

func NewEngine(scenario *Scenario) *Engine {
	return &Engine{
		state: &GameState{
			Scenario:        scenario,
			CurrentRoom:     scenario.Rooms[0].ID, // Start in first room
			Inventory:       []string{},
			SolvedPuzzles:   []string{},
			DiscoveredItems: []string{},
			Moves:           0,
			StartTime:       time.Now(),
			GameWon:         false,
		},
	}
}

func (e *Engine) GetState() *GameState {
	return e.state
}

func (e *Engine) GetCurrentRoom() (*Room, error) {
	return e.state.Scenario.GetRoom(e.state.CurrentRoom)
}

func (e *Engine) HasItem(itemID string) bool {
	for _, item := range e.state.Inventory {
		if item == itemID {
			return true
		}
	}
	return false
}

func (e *Engine) IsPuzzleSolved(puzzleID string) bool {
	for _, solved := range e.state.SolvedPuzzles {
		if solved == puzzleID {
			return true
		}
	}
	return false
}

func (e *Engine) IsItemDiscovered(itemID string) bool {
	for _, discovered := range e.state.DiscoveredItems {
		if discovered == itemID {
			return true
		}
	}
	return false
}

func (e *Engine) ProcessCommand(command string) error {
	e.state.Moves++
	e.state.LastAction = command
	
	parts := strings.Fields(strings.ToLower(command))
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	verb := parts[0]
	
	switch verb {
	case "look", "examine":
		return e.handleLook(parts[1:])
	case "take", "get", "pick":
		return e.handleTake(parts[1:])
	case "use":
		return e.handleUse(parts[1:])
	case "go", "move", "walk":
		return e.handleMove(parts[1:])
	case "inventory", "inv", "i":
		return e.handleInventory()
	case "solve":
		return e.handleSolve(parts[1:])
	case "hint":
		return e.handleHint(parts[1:])
	default:
		e.state.LastResult = "I don't understand that command."
		return nil
	}
}

func (e *Engine) handleLook(args []string) error {
	if len(args) == 0 {
		room, err := e.GetCurrentRoom()
		if err != nil {
			return err
		}
		e.state.LastResult = fmt.Sprintf("You are in %s. %s", room.Name, room.Description)
		return nil
	}
	
	target := strings.Join(args, " ")
	
	// Check items in current room
	room, _ := e.GetCurrentRoom()
	for _, itemID := range room.Items {
		item, err := e.state.Scenario.GetItem(itemID)
		if err != nil {
			continue
		}
		if strings.Contains(strings.ToLower(item.Name), target) && !item.Hidden {
			e.state.LastResult = item.Description
			return nil
		}
	}
	
	e.state.LastResult = "You don't see that here."
	return nil
}

func (e *Engine) handleTake(args []string) error {
	if len(args) == 0 {
		e.state.LastResult = "Take what?"
		return nil
	}
	
	target := strings.Join(args, " ")
	room, _ := e.GetCurrentRoom()
	
	for _, itemID := range room.Items {
		item, err := e.state.Scenario.GetItem(itemID)
		if err != nil {
			continue
		}
		
		if strings.Contains(strings.ToLower(item.Name), target) && !item.Hidden {
			if !e.HasItem(itemID) {
				e.state.Inventory = append(e.state.Inventory, itemID)
				e.state.LastResult = fmt.Sprintf("You take the %s.", item.Name)
				return nil
			} else {
				e.state.LastResult = "You already have that."
				return nil
			}
		}
	}
	
	e.state.LastResult = "You don't see that here."
	return nil
}

func (e *Engine) handleUse(args []string) error {
	if len(args) == 0 {
		e.state.LastResult = "Use what?"
		return nil
	}
	
	target := strings.Join(args, " ")
	
	for _, itemID := range e.state.Inventory {
		item, err := e.state.Scenario.GetItem(itemID)
		if err != nil {
			continue
		}
		
		if strings.Contains(strings.ToLower(item.Name), target) && item.Usable {
			// Check if item reveals something
			if item.RevealedBy != "" {
				for _, otherItemID := range e.state.Scenario.Items {
					if otherItemID.ID == item.RevealedBy && otherItemID.Hidden {
						otherItemID.Hidden = false
						e.state.DiscoveredItems = append(e.state.DiscoveredItems, otherItemID.ID)
					}
				}
			}
			
			e.state.LastResult = fmt.Sprintf("You use the %s.", item.Name)
			return nil
		}
	}
	
	e.state.LastResult = "You don't have that item or can't use it."
	return nil
}

func (e *Engine) handleMove(args []string) error {
	if len(args) == 0 {
		e.state.LastResult = "Go where?"
		return nil
	}
	
	direction := strings.Join(args, " ")
	room, _ := e.GetCurrentRoom()
	
	for _, exitID := range room.Exits {
		exitRoom, err := e.state.Scenario.GetRoom(exitID)
		if err != nil {
			continue
		}
		
		if strings.Contains(strings.ToLower(exitRoom.Name), direction) || strings.Contains(strings.ToLower(exitID), direction) {
			if exitRoom.Locked {
				if exitRoom.UnlockKey == "" || !e.HasItem(exitRoom.UnlockKey) {
					e.state.LastResult = "That way is locked."
					return nil
				}
			}
			
			e.state.CurrentRoom = exitID
			e.state.LastResult = fmt.Sprintf("You move to %s.", exitRoom.Name)
			return nil
		}
	}
	
	e.state.LastResult = "You can't go that way."
	return nil
}

func (e *Engine) handleInventory() error {
	if len(e.state.Inventory) == 0 {
		e.state.LastResult = "Your inventory is empty."
		return nil
	}
	
	var items []string
	for _, itemID := range e.state.Inventory {
		item, err := e.state.Scenario.GetItem(itemID)
		if err == nil {
			items = append(items, item.Name)
		}
	}
	
	e.state.LastResult = "You have: " + strings.Join(items, ", ")
	return nil
}

func (e *Engine) handleSolve(args []string) error {
	if len(args) == 0 {
		e.state.LastResult = "Solve what?"
		return nil
	}
	
	answer := strings.Join(args, " ")
	room, _ := e.GetCurrentRoom()
	
	for _, puzzleID := range room.Puzzles {
		puzzle, err := e.state.Scenario.GetPuzzle(puzzleID)
		if err != nil {
			continue
		}
		
		if e.IsPuzzleSolved(puzzleID) {
			continue
		}
		
		// Check if player has required items
		hasAllItems := true
		for _, requiredItem := range puzzle.RequiredItems {
			if !e.HasItem(requiredItem) {
				hasAllItems = false
				break
			}
		}
		
		if !hasAllItems {
			e.state.LastResult = "You don't have everything needed to solve this puzzle."
			return nil
		}
		
		if strings.EqualFold(answer, puzzle.Solution) {
			e.state.SolvedPuzzles = append(e.state.SolvedPuzzles, puzzleID)
			e.state.LastResult = fmt.Sprintf("Correct! %s", puzzle.Reward)
			
			// Check win condition
			if len(e.state.SolvedPuzzles) >= len(e.state.Scenario.Puzzles) {
				e.state.GameWon = true
			}
			
			return nil
		} else {
			e.state.LastResult = "That's not correct."
			return nil
		}
	}
	
	e.state.LastResult = "There's no puzzle here to solve."
	return nil
}

func (e *Engine) handleHint(args []string) error {
	room, _ := e.GetCurrentRoom()
	
	if hint, exists := e.state.Scenario.Hints[room.ID]; exists {
		e.state.LastResult = hint
	} else {
		e.state.LastResult = "No hints available for this location."
	}
	
	return nil
}

func (e *Engine) IsGameWon() bool {
	return e.state.GameWon
}

func (e *Engine) GetGameStats() string {
	duration := time.Since(e.state.StartTime)
	return fmt.Sprintf("Moves: %d, Time: %v, Puzzles solved: %d/%d", 
		e.state.Moves, 
		duration.Round(time.Second),
		len(e.state.SolvedPuzzles),
		len(e.state.Scenario.Puzzles))
}