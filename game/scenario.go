package game

import (
	"encoding/json"
	"fmt"
)

type Scenario struct {
	Theme       string            `json:"theme"`
	Setting     string            `json:"setting"`
	BackStory   string            `json:"backstory"`
	Rooms       []Room            `json:"rooms"`
	Items       []Item            `json:"items"`
	Puzzles     []Puzzle          `json:"puzzles"`
	WinCondition string           `json:"win_condition"`
	Hints       map[string]string `json:"hints"`
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