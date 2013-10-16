// Package mvc provides a simple framework for implementing the mvc pattern.
package mvc

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"
	"reflect"
	"strings"
)

// Controller provides a base type, from which a user defined controller would extend.
type Controller struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	Name           string
	ViewBag        map[string]interface{}
	CurrentAction  string
	ExecuteAction  bool
}

// View is a type pre-populated by this framework, with values accessible within views.
type View struct {
	Action     string
	Controller string
	Name       string
	Bag        map[string]interface{}
	Model      interface{}
}

// IsAction is a helper method, callable on the View instance passed into a view template.
// This provides a way to check if a view was rendered by a particular action.
func (v *View) IsAction(action string) bool {
	return v.Action == action
}

// IsController is a helper method, callable on the View instance passed into a view template.
// This provides a way to check if a view was rendered from a particular controller.
func (v *View) IsController(controller string) bool {
	return v.Controller == controller
}

// IsActionOfController is a helper method, callable on the View instance passed into a view template.
// This provides a way to check if a view was rendered by a particular action from a particular controller.
func (v *View) IsActionOfController(action, controller string) bool {
	return v.Action == action && v.Controller == controller
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
func NewController(w http.ResponseWriter, r *http.Request, name, action string) *Controller {
	return &Controller{w, r, name, make(map[string]interface{}), action, true}
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

func render(w http.ResponseWriter, name string, vm interface{}) {
	t, ok := templates[name]

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
	v := &View{c.CurrentAction, c.Name, view, c.ViewBag, viewModel}

	render(c.ResponseWriter, fmt.Sprintf("%s/%s/%s", viewRootDir, c.Name, view), v)
}

// Render by convention uses the path "[view root dir]/[controller]/[view]" to lookup
// a view to render. A view is rendered by executing the base.html template
// associated with that view.
func (c *Controller) RenderView(view string) {
	c.RenderViewModel(view, nil)
}

// Render renders the view of the current action
func (c *Controller) Render() {
	c.RenderView(c.CurrentAction)
}

var actionChecked map[string]bool = make(map[string]bool)

var controllerChecked map[string]bool = make(map[string]bool)

var controllerNames map[string]string = make(map[string]string)

// Action can be used to wrap a controller method as an http.HandlerFunc.
// The controller parameter should be of the type: func (*Controller) *SomeCustomController
// The action parameter should be of the type: func (*SomeCustomController).
// Any void method with no parameters on *SomeCustomController would be suitable to pass as the action.
func Action(name string, controller interface{}, action interface{}) http.HandlerFunc {
	actionFn := reflect.ValueOf(action)
	controllerFn := reflect.ValueOf(controller)

	actionFnType := actionFn.Type()
	controllerFnType := controllerFn.Type()

	actionFnName := actionFnType.String()
	controllerFnName := controllerFnType.String()

	if !actionChecked[actionFnName] {

		if actionFnType.Kind() != reflect.Func {
			panic("Action should be a function")
		}

		if actionFnType.NumIn() != 1 || actionFnType.NumOut() != 0 {
			panic("The signature of the action function given is not supported")
		}

		actionChecked[actionFnName] = true
	}

	if !controllerChecked[controllerFnName] {

		if controllerFnType.Kind() != reflect.Func {
			panic("Controller should be a function")
		}

		if controllerFnType.NumIn() != 1 || controllerFnType.NumOut() != 1 || controllerFnType.In(0) != reflect.TypeOf(new(Controller)) {
			panic("The signature of the controller function given is not supported")
		}

		controllerChecked[controllerFnName] = true
	}

	if controllerFnType.Out(0) != actionFnType.In(0) {
		panic("The output of the controller function is incompatible with the specified action function.")
	}

	controllerTypeName := controllerFnType.Out(0).String()

	controllerName, ok := controllerNames[controllerTypeName]

	if !ok {
		controllerName = controllerTypeName[strings.LastIndex(controllerTypeName, ".")+1:]
		controllerName = strings.Replace(strings.ToLower(controllerName), "controller", "", 1)

		controllerNames[controllerTypeName] = controllerName
	}

	return func(w http.ResponseWriter, r *http.Request) {
		controller := NewController(w, r, controllerName, name)

		childController := controllerFn.Call([]reflect.Value{reflect.ValueOf(controller)})

		if controller.ExecuteAction {
			actionFn.Call(childController)
		}
	}
}
