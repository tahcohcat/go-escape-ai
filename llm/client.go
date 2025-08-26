package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
	"github.com/tahcohcat/go-escape-ai/game"
)

type Client struct {
	client *openai.Client
}

func NewClient() *Client {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Warning: OPENAI_API_KEY not set. LLM features will not work.")
		return nil
	}
	
	return &Client{
		client: openai.NewClient(apiKey),
	}
}

func (c *Client) GenerateScenario(theme string) (*game.Scenario, error) {
	if c == nil || c.client == nil {
		return nil, fmt.Errorf("LLM client not initialized")
	}

	prompt := fmt.Sprintf(`Generate a complete escape room scenario with the theme: %s

Create a JSON response with the following structure:
- theme: The main theme
- setting: Where the escape room takes place
- backstory: Brief backstory explaining why the player is trapped
- rooms: Array of 3-5 interconnected rooms with:
  - id, name, description
  - items (array of item IDs found in this room)
  - puzzles (array of puzzle IDs in this room)
  - exits (array of room IDs this room connects to)
  - locked (boolean), unlock_key (item ID needed to unlock)
- items: Array of interactive objects with:
  - id, name, description
  - usable (boolean), use_with (what it's used with)
  - hidden (boolean), revealed_by (item ID that reveals this)
- puzzles: Array of challenges with:
  - id, name, description, solution
  - required_items (array of item IDs needed)
  - reward (what happens when solved)
- win_condition: How the player escapes
- hints: Object mapping room IDs to helpful hints

Make it challenging but solvable. Include at least 3 puzzles and 5 items. Some items should be hidden initially.`, theme)

	resp, err := c.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a creative escape room designer. Generate detailed, immersive scenarios with logical puzzles and interconnected elements. Always respond with valid JSON only.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.8,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to generate scenario: %w", err)
	}

	var scenario game.Scenario
	content := resp.Choices[0].Message.Content
	
	if err := json.Unmarshal([]byte(content), &scenario); err != nil {
		return nil, fmt.Errorf("failed to parse generated scenario: %w", err)
	}

	return &scenario, nil
}

type NarrationContext struct {
	CurrentRoom    *game.Room
	LastAction     string
	LastResult     string
	Inventory      []string
	GameState      *game.GameState
	PlayerInput    string
}

func (c *Client) GenerateNarration(ctx NarrationContext) (string, error) {
	if c == nil || c.client == nil {
		// Fallback to basic narration if LLM unavailable
		return c.fallbackNarration(ctx), nil
	}

	scenario := ctx.GameState.Scenario
	
	prompt := fmt.Sprintf(`You are an AI narrator for an escape room game. Provide immersive, atmospheric narration based on the current situation.

Theme: %s
Setting: %s
Current Room: %s - %s

Last Action: %s
Last Result: %s
Player Input: %s

Inventory: %v
Moves: %d
Puzzles Solved: %d/%d

Provide a 2-3 sentence response that:
1. Acknowledges the player's action
2. Describes the current atmosphere/mood
3. Gives subtle hints about what to do next (without being obvious)

Keep the tone mysterious and engaging. Don't repeat the exact room description.`,
		scenario.Theme,
		scenario.Setting,
		ctx.CurrentRoom.Name,
		ctx.CurrentRoom.Description,
		ctx.LastAction,
		ctx.LastResult,
		ctx.PlayerInput,
		ctx.Inventory,
		ctx.GameState.Moves,
		len(ctx.GameState.SolvedPuzzles),
		len(scenario.Puzzles))

	resp, err := c.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a mysterious, slightly ominous AI narrator for an escape room. Be atmospheric and engaging, but don't give away solutions directly.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.7,
			MaxTokens:   150,
		},
	)

	if err != nil {
		return c.fallbackNarration(ctx), nil
	}

	return resp.Choices[0].Message.Content, nil
}

func (c *Client) fallbackNarration(ctx NarrationContext) string {
	responses := []string{
		fmt.Sprintf("%s The air feels thick with mystery.", ctx.LastResult),
		fmt.Sprintf("%s Something important must be nearby.", ctx.LastResult),
		fmt.Sprintf("%s You sense you're getting closer to the truth.", ctx.LastResult),
		fmt.Sprintf("%s The silence is almost deafening.", ctx.LastResult),
		fmt.Sprintf("%s Every detail might be crucial.", ctx.LastResult),
	}
	
	// Simple hash to pick consistent response
	index := (len(ctx.LastAction) + ctx.GameState.Moves) % len(responses)
	return responses[index]
}