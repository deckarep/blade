/*
Open Source Initiative OSI - The MIT License (MIT):Licensing
The MIT License (MIT)
Copyright (c) 2017 Ralph Caraveo (deckarep@gmail.com)
Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package recipe

func NewRecipe() *BladeRecipe {
	return &BladeRecipe{
		Required: &RequiredRecipe{},
		Overrides: &OverridesRecipe{
			Port: 22,
		},
		Help:        &HelpRecipe{},
		Interaction: &InteractionRecipe{},
		Resilience:  &ResilienceRecipe{},
		Meta:        &MetaRecipe{},
	}
}

// // StepRecipe is an ordered series of recipes that will be attempted in the specified order.
// // The parameters specified in this recipe supercede the parameters in the individual recipe.
// type StepRecipe struct {
// 	// Hmmm...does a step recipe inherit the properties above? Or does it have it's own similar specialized properties.
// 	Recipe
// 	Steps             []*Recipe
// 	StepPauseDuration string
// }

// type AggregateRecipe struct {
// 	Recipe
// }

// BladeRecipe is the root recipe type.
type BladeRecipe struct {
	Required    *RequiredRecipe
	Overrides   *OverridesRecipe
	Help        *HelpRecipe
	Interaction *InteractionRecipe
	Resilience  *ResilienceRecipe
	Meta        *MetaRecipe
}
