Go MVC framework
================

About
-----

The mvc package provides a lightweight mvc patterned web application framework in Go. Featuring action handlers and template based views.

Installation
------------

Get the packages:

$ go get github.com/mattds/mvc

$ go get github.com/mattds/mvc/views

Usage
-----

###Imports

The below import would be used once-off to parse the view templates located in the views directory.

```go
import _ "github.com/mattds/mvc/views"
```

To use the framework, import the mvc package.

```go
import "github.com/mattds/mvc"
```
 
###Controllers

The framework provides a Controller type; defined as below:

```go
type Controller struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	Name           string
	ViewBag        map[string]interface{}
	CurrentAction  string
	ExecuteAction  bool
}
```

A user defined controller would be created as a struct with an *mvc.Controller field, e.g.

```go
type HomeController struct{ *mvc.Controller }
```

Every controller needs a constructor method with a single *mvc.Controller parameter, e.g.

```go
func NewHomeController(c *mvc.Controller) *HomeController {
	return &HomeController{c}
}
```

A new controller is created for each action call, and pre-populated as a context for the action. The user defined constructor can execute logic before each action is called, modify the context itself and cancel the execution of the original action if needed.

###Actions

An action is any parameterless void method defined on a controller, e.g.

```go
func (c *HomeController) Index() {
 	c.ViewBag["title"] = "A blog"
 	c.Render()
}
```
 
###Routing
 
By design, the mvc package does not provide custom url routing. This functionality is sufficiently catered for by the http package and external packages such as Gorilla mux. It does however provide an Action function to wrap a controller method as an http.HandlerFunc as below:

```go
http.HandleFunc("/", mvc.Action("index", NewHomeController, (*HomeController).Index))
```

###Views
 
The framework defines a View type, passed along to the templates constituting a view, defined as below.

```go
type View struct {
 	Action     string
 	Controller string
 	Name       string
 	Bag        map[string]interface{}
 	Model      interface{}
}
```
  
To define views corresponding to actions, native go templates would be created in a convention driven location. The conventional location being [view root dir]/[controller]/[view] where [view root dir] = "views" if not specifically set otherwise. Templates are shared by subfolders unless a template by the same name is defined in a subfolder which is then used instead. The primary template should by convention be named "base.html".

An example folder and template layout is presented below:

	views
	├── base.html
	├── styles.html
	├── admin
	│   └── index
	│       └── content.html
	│       └── styles.html
	├── home
	│   ├── base.html
	│   ├── contact
	│   │   └── content.html
	└── user
	    └── base.html (view for all user actions)

Views can be constructed from multiple templates, and embedded within each other, e.g. base.html may be defined as

```go
...
{{template "content.html" .}}
...
```
  
Where content.html could be as simple as:

```html
<p>Hello World</p>
```
 
Views are rendered within actions, e.g.
 
```go
c.Render() // Renders a view associated with the action of the controller.
c.RenderView(viewName) // Renders a view associated with the controller.
c.RenderViewModel(viewName,viewModel) // viewModel is assigned to the Model field of the View struct accessible from a view template.
```
 
Licence
-------

This project is released under the Apache License 2.0.