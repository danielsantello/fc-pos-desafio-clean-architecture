package web

import (
	"encoding/json"
	"net/http"

	"github.com/danielsantello/fc-pos-desafio-clean-architecture/internal/usecase"
)

type WebListOrdersHandler struct {
	ListOrdersUseCase *usecase.ListOrdersUseCase
}

func NewWebListOrdersHandler(listOrdersUseCase *usecase.ListOrdersUseCase) *WebListOrdersHandler {
	return &WebListOrdersHandler{
		ListOrdersUseCase: listOrdersUseCase,
	}
}

func (h *WebListOrdersHandler) List(w http.ResponseWriter, r *http.Request) {
	output, err := h.ListOrdersUseCase.Execute()
	if err != nil {
		http.Error(w, "failed to list orders", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(output); err != nil {
		return
	}
}
