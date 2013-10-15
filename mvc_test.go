package mvc

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"
)

func createTemplateFile(dir, name, content string, t *testing.T) {
	template, err := os.Create(path.Join(dir, name))

	if err != nil {
		t.Errorf(err.Error())
	} else {
		defer template.Close()
	}

	fmt.Fprintf(template, content)
}

func createFolder(dir, name string, t *testing.T) string {
	newDir := path.Join(dir, name)

	err := os.Mkdir(newDir, 0700)

	if err != nil {
		t.Errorf(err.Error())
	}

	return newDir
}

func mockController(name string) *Controller {
	w := &mockResponseWriter{}

	r, _ := http.NewRequest("GET", "/", nil)

	return NewController(w, r, name)
}

func TestViewTemplates(t *testing.T) {
	root, err := ioutil.TempDir("", "mvc_test")

	fmt.Println(root)

	if err == nil {
		defer os.RemoveAll(root)
	}

	// Template shared by all views unless overwritten

	createTemplateFile(root, "base.html", `Top: {{template "content.html" .}}`, t)

	// Home controller

	hcDir := createFolder(root, "home", t)

	// Template shared by all home controller views

	createTemplateFile(hcDir, "base.html", `a {{template "content.html" .}}`, t)

	// Home controller index action

	hcIndexActionDir := createFolder(hcDir, "index", t)

	createTemplateFile(hcIndexActionDir, "content.html", `plane`, t)

	// Home controller contact action

	hcContactActionDir := createFolder(hcDir, "contact", t)

	createTemplateFile(hcContactActionDir, "content.html", `bird`, t)

	// User Controller

	ucDir := createFolder(root, "user", t)

	// Template shared by all user controller views

	createTemplateFile(ucDir, "base.html", `Hello {{template "content.html" .}}`, t)

	// User controller index action

	ucIndexActionDir := createFolder(ucDir, "index", t)

	createTemplateFile(ucIndexActionDir, "content.html", `{{.Model}}`, t)

	// Admin Controller

	aDir := createFolder(root, "admin", t)

	// Admin controller index action

	aIndexActionDir := createFolder(aDir, "index", t)

	createTemplateFile(aIndexActionDir, "content.html", `level`, t)

	SetupViews(root)

	type testCase struct {
		controller, action, expected, viewModel string
	}

	testCases := []testCase{
		testCase{"home", "index", "a plane", ""},
		testCase{"home", "contact", "a bird", ""},
		testCase{"user", "index", "Hello everyone", "everyone"},
		testCase{"admin", "index", "Top: level", ""},
	}

	for _, tc := range testCases {
		c := mockController(tc.controller)

		if tc.viewModel != "" {
			c.RenderViewModel(tc.action, tc.viewModel)
		} else {
			c.Render(tc.action)
		}

		expectedResult := []byte(tc.expected)

		if !bytes.Equal(c.ResponseWriter.(*mockResponseWriter).Body(), expectedResult) {
			t.Errorf("Result was '%s', expected '%s'", c.ResponseWriter.(*mockResponseWriter).Body(), expectedResult)
		}
	}
}

type mockResponseWriter struct {
	buffer bytes.Buffer
}

func (w *mockResponseWriter) Header() http.Header { return make(map[string][]string) }

func (w *mockResponseWriter) Write(b []byte) (int, error) { return w.buffer.Write(b) }

func (w *mockResponseWriter) WriteHeader(int) {}

func (w *mockResponseWriter) Body() []byte { return w.buffer.Bytes() }
