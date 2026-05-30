package response

type Response[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Data    any    `json:"data"`
	TraceID string `json:"trace_id,omitempty"`
}

func Success[T any](data T) Response[T] {
	return Response[T]{
		Code: 200,
		Msg:  "请求成功",
		Data: data,
	}
}

func Error(code int, msg string, traceID string) ErrorResponse {
	return ErrorResponse{
		Code:    code,
		Msg:     msg,
		Data:    nil,
		TraceID: traceID,
	}
}
