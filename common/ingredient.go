package common

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type IngredientDependency func() interface{}

type Ingredient interface {
	Upsert(ctx *pulumi.Context) error
	Result() interface{}
}
