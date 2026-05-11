package shortcut

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetStory(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/stories/12345", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Shortcut-Token") != "tok" {
			t.Errorf("missing token header")
		}
		_, _ = w.Write([]byte(`{"id":12345,"name":"the story","epic_id":4242,"app_url":"https://app.shortcut.com/x/story/12345"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := New("tok")
	c.BaseURL = srv.URL

	s, err := c.GetStory(context.Background(), 12345)
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "the story" {
		t.Errorf("name = %q", s.Name)
	}
	if s.EpicID == nil || *s.EpicID != 4242 {
		t.Errorf("epic id = %v", s.EpicID)
	}
}

func TestGetEpic(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/epics/4242", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":4242,"name":"the epic","milestone_id":99}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := New("tok")
	c.BaseURL = srv.URL

	e, err := c.GetEpic(context.Background(), 4242)
	if err != nil {
		t.Fatal(err)
	}
	if e.Name != "the epic" {
		t.Errorf("name = %q", e.Name)
	}
}

func TestGetStoryNotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/stories/1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := New("tok")
	c.BaseURL = srv.URL

	if _, err := c.GetStory(context.Background(), 1); err == nil {
		t.Errorf("expected error")
	}
}

func TestGetWorkflows(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/workflows", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"id":1,"name":"Eng","states":[{"id":100,"name":"Backlog"},{"id":101,"name":"In Dev"}]}]`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := New("tok")
	c.BaseURL = srv.URL

	ws, err := c.GetWorkflows(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(ws) != 1 || len(ws[0].States) != 2 || ws[0].States[1].Name != "In Dev" {
		t.Errorf("got %+v", ws)
	}
}

func TestGetEpicWorkflow(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/epic-workflow", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"epic_states":[{"id":1,"name":"To Do"},{"id":2,"name":"Done"}]}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := New("tok")
	c.BaseURL = srv.URL

	ew, err := c.GetEpicWorkflow(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(ew.EpicStates) != 2 || ew.EpicStates[0].Name != "To Do" {
		t.Errorf("got %+v", ew)
	}
}

func TestGetObjective(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/objectives/99", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":99,"name":"the objective"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := New("tok")
	c.BaseURL = srv.URL

	o, err := c.GetObjective(context.Background(), 99)
	if err != nil {
		t.Fatal(err)
	}
	if o.Name != "the objective" {
		t.Errorf("name = %q", o.Name)
	}
}
