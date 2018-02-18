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

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v1"
)

func LoadRecipeYaml(path string) (*BladeRecipeYaml, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		// TODO: errors.Wrap
		return nil, err
	}

	var rec BladeRecipeYaml
	err = yaml.Unmarshal(b, &rec)
	if err != nil {
		return nil, err
	}

	// Each argument needs to capture it's arg name for later processing.
	if len(rec.Args) > 0 {
		for argName, argVal := range rec.Args {
			argVal.argName = argName
		}
	}

	if rec.Help == nil {
		rec.Help = &BladeRecipeHelp{}
	}

	if rec.Overrides == nil {
		rec.Overrides = &BladeRecipeOverrides{}
	}

	if rec.Resilience == nil {
		rec.Resilience = &BladeRecipeResilience{}
	}

	return &rec, nil
}
