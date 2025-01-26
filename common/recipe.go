package common

type Recipe interface {
	Name() string
	Ingredients() []Ingredient
	Append(ingredients ...Ingredient)
}
