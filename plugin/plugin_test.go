package plugin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	"zuri.chat/zccore/utils"
)

func TestMain(m *testing.M) {
	setUp()

	exitCode := m.Run()

	tearDown()

	os.Exit(exitCode)
}

func setUp() {
	godotenv.Load("../.env")
	utils.ConnectToDB(os.Getenv("TEST_CLUSTER_URL"))
}

func tearDown() {
	coll := utils.GetCollection("plugins")
	ctx := context.TODO()
	coll.Drop(ctx)
	coll.Database().Client().Disconnect(ctx)
}

func TestRegister(t *testing.T) {
	jsonStr1 := `
{
  "name": "test plugin",
  "description": "test description",
  "developer_name": "thunder",
  "developer_email": "yourfather@baba.com",
  "template_url": "template_urly.com",
  "sidebar_url": "sidebar_url.com",
  "install_url": "install_url.com",
  "icon_url": "iconic.png"
}
   `

	jsonStr2 := `
{
  "name": "test plugin 2",
  "description": "test description 2",
  "developer_name": "thunder",
  "developer_email": "yourfather@baba.com",
  "template_url": "test2/template_url.com",
  "sidebar_url": "sidebar_url.com",
  "install_url": "install_url.com",
  "icon_url": "iconic.png"
}
   `

	t.Run("test plugin creation", func(t *testing.T) {
		w := httptest.NewRecorder()

		r, _ := http.NewRequest("POST", "/plugins/register", strings.NewReader(jsonStr1))
		Register(w, r)

		assertStatusCode(t, 201, w.Code)
	})

	t.Run("plugins cannot register same data more than once", func(t *testing.T) {
		r, _ := http.NewRequest("POST", "/plugins/register", strings.NewReader(jsonStr2))

		r2, _ := http.NewRequest("POST", "/plugins/register", strings.NewReader(jsonStr2))

		Register(httptest.NewRecorder(), r)
		w := httptest.NewRecorder()
		Register(w, r2)
		assertStatusCode(t, 403, w.Code)

		if !strings.Contains(w.Body.String(), "duplicate plugin registration") {
			t.Fail()
		}
	})
}

func assertStatusCode(t testing.TB, want, got int) {
	t.Helper()
	if got != want {
		t.Errorf("expected status code %d, but got %d", want, got)
	}
}
