package handler

import (
	"fmt"
	"log"
	"net/http"
	"workout-tracker-api/internal/apperrors"
	"workout-tracker-api/internal/service"
	"workout-tracker-api/internal/util"
	"workout-tracker-api/internal/util/helper"
	"workout-tracker-api/pkg/api"
)

type ReportHandler struct {
	ReportService service.ReportServiceInterface
}

func NewReportHandler(rs service.ReportServiceInterface) *ReportHandler {
	return &ReportHandler{
		ReportService: rs,
	}
}

func (rc *ReportHandler) ReportProgress(w http.ResponseWriter, r *http.Request) {
	userInfo, ok := helper.GetUserInfoFromContext(r.Context())

	if !ok {
		log.Printf("Failed to get user info from context")
		helper.SendErrorResponse(w, apperrors.ErrUnauthorized)
		return

	}

	serviceProgress, err := rc.ReportService.Progress(r.Context(), userInfo.Id)

	if err != nil {

		helper.SendErrorResponse(w, fmt.Errorf("failed to fetch progress: %w", err))
		return
	}

	var progress = toAPIProgress(serviceProgress)

	response := api.Success{
		Code:    api.FETCH,
		Message: "successsfully fetch progress report",
		Payload: &map[string]interface{}{
			"progress": progress,
		},
	}

	helper.SendSuccessResponse(w, http.StatusOK, &response)

}

func toAPIProgress(progress *service.ProgressStatus) *api.Progress {
	if progress == nil {
		return nil
	}
	return &api.Progress{
		CompletedWorkouts: util.IntTo64(progress.CompleteWorkouts),
		TotalWorkouts:     util.IntTo64(progress.TotalWorkouts),
	}
}
