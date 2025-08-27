package game

import (
	"fmt"
	"strings"
	"time"
)

type GameState struct {
	Scenario          *Scenario         `json:"scenario"`
	CurrentRoom       string            `json:"current_room"`
	Inventory         []string          `json:"inventory"`
	SolvedPuzzles     []string          `json:"solved_puzzles"`
	DiscoveredItems   []string          `json:"discovered_items"`
	PerformedActions  []string          `json:"performed_actions"`
	FailedAttempts    map[string]int    `json:"failed_attempts"`
	CommandAttempts   int               `json:"command_attempts"`
	Moves             int               `json:"moves"`
	StartTime         time.Time         `json:"start_time"`
	GameWon           bool              `json:"game_won"`
	LastAction        string            `json:"last_action"`
	LastResult        string            `json:"last_result"`
}

type Engine struct {
	state *GameState
}

func NewEngine(scenario *Scenario) *Engine {
	return &Engine{
		state: &GameState{
			Scenario:         scenario,
			CurrentRoom:      scenario.Rooms[0].ID, // Start in first room
			Inventory:        []string{},
			SolvedPuzzles:    []string{},
			DiscoveredItems:  []string{},
			PerformedActions: []string{},
			FailedAttempts:   make(map[string]int),
			CommandAttempts:  0,
			Moves:            0,
			StartTime:        time.Now(),
			GameWon:          false,
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
	e.state.CommandAttempts++
	e.state.LastAction = command
	
	parts := strings.Fields(strings.ToLower(command))
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	verb := parts[0]
	
	switch verb {
	case "look", "examine":
		target := strings.Join(parts[1:], " ")
		// First try actions - they can override default behavior
		actionProcessed := e.processActions("examine", target, "")
		if !actionProcessed {
			return e.handleLook(parts[1:])
		}
		return nil
	case "take", "get", "pick":
		target := strings.Join(parts[1:], " ")
		err := e.handleTake(parts[1:])
		e.processActions("take", target, "")
		return err
	case "use":
		target := strings.Join(parts[1:], " ")
		err := e.handleUse(parts[1:])
		if strings.Contains(target, " with ") || strings.Contains(target, " on ") {
			parts := strings.Split(target, " with ")
			if len(parts) == 1 {
				parts = strings.Split(target, " on ")
			}
			if len(parts) == 2 {
				e.processActions("use_with", strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			}
		} else {
			e.processActions("use", target, "")
		}
		return err
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
		e.state.LastResult = room.Description
		return nil
	}
	
	target := strings.Join(args, " ")
	room, _ := e.GetCurrentRoom()
	
	// Check for room features mentioned in description
	roomDesc := strings.ToLower(room.Description)
	if strings.Contains(roomDesc, strings.ToLower(target)) {
		// Generic response for room features
		e.state.LastResult = fmt.Sprintf("You examine the %s more closely, but don't notice anything special.", target)
		return nil
	}
	
	// Check items in current room
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
	
	// Check for "use X with Y" or "use X on Y" patterns
	if strings.Contains(target, " with ") || strings.Contains(target, " on ") {
		parts := strings.Split(target, " with ")
		if len(parts) == 1 {
			parts = strings.Split(target, " on ")
		}
		if len(parts) == 2 {
			return e.handleItemCombination(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}
	
	// Single item use
	for _, itemID := range e.state.Inventory {
		item, err := e.state.Scenario.GetItem(itemID)
		if err != nil {
			continue
		}
		
		if strings.Contains(strings.ToLower(item.Name), target) && item.Usable {
			e.state.LastResult = fmt.Sprintf("You use the %s.", item.Name)
			return nil
		}
	}
	
	e.state.LastResult = "You don't have that item or can't use it."
	return nil
}

func (e *Engine) handleItemCombination(item1Name, item2Name string) error {
	// Find both items
	var item1, item2 *Item
	var item1ID, item2ID string
	
	// Check inventory for item1
	for _, itemID := range e.state.Inventory {
		item, err := e.state.Scenario.GetItem(itemID)
		if err != nil {
			continue
		}
		if strings.Contains(strings.ToLower(item.Name), item1Name) {
			item1 = item
			item1ID = itemID
			break
		}
	}
	
	// Check inventory and room for item2
	for _, itemID := range e.state.Inventory {
		item, err := e.state.Scenario.GetItem(itemID)
		if err != nil {
			continue
		}
		if strings.Contains(strings.ToLower(item.Name), item2Name) {
			item2 = item
			item2ID = itemID
			break
		}
	}
	
	// If not in inventory, check current room
	if item2 == nil {
		room, _ := e.GetCurrentRoom()
		for _, itemID := range room.Items {
			item, err := e.state.Scenario.GetItem(itemID)
			if err != nil {
				continue
			}
			if strings.Contains(strings.ToLower(item.Name), item2Name) && !item.Hidden {
				item2 = item
				item2ID = itemID
				break
			}
		}
	}
	
	if item1 == nil {
		e.state.LastResult = fmt.Sprintf("You don't have %s.", item1Name)
		return nil
	}
	
	if item2 == nil {
		e.state.LastResult = fmt.Sprintf("You don't see %s here.", item2Name)
		return nil
	}
	
	// Handle specific combinations
	if item1ID == "matches" && item2ID == "candle" {
		// Light the candle
		e.state.Inventory = append(e.state.Inventory, "lit_candle")
		e.state.LastResult = "You light the candle with the matches. The room is now brightly illuminated!"
		
		// Reveal any hidden items that should be revealed by lighting candle
		for i := range e.state.Scenario.Items {
			if e.state.Scenario.Items[i].RevealedBy == "use_matches" {
				e.state.Scenario.Items[i].Hidden = false
			}
		}
		return nil
	}
	
	// Check if item1 is meant to be used with item2
	if item1.UseWith == item2ID {
		e.state.LastResult = fmt.Sprintf("You use the %s with the %s.", item1.Name, item2.Name)
		return nil
	}
	
	e.state.LastResult = fmt.Sprintf("You can't use the %s with the %s.", item1.Name, item2.Name)
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

func (e *Engine) processActions(actionType, target, withItem string) bool {
	actionProcessed := false
	
	for _, action := range e.state.Scenario.Actions {
		if e.matchesAction(action, actionType, target, withItem) {
			if action.OneTimeOnly && e.hasPerformedAction(action.ID) {
				continue
			}
			
			if e.checkActionConditions(action.Conditions) {
				e.executeActionEffects(action.Effects)
				if action.Message != "" {
					e.state.LastResult = action.Message
				}
				
				if action.OneTimeOnly {
					e.state.PerformedActions = append(e.state.PerformedActions, action.ID)
				}
				actionProcessed = true
			}
		}
	}
	
	return actionProcessed
}

func (e *Engine) matchesAction(action Action, actionType, target, withItem string) bool {
	if action.Trigger.Type != actionType {
		return false
	}
	
	if !strings.Contains(strings.ToLower(target), strings.ToLower(action.Trigger.Target)) {
		return false
	}
	
	if action.Trigger.With != "" && !strings.Contains(strings.ToLower(withItem), strings.ToLower(action.Trigger.With)) {
		return false
	}
	
	return true
}

func (e *Engine) hasPerformedAction(actionID string) bool {
	for _, performed := range e.state.PerformedActions {
		if performed == actionID {
			return true
		}
	}
	return false
}

func (e *Engine) checkActionConditions(conditions []ActionCondition) bool {
	for _, condition := range conditions {
		switch condition.Type {
		case "has_item":
			if !e.HasItem(condition.Value) {
				return false
			}
		case "in_room":
			if e.state.CurrentRoom != condition.Value {
				return false
			}
		case "puzzle_solved":
			if !e.IsPuzzleSolved(condition.Value) {
				return false
			}
		case "action_performed":
			if !e.hasPerformedAction(condition.Value) {
				return false
			}
		}
	}
	return true
}

func (e *Engine) executeActionEffects(effects []ActionEffect) {
	for _, effect := range effects {
		switch effect.Type {
		case "reveal_item":
			for i := range e.state.Scenario.Items {
				if e.state.Scenario.Items[i].ID == effect.Target {
					e.state.Scenario.Items[i].Hidden = false
					break
				}
			}
		case "hide_item":
			for i := range e.state.Scenario.Items {
				if e.state.Scenario.Items[i].ID == effect.Target {
					e.state.Scenario.Items[i].Hidden = true
					break
				}
			}
		case "unlock_room":
			for i := range e.state.Scenario.Rooms {
				if e.state.Scenario.Rooms[i].ID == effect.Target {
					e.state.Scenario.Rooms[i].Locked = false
					break
				}
			}
		case "add_inventory":
			if !e.HasItem(effect.Target) {
				e.state.Inventory = append(e.state.Inventory, effect.Target)
			}
		case "remove_inventory":
			for i, item := range e.state.Inventory {
				if item == effect.Target {
					e.state.Inventory = append(e.state.Inventory[:i], e.state.Inventory[i+1:]...)
					break
				}
			}
		}
	}
}

func (e *Engine) GetProgressiveHints() []string {
	var hints []string
	
	for _, hint := range e.state.Scenario.ProgressiveHints {
		if e.shouldShowHint(hint) {
			hints = append(hints, hint.HintText)
		}
	}
	
	return hints
}

func (e *Engine) shouldShowHint(hint ProgressiveHint) bool {
	contextMatch := false
	
	// Check if hint applies to current context
	if hint.Context == e.state.CurrentRoom {
		contextMatch = true
	}
	
	// Check if it's a puzzle-specific hint for unsolved puzzles in current room
	if !contextMatch {
		room, _ := e.GetCurrentRoom()
		for _, puzzleID := range room.Puzzles {
			if hint.Context == puzzleID && !e.IsPuzzleSolved(puzzleID) {
				contextMatch = true
				break
			}
		}
	}
	
	if !contextMatch {
		return false
	}
	
	// Check if any trigger conditions are met
	for _, trigger := range hint.Triggers {
		switch trigger.Type {
		case "failed_attempts":
			if e.state.FailedAttempts[hint.Context] >= trigger.Threshold {
				return true
			}
		case "time_spent":
			duration := time.Since(e.state.StartTime)
			if int(duration.Minutes()) >= trigger.Threshold {
				return true
			}
		case "commands_tried":
			if e.state.CommandAttempts >= trigger.Threshold {
				return true
			}
		}
	}
	
	return false
}