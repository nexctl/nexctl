package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

// 保证 DELETE /api/v1/nodes/{id} 在 chi 子路由树下能匹配（避免仅表现为「删除节点 404」）。
func TestChi_DeleteNodesByIDRoute(t *testing.T) {
	r := chi.NewRouter()
	r.Route("/api/v1", func(api chi.Router) {
		api.Route("/nodes", func(nr chi.Router) {
			nr.Delete("/{nodeID}", func(w http.ResponseWriter, req *http.Request) {
				if chi.URLParam(req, "nodeID") != "42" {
					t.Errorf("nodeID = %q", chi.URLParam(req, "nodeID"))
				}
				w.WriteHeader(http.StatusNoContent)
			})
		})
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/nodes/42", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("DELETE /api/v1/nodes/42: status %d", rr.Code)
	}
}
