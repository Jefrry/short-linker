package middleware

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"net/http"

	"short-linker/internal/model"
)

func AuthMiddleware(jwtSecret []byte) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString, err := getTokenFromRequest(r)
			if err != nil || tokenString == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return jwtSecret, nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			userIDFloat, ok := claims[string(model.JWTUserIDKey)].(float64)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			userID := int64(userIDFloat)

			ctx := WithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func OptionalAuthMiddleware(jwtSecret []byte) func(next http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            tokenString, _ := getTokenFromRequest(r)
            if tokenString == "" {
                next.ServeHTTP(w, r)
                return
            }

            token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
                if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                    return nil, jwt.ErrSignatureInvalid
                }
                return jwtSecret, nil
            })
            if err != nil || !token.Valid {
                next.ServeHTTP(w, r)
                return
            }

            claims, ok := token.Claims.(jwt.MapClaims)
            if !ok {
                next.ServeHTTP(w, r)
                return
            }

            userIDFloat, ok := claims[string(model.JWTUserIDKey)].(float64)
            if !ok {
                next.ServeHTTP(w, r)
                return
            }
            userID := int64(userIDFloat)

            ctx := WithUserID(r.Context(), userID)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func WithUserID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, model.JWTUserIDKey, id)
}

func GetUserID(ctx context.Context) (int64, bool) {
	v := ctx.Value(model.JWTUserIDKey)
	if v == nil {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}

func getTokenFromRequest(r *http.Request) (string, error) {
	if c, err := r.Cookie("session_token"); err == nil && c.Value != "" {
		return c.Value, nil
	}

	return "", nil
}
