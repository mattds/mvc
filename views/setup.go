package views

import "github.com/mattds/mvc"

func init() {
	err := mvc.SetupViews("views")
	
	if err != nil {
		panic(err)
	}
}