Go MVC framework
================

About
-----

The mvc package provides a lightweight mvc patterned web application framework in Go. The focus of this framework is on the view aspect, providing a simple way to share templates between views.

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
	http.ResponseWriter
	Request        *http.Request
	Name           string
	ViewBag        map[string]interface{}
}
```

A user defined controller would be created as a struct with an *mvc.Controller field, e.g.

```go
type HomeController struct{ *mvc.Controller }
```

To simplify routing, a standard approach to creating a handler is suggested. The recommended approach is for every controller to define an action type of the below form and a corresponding ServeHTTP method associated with that type as the below example illustrates.

```go
type HomeControllerAction func(*HomeController)

func (action HomeControllerAction) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	action(&HomeController{mvc.NewController(w, r, "home")})
}
```

This approach creates a new controller for each action call, which provides a context for the action. This approach also gives the flexibility to execute an action conditionally or execute logic before and after any action is called in a given controller.

###Actions

An action is any parameterless void method defined on a controller, e.g.

```go
func (c *HomeController) Index() {
 	c.ViewBag["title"] = "A blog"
 	c.Render("index")
}
```
 
###Routing
 
By design, the mvc package does not provide custom url routing. This functionality is sufficiently catered for by the http package and external packages such as Gorilla mux. An example of how to handle a route via an action is given below:

```go
http.Handle("/", HomeControllerAction((*HomeController).Index))
```

###Views
 
The framework defines a View type, passed along to the templates constituting a view, defined as below.

```go
type View struct {
 	Controller string
 	Name       string
 	Bag        map[string]interface{}
 	Model      interface{}
}
```
  
Native go templates are used to define views renderable within a controller. These views should be created in a convention driven location. The conventional location being [view root dir]/[controller]/[view] where [view root dir] = "views" if not specifically set otherwise. Templates are shared by subfolders; a template with the same name in a lower level subfolder would take precedence. A primary template named "base.html" is required, whether shared between all views or specific to a particular view.

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
	    └── base.html (used by all views rendered from the user controller, individual folders per view need not be explicitly created)

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
c.Render(viewName) // Renders a view associated with the controller.
c.RenderViewModel(viewName,viewModel) // viewModel is assigned to the Model field of the View struct accessible from a view template.
```
 
Licence
-------

This project is released under the Apache License 2.0.
