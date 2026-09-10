package result

import (
	"errors"
	"net/http"

	"github.com/EbadiDev/anahix-server/pkg/xerr"
	"github.com/gin-gonic/gin"
)

func HttpResult(ctx *gin.Context, resp interface{}, err error) {
	if err == nil {
		ctx.JSON(http.StatusOK, Success(resp))
		return
	}

	code := xerr.ERROR
	msg := xerr.MapErrMsg(xerr.ERROR)

	var e *xerr.CodeError
	if errors.As(err, &e) {
		code = e.GetErrCode()
		msg = e.GetErrMsg()
	} else {
		msg = err.Error()
	}

	ctx.JSON(http.StatusOK, Error(code, msg))
}

func ParamErrorResult(ctx *gin.Context, err error) {
	errMsg := err.Error()
	ctx.JSON(http.StatusOK, Error(xerr.InvalidParams, errMsg))
}
