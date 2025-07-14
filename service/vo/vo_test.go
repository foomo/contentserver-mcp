package vo

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestDocument(t *testing.T) {
	doc := Document{
		DocumentSummary: DocumentSummary{
			URL: "/recipes/italian",
			ContentSummary: ContentSummary{
				Title:       "Italian Recipes",
				Description: "Discover authentic Italian cuisine with traditional recipes from pasta dishes to regional specialties.",
			},
		},
		Articles: []Article{
			{
				ContentSummary: ContentSummary{
					Title:       "Essential Italian Cooking Techniques",
					Description: "Master the fundamental techniques that define authentic Italian cuisine.",
				},
				Markdown: `# Italian Recipes

Explore the rich culinary traditions of Italy with our collection of authentic recipes. From classic pasta dishes to regional specialties, discover the flavors that make Italian cuisine beloved worldwide.

## Popular Categories

- **Pasta Dishes**: From simple pomodoro to complex carbonara
- **Pizza & Bread**: Traditional Neapolitan pizza and focaccia
- **Seafood**: Mediterranean fish and shellfish preparations
- **Desserts**: Tiramisu, cannoli, and gelato recipes
- **Regional Specialties**: Dishes from Tuscany, Sicily, and beyond

## Cooking Tips

- Use high-quality olive oil for authentic flavor
- Fresh herbs like basil, oregano, and rosemary are essential
- Don't overcook pasta - al dente is the Italian way
- Let ingredients shine with simple, quality-focused preparation

## Essential Italian Cooking Techniques

### Pasta Cooking
- Use plenty of salted water (1 tablespoon salt per pound of pasta)
- Cook pasta al dente - firm to the bite
- Reserve pasta water for sauce consistency
- Never rinse cooked pasta

### Sauce Making
- Start with quality olive oil and fresh garlic
- Build flavors slowly and layer ingredients
- Use fresh herbs added at the end
- Balance acidity with sweetness naturally

### Regional Variations
- Northern Italy: Rich, buttery sauces and risotto
- Central Italy: Simple, olive oil-based dishes
- Southern Italy: Spicy, tomato-heavy preparations`,
			},
		},
		Children: []DocumentSummary{
			{
				URL: "/recipes/italian/pasta-con-pommodori",
				ContentSummary: ContentSummary{
					Title:       "Pasta con Pomodori",
					Description: "A classic Italian pasta dish with fresh tomatoes, basil, and garlic. Simple yet delicious traditional recipe.",
				},
			},
			{
				URL: "/recipes/italian/pasta-carbonara",
				ContentSummary: ContentSummary{
					Title:       "Pasta Carbonara",
					Description: "Classic Roman pasta with eggs, cheese, and pancetta.",
				},
			},
			{
				URL: "/recipes/italian/pasta-puttanesca",
				ContentSummary: ContentSummary{
					Title:       "Pasta Puttanesca",
					Description: "Bold pasta with olives, capers, and anchovies.",
				},
			},
			{
				URL: "/recipes/italian/margherita-pizza",
				ContentSummary: ContentSummary{
					Title:       "Pizza Margherita",
					Description: "Traditional Neapolitan pizza with tomato, mozzarella, and basil.",
				},
			},
		},
		PrevSiblings: []DocumentSummary{
			{
				URL: "/recipes/french",
				ContentSummary: ContentSummary{
					Title:       "French Recipes",
					Description: "Classic French cuisine with sophisticated techniques and rich flavors.",
				},
			},
		},
		NextSiblings: []DocumentSummary{
			{
				URL: "/recipes/spanish",
				ContentSummary: ContentSummary{
					Title:       "Spanish Recipes",
					Description: "Vibrant Spanish dishes from paella to tapas and Mediterranean flavors.",
				},
			},
		},
		Breadcrump: []DocumentSummary{
			{
				URL: "/",
				ContentSummary: ContentSummary{
					Title:       "Lifestyle Homepage",
					Description: "Your daily source for recipes, wellness tips, home decor, and lifestyle inspiration.",
				},
			},
			{
				URL: "/recipes",
				ContentSummary: ContentSummary{
					Title:       "Recipes",
					Description: "Collection of cooking recipes",
				},
			},
		},
	}

	jsonData, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	fmt.Println(string(jsonData))
}
