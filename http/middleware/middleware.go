package middleware

import (
	"LabSystem/api"
	"LabSystem/internal/httperr"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 该文件实现装饰器模式的中间件
// 中间件用于辅助处理 HTTP 请求或添加其他自定义功能
//
// 中间件围绕自定义的 HandlerFunc 而非 http.HandlerFunc 设计，
// 以便后置日志中间件可直接获取下游返回的具体错误信息。
// 中间件链最终通过 Adapt 统一适配为 http.HandlerFunc 挂到 router 上。

// HandlerFunc 自定义Handler类型，返回error由上层中间件/适配器统一处理
type HandlerFunc func(http.ResponseWriter, *http.Request) error

// contextKey 中间件与控制层之间传递身份信息的 context 键类型
type contextKey string

const (
	// CtxKeyUserID 登录用户编号（学号/工号）
	CtxKeyUserID contextKey = "userID"
	// CtxKeyRole 登录用户身份（student / teacher / operator）
	CtxKeyRole contextKey = "role"
)

// Secret JWT 签名密钥，由启动阶段注入，与签发侧共用同一值
var Secret []byte

// authCookieName 登录 JWT 所在的 cookie 名，与签发侧保持一致
const authCookieName = "auth_token"

// ServeError 统一的 HTTP 层错误处理
// 将 error 中提取的状态码与消息以 JSON 形式写入响应
func ServeError(w http.ResponseWriter, err error) {
	status := httperr.HTTPStatus(err)
	msg := "Internal Server Error"
	// 不暴露服务器内部错误
	if status < 500 {
		msg = err.Error()
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// Adapt 将 HandlerFunc 适配为 http.HandlerFunc
// 用于中间件链最外层，把返回的 error 统一交给 ServeError 处理
func Adapt(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			ServeError(w, err)
		}
	}
}

// JwtValidator 通过验证JWT，保护需要登录态才能访问的接口
// 拦截验证不通过的请求，同时将解析到的JWT值加入上下文(context)
// 验证的JWT结构为 api.LoginJWT
func JwtValidator(next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ck, err := r.Cookie(authCookieName)
		if err != nil {
			return httperr.WithStatus(
				errors.New("missing auth token"),
				http.StatusUnauthorized)
		}

		claims := &api.LoginJWT{}
		token, err := jwt.ParseWithClaims(ck.Value, claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return Secret, nil
		})
		if err != nil || token == nil || !token.Valid {
			return httperr.WithStatus(
				errors.New("invalid or expired token"),
				http.StatusUnauthorized)
		}

		ctx := context.WithValue(r.Context(), CtxKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, CtxKeyRole, claims.Role)
		return next(w, r.WithContext(ctx))
	}
}

// LoginCheck 登录态检查。
// 访问站点主页时（URI为"/"）若未携带JWT，重定向到登录页面（"/sessions"）
func LoginCheck(next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		if r.URL.Path == "/" {
			if _, err := r.Cookie(authCookieName); err != nil {
				http.Redirect(w, r, "/sessions", http.StatusFound)
				return nil
			}
		}
		return next(w, r)
	}
}

// Logger 日志中间件
// 分为前置与后置两个部分。
// 前置部分记录到达的请求信息，后置部分记录服务端的处理错误
func Logger(next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		start := time.Now()
		//FIXME: 使用文件记录
		log.Printf("[REQ] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

		err := next(w, r)

		defer func() {
			elapsed := time.Since(start)
			if err != nil {
				log.Printf("[ERR] %s %s -> %d %v (%s)",
					r.Method, r.URL.Path, httperr.HTTPStatus(err), err, elapsed)
			} else {
				log.Printf("[RES] %s %s -> ok (%s)", r.Method, r.URL.Path, elapsed)
			}
		}()

		return err
	}
}

// Recovery 捕获panic，确保服务运行
// 将 panic 转为 500 error 向上返回，由 Logger / Adapt 统一处理
func Recovery(next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (err error) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("[PANIC] %s %s: %v\n%s",
					r.Method, r.URL.Path, rec, debug.Stack())
				err = httperr.WithStatus(
					fmt.Errorf("panic: %v", rec),
					http.StatusInternalServerError)
			}
		}()
		return next(w, r)
	}
}
