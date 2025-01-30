package rest

import (
	"context"
	"fmt"
	"github.com/ardianferdianto/reconciliation-service/internal/domain"
	"github.com/ardianferdianto/reconciliation-service/internal/usecase/reconcile"
	"github.com/ardianferdianto/reconciliation-service/internal/usecase/workflow"
	"github.com/ardianferdianto/reconciliation-service/pkg/contract"
	"github.com/ardianferdianto/reconciliation-service/pkg/request"
	"github.com/ardianferdianto/reconciliation-service/pkg/response"
	"github.com/gorilla/mux"
	"net/http"
)

type WorkflowHandler struct {
	workflowUC  workflow.IUseCase
	reconcileUC reconcile.IUseCase
}

func NewWorkflowHandler(workflowUC workflow.IUseCase, reconcileUC reconcile.IUseCase) *WorkflowHandler {
	return &WorkflowHandler{workflowUC: workflowUC, reconcileUC: reconcileUC}
}

func (h *WorkflowHandler) StartWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	var req contract.StartWorkflowRequest

	if err := request.ReadJSON(r, &req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if req.SystemTransactionFilePath == "" || len(req.BankStatementFilePaths) == 0 {
		http.Error(w, "Missing required file paths", http.StatusBadRequest)
		return
	}
	if req.StartDate.After(req.EndDate) {
		http.Error(w, "Start date must be before end date", http.StatusBadRequest)
		return
	}

	workflowID, err := h.workflowUC.StartWorkflow(
		context.Background(),
		req.SystemTransactionFilePath, // Bucket name for system transactions
		req.BankStatementFilePaths,    // Assuming same bucket for bank statements
		req.StartDate,
		req.EndDate,
	)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to start workflow: %v", err), http.StatusInternalServerError)
		return
	}

	statusCode := http.StatusOK
	responseBody := map[string]string{"workflow_id": workflowID}
	response.WriteJSON(r.Context(), w, statusCode, responseBody)
}

func (h *WorkflowHandler) GetWorkflowSummary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workflowID := vars["workflowID"]

	ctx := r.Context()
	wf, err := h.workflowUC.GetWorkflowSummary(ctx, workflowID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to retrieve workflow summary: %v", err), http.StatusInternalServerError)
		return
	}
	var rec domain.ReconciliationSummary
	if wf.ReconciliationJobID != nil {
		rec, err = h.reconcileUC.GetReconciliationSummary(ctx, *wf.ReconciliationJobID)
	}

	resp := contract.WorkflowSummaryResponse{
		WorkflowID:       wf.WorkflowID,
		Status:           wf.Status,
		StartDate:        wf.StartDate,
		EndDate:          wf.EndDate,
		ReconcileSummary: &rec,
	}

	w.Header().Set("Content-Type", "application/json")
	statusCode := http.StatusOK
	response.WriteJSON(r.Context(), w, statusCode, resp)
}
