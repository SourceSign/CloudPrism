package common

type defaultRecipe struct {
	name        string
	ingredients []Ingredient
}

// Append implements Recipe.
func (dr *defaultRecipe) Append(ingredients ...Ingredient) {
	dr.ingredients = append(dr.ingredients, ingredients...)
}

// Ingredients implements Recipe.
func (dr *defaultRecipe) Ingredients() []Ingredient {
	return dr.ingredients
}

// Name implements Recipe.
func (dr *defaultRecipe) Name() string {
	return dr.name
}

func GetDefaultRecipe(name string) Recipe {
	return &defaultRecipe{
		name:        name,
		ingredients: make([]Ingredient, 0),
	}
}
