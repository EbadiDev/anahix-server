package admin

import (
	"github.com/EbadiDev/anahix-server/internal/svc"
	"github.com/EbadiDev/anahix-server/pkg/result"
	"github.com/EbadiDev/anahix-server/pkg/upload"
	"github.com/EbadiDev/anahix-server/pkg/xerr"
	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	svcCtx *svc.ServiceContext
}

func NewUploadHandler(svcCtx *svc.ServiceContext) *UploadHandler {
	return &UploadHandler{svcCtx: svcCtx}
}

func (h *UploadHandler) UploadMedia(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		result.ParamErrorResult(c, xerr.NewErrMsg("فایل ارسال نشده است"))
		return
	}

	resp, err := upload.SaveMediaImage(fileHeader)
	result.HttpResult(c, resp, err)
}
