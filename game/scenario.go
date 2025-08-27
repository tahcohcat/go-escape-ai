package game

import (
	"encoding/json"
	"fmt"
)

type Scenario struct {
	Theme        string            `json:"theme"`
	Setting      string            `json:"setting"`
	BackStory    string            `json:"backstory"`
	Rooms        []Room            `json:"rooms"`
	Items        []Item            `json:"items"`
	Puzzles      []Puzzle          `json:"puzzles"`
	Actions      []Action          `json:"actions"`
	WinCondition string            `json:"win_condition"`
	Hints        map[string]string `json:"hints"`
	ProgressiveHints []ProgressiveHint `json:"progressive_hints"`
}

type Room struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Items       []string `json:"items"`
	Puzzles     []string `json:"puzzles"`
	Exits       []string `json:"exits"`
	Locked      bool     `json:"locked"`
	UnlockKey   string   `json:"unlock_key,omitempty"`
}

type Item struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Usable      bool   `json:"usable"`
	UseWith     string `json:"use_with,omitempty"`
	Hidden      bool   `json:"hidden"`
	RevealedBy  string `json:"revealed_by,omitempty"`
}

type Puzzle struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Solution    string   `json:"solution"`
	RequiredItems []string `json:"required_items"`
	Reward      string   `json:"reward"`
	Solved      bool     `json:"solved"`
}

type Action struct {
	ID          string            `json:"id"`
	Trigger     ActionTrigger     `json:"trigger"`
	Conditions  []ActionCondition `json:"conditions"`
	Effects     []ActionEffect    `json:"effects"`
	Message     string            `json:"message"`
	OneTimeOnly bool              `json:"one_time_only"`
}

type ActionTrigger struct {
	Type   string `json:"type"` // "examine", "use", "use_with", "take"
	Target string `json:"target"` // item ID, room feature, etc.
	With   string `json:"with,omitempty"` // for "use_with" actions
}

type ActionCondition struct {
	Type  string `json:"type"` // "has_item", "in_room", "puzzle_solved", "action_performed"
	Value string `json:"value"`
}

type ActionEffect struct {
	Type   string `json:"type"` // "reveal_item", "hide_item", "unlock_room", "add_inventory", "remove_inventory"
	Target string `json:"target"`
	Value  string `json:"value,omitempty"`
}

type ProgressiveHint struct {
	Context     string   `json:"context"` // "room_id" or "puzzle_id" 
	Triggers    []HintTrigger `json:"triggers"`
	HintText    string   `json:"hint_text"`
	Priority    int      `json:"priority"` // Higher priority hints show first
}

type HintTrigger struct {
	Type      string `json:"type"` // "failed_attempts", "time_spent", "commands_tried"
	Threshold int    `json:"threshold"`
}

func (s *Scenario) ToJSON() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

func ScenarioFromJSON(data []byte) (*Scenario, error) {
	var scenario Scenario
	err := json.Unmarshal(data, &scenario)
	return &scenario, err
}

func (s *Scenario) GetRoom(id string) (*Room, error) {
	for i := range s.Rooms {
		if s.Rooms[i].ID == id {
			return &s.Rooms[i], nil
		}
	}
	return nil, fmt.Errorf("room %s not found", id)
}

func (s *Scenario) GetItem(id string) (*Item, error) {
	for i := range s.Items {
		if s.Items[i].ID == id {
			return &s.Items[i], nil
		}
	}
	return nil, fmt.Errorf("item %s not found", id)
}

func (s *Scenario) GetPuzzle(id string) (*Puzzle, error) {
	for i := range s.Puzzles {
		if s.Puzzles[i].ID == id {
			return &s.Puzzles[i], nil
		}
	}
	return nil, fmt.Errorf("puzzle %s not found", id)
}