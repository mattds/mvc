/*
Copyright 2013 Matt Stephanou

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package mvc provides a simple framework for implementing the mvc pattern.
package mvc

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"
)

// Controller provides a base type, from which a user defined controller would extend.
type Controller struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	Name           string
	ViewBag        map[string]interface{}
}

// View is a type pre-populated by this framework, with values accessible within views.
type View struct {
	Controller string
	Name       string
	Bag        map[string]interface{}
	Model      interface{}
}

// IsView is a helper method, callable on the View instance passed into a view template.
// This provides a way to check the name of a view rendered.
func (v *View) IsView(name string) bool {
	return v.Name == name
}

// IsController is a helper method, callable on the View instance passed into a view template.
// This provides a way to check which controller a view was rendered from.
func (v *View) IsController(controller string) bool {
	return v.Controller == controller
}

// IsViewForController is a helper method, callable on the View instance passed into a view template.
// This provides a way to check if a view was rendered by a particular controller.
func (v *View) IsViewForController(viewName, controller string) bool {
	return v.Name == viewName && v.Controller == controller
}

var templates map[string]*template.Template

var viewRootDir string = ""

// SetupViews pre-populates the templates map with parsed view templates.
func SetupViews(rootDir string) error {
	if viewRootDir != "" {
		return errors.New("Views cannot have more than one root directory.")
	}

	templates = make(map[string]*template.Template)

	viewRootDir = rootDir

	return parseViewDirectory(viewRootDir, nil)
}

// NewController can be used to instantiate a Controller instance.
func NewController(w http.ResponseWriter, r *http.Request, name string) *Controller {
	return &Controller{w, r, name, make(map[string]interface{})}
}

// funcMap defines a set of additional functions callable within view templates.
var funcMap = template.FuncMap{
	// noescape provides a way to output text within a view which is not escaped,
	// this can be used to ouput html comments for instance.
	"noescape": func(x string) template.HTML {
		return template.HTML(x)
	},
}

// parseViewDirectory is used to recursively walk a directory and parse the templates within.
// A given folder defines a view. A view is composed of the templates stored within the
// root view folder down to the sub folder which defines the view.
// For a given view, Templates in subfolders override templates with the
// same name in a parent folder.
func parseViewDirectory(dirname string, parentViews map[string]string) error {
	views := make(map[string]string)

	if parentViews != nil {
		for k, v := range parentViews {
			views[k] = v
		}
	}

	f, err := os.Open(dirname)

	if err != nil {
		return err
	}

	defer f.Close()

	list, err := f.Readdir(-1)

	if err != nil {
		return err
	}

	for _, f := range list {

		isHtml, err := path.Match("*.html", f.Name())

		if err != nil {
			return err
		}

		if !f.IsDir() && isHtml {
			// this will override templates stored in parent views
			views[f.Name()] = path.Join(dirname, f.Name())
		}
	}

	for _, f := range list {

		if f.IsDir() {
			parseViewDirectory(path.Join(dirname, f.Name()), views)
		}
	}

	if len(views) > 0 {
		htmlTemplates := make([]string, len(views))

		i := 0

		for _, v := range views {
			htmlTemplates[i] = v
			i++
		}

		t := template.New("base.html").Funcs(funcMap)

		templates[dirname] = template.Must(t.ParseFiles(htmlTemplates...))
	}

	return nil
}

func render(w http.ResponseWriter, controllerName, view string, vm interface{}) {
	name := fmt.Sprintf("%s/%s/%s", viewRootDir, controllerName, view)

	t, ok := templates[name]

	if !ok {
		name = fmt.Sprintf("%s/%s", viewRootDir, controllerName)

		t, ok = templates[name]
	}

	if !ok {
		name = viewRootDir

		t, ok = templates[name]
	}

	if !ok {
		http.Error(w, fmt.Sprintf("The templates for %v were not found.", name), http.StatusInternalServerError)
		return
	}

	err := t.ExecuteTemplate(w, "base.html", vm)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// RenderViewModel has the same functionality as Render, as well as the ability
// to pass along a viewModel to the templates associated with the view.
func (c *Controller) RenderViewModel(view string, viewModel interface{}) {
	v := &View{c.Name, view, c.ViewBag, viewModel}

	render(c.ResponseWriter, c.Name, view, v)
}

// Render by convention uses the path "[view root dir]/[controller]/[view]" to lookup
// a view to render. A view is rendered by executing the base.html template
// associated with that view.
func (c *Controller) Render(view string) {
	c.RenderViewModel(view, nil)
}
