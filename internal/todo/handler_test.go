package todo

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCreateAndList(t *testing.T) {
	store := newFake()
	mux := http.NewServeMux()
	Register(mux, store)

	res := request(t, mux, http.MethodPost, "/todos", `{"title":"  Buy milk  "}`)
	if res.Code != http.StatusCreated {
		t.Fatalf("create status = %d body = %s", res.Code, res.Body.String())
	}
	var created Todo
	decode(t, res, &created)
	if created.ID != 1 || created.Title != "Buy milk" || created.Done {
		t.Fatalf("created = %+v", created)
	}

	res = request(t, mux, http.MethodGet, "/todos", "")
	if res.Code != http.StatusOK {
		t.Fatalf("list status = %d", res.Code)
	}
	var list TodoList
	decode(t, res, &list)
	if len(list.Todos) != 1 || list.Todos[0].Title != "Buy milk" {
		t.Fatalf("list = %+v", list)
	}
}

func TestCreateValidation(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, newFake())

	cases := []struct {
		body string
		want string
	}{
		{body: `{}`, want: "title is required"},
		{body: `{"title":"   "}`, want: "title is required"},
		{body: `{"title":"` + strings.Repeat("a", 201) + `"}`, want: "title must be 200 characters or fewer"},
		{body: `not-json`, want: "invalid json"},
	}
	for _, tc := range cases {
		res := request(t, mux, http.MethodPost, "/todos", tc.body)
		if res.Code != http.StatusBadRequest {
			t.Fatalf("body %q status = %d", tc.body, res.Code)
		}
		var errBody ErrorBody
		decode(t, res, &errBody)
		if errBody.Error != tc.want {
			t.Fatalf("body %q error = %q, want %q", tc.body, errBody.Error, tc.want)
		}
	}
}

func TestGetAndUpdate(t *testing.T) {
	store := newFake()
	mux := http.NewServeMux()
	Register(mux, store)
	request(t, mux, http.MethodPost, "/todos", `{"title":"Write docs"}`)

	res := request(t, mux, http.MethodGet, "/todos/1", "")
	if res.Code != http.StatusOK {
		t.Fatalf("get status = %d", res.Code)
	}

	res = request(t, mux, http.MethodPost, "/todos/1", `{"done":true}`)
	if res.Code != http.StatusOK {
		t.Fatalf("update status = %d body = %s", res.Code, res.Body.String())
	}
	var updated Todo
	decode(t, res, &updated)
	if !updated.Done || updated.Title != "Write docs" {
		t.Fatalf("updated = %+v", updated)
	}

	res = request(t, mux, http.MethodGet, "/todos/99", "")
	if res.Code != http.StatusNotFound {
		t.Fatalf("missing status = %d", res.Code)
	}
	res = request(t, mux, http.MethodGet, "/todos/nope", "")
	if res.Code != http.StatusBadRequest {
		t.Fatalf("bad id status = %d", res.Code)
	}
	res = request(t, mux, http.MethodPost, "/todos/1", `{}`)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("empty update status = %d", res.Code)
	}
}

func TestHealth(t *testing.T) {
	store := newFake()
	mux := http.NewServeMux()
	Register(mux, store)

	res := request(t, mux, http.MethodGet, "/health", "")
	if res.Code != http.StatusOK {
		t.Fatalf("health status = %d", res.Code)
	}

	store.pingErr = errors.New("down")
	res = request(t, mux, http.MethodGet, "/health", "")
	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("unhealthy status = %d", res.Code)
	}
}

func request(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	return res
}

func decode(t *testing.T, res *httptest.ResponseRecorder, dst any) {
	t.Helper()
	if err := json.Unmarshal(res.Body.Bytes(), dst); err != nil {
		t.Fatalf("decode %s: %v", res.Body.String(), err)
	}
}

type fakeStore struct {
	next    int64
	items   map[int64]Todo
	pingErr error
}

func newFake() *fakeStore {
	return &fakeStore{items: map[int64]Todo{}}
}

func (f *fakeStore) Ping(context.Context) error { return f.pingErr }

func (f *fakeStore) List(context.Context) ([]Todo, error) {
	out := []Todo{}
	for id := int64(1); id <= f.next; id++ {
		if item, ok := f.items[id]; ok {
			out = append(out, item)
		}
	}
	return out, nil
}

func (f *fakeStore) Create(_ context.Context, title string) (Todo, error) {
	f.next++
	item := Todo{ID: f.next, Title: title, CreatedAt: time.Unix(0, 0).UTC()}
	f.items[item.ID] = item
	return item, nil
}

func (f *fakeStore) Get(_ context.Context, id int64) (Todo, error) {
	item, ok := f.items[id]
	if !ok {
		return Todo{}, ErrNotFound
	}
	return item, nil
}

func (f *fakeStore) Update(_ context.Context, id int64, in UpdateTodoRequest) (Todo, error) {
	item, ok := f.items[id]
	if !ok {
		return Todo{}, ErrNotFound
	}
	if in.Title != nil {
		item.Title = *in.Title
	}
	if in.Done != nil {
		item.Done = *in.Done
	}
	f.items[id] = item
	return item, nil
}
