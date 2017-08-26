package recipe

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

func LoadRecipe(path string) (*BladeRecipe, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		// TODO: errors.Wrap
		return nil, err
	}

	rec := NewRecipe()
	_, err = toml.Decode(string(b), rec)
	if err != nil {
		// TODO: errors.Wrap
		return nil, err
	}

	// The first _ in toml.Decode is a meta property with more interesting details.
	//fmt.Println(meta.IsDefined("PromptBanner"))

	return rec, nil
}
