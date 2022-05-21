package httpsrv

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/stn81/kate/log/ctxzap"
	"go.uber.org/zap"
)

// Error writes out an error response
func Error(ctx context.Context, w http.ResponseWriter, err interface{}) {
	errInfo, ok := err.(ErrorInfo)
	if !ok {
		errInfo = ErrServerInternal
	}

	result := &Result{
		ErrNO:  errInfo.Code(),
		ErrMsg: errInfo.Error(),
	}

	if errInfoWithData, ok := errInfo.(ErrorInfoWithData); ok {
		result.Data = errInfoWithData.Data()
	}

	if err := WriteJSON(w, result); err != nil {
		ctxzap.Extract(ctx).Error("write json response", zap.Error(err))
	}
}

// OK writes out a success response without data, used typically in an `update` api.
func OK(ctx context.Context, w http.ResponseWriter) {
	OKData(ctx, w, nil)
}

// OKData writes out a success response with data, used typically in an `get` api.
func OKData(ctx context.Context, w http.ResponseWriter, data interface{}) {
	result := &Result{
		ErrNO:  ErrSuccess.Code(),
		ErrMsg: ErrSuccess.Error(),
		Data:   data,
	}

	if err := WriteJSON(w, result); err != nil {
		ctxzap.Extract(ctx).Error("write json response", zap.Error(err))
	}
}

// EncodeJSON is a wrapper of json.Marshal()
func EncodeJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// WriteJSON writes out an object which is serialized as json.
func WriteJSON(w http.ResponseWriter, v interface{}) error {
	b, err := EncodeJSON(v)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", MIMEApplicationJSONCharsetUTF8)
	if _, err = w.Write(b); err != nil {
		return err
	}
	return nil
}

// WritePlain writes out the content as text/plain
func WritePlain(w http.ResponseWriter, content []byte) error {
	w.Header().Set("Content-Type", MIMETextPlainCharsetUTF8)
	if _, err := w.Write(content); err != nil {
		return err
	}
	return nil
}
