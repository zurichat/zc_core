package plugin

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestRegister(t *testing.T) {
	jsonData := `{
  "name": "test plugin",
  "description": "test description",
  "developer_name": "thunder",
  "developer_email": "yourfather@baba.com",
  "template_url": "template_urly.com",
  "sidebar_url": "sidebar_url.com",
  "install_url": "install_url.com",
  "icon_url": "iconic.png"
}`
	t.Run("test successful plugin creation", func(t *testing.T) {
		ts := &testService{}
		ph := NewHandler(ts)
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/plugins/register", strings.NewReader(jsonData))

		ph.Register(w, r)

		assertStatusCode(t, 201, w.Code)
	})

	t.Run("plugins cannot register same data more than once", func(t *testing.T) {
		r1, _ := http.NewRequest("POST", "/plugins/register", strings.NewReader(jsonData))
		r2, _ := http.NewRequest("POST", "/plugins/register", strings.NewReader(jsonData))
		ts := &testService{}
		ph := NewHandler(ts)

		ph.Register(httptest.NewRecorder(), r1)
		w := httptest.NewRecorder()
		ph.Register(w, r2)

		assertStatusCode(t, 403, w.Code)

		if !strings.Contains(w.Body.String(), "plugin exists") {
			t.Fail()
		}
	})
}

func TestUpdate(t *testing.T) {
	jsonData := `{
"name": "changed name",
"description": "changed description"
}`
	t.Run("test successful plugin update", func(t *testing.T) {
		store := []*Plugin{
			{
				ID:          primitive.NewObjectID(),
				Name:        "old name",
				Description: "old description",
			},
		}
		ts := &testService{store}
		ph := NewHandler(ts)
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("PATCH", fmt.Sprintf("/plugins/%s", store[0].ID.Hex()), strings.NewReader(jsonData))

		ph.Update(w, r)

		assertStatusCode(t, 200, w.Code)
		assertStringsEqual(t, store[0].Name, "changed name")
		assertStringsEqual(t, store[0].Description, "changed description")
	})
}

/*
func TestMain(m *testing.M) {
	setUp()

	exitCode := m.Run()

	tearDown()

	os.Exit(exitCode)
}

func setUp() {
}

func tearDown() {
}
*/
