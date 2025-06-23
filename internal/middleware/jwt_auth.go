package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"workout-tracker-api/internal/apperrors"
	"workout-tracker-api/internal/util/auth"
	"workout-tracker-api/internal/util/helper"
)

func JWTAuthMiddleware(tokenService auth.TokenInterface) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// extract Token from AUthorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				helper.SendErrorResponse(w, apperrors.ErrUnauthorized)
				return
			}

			tokenParts := strings.Split(authHeader, " ")

			if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
				helper.SendErrorResponse(w, apperrors.ErrUnauthorized)
				return
			}

			tokenString := tokenParts[1]
			// parse and validate token
			claims, err := tokenService.ParseToken(r.Context(), tokenString)

			if err != nil {
				log.Printf("JWT parsing/validation error: %v", err)
				helper.SendErrorResponse(w, apperrors.ErrUnauthorized)
				return
			}

			// check whether the token in balcklist
			if claims.RegisteredClaims.ID != "" {
				isBlaclisted, err := tokenService.CheckBlacklist(r.Context(), claims.ID)

				if err != nil {
					helper.SendErrorResponse(w, fmt.Errorf("error checking blacklist token: %w", err))
					return
				}

				if isBlaclisted {
					log.Printf("Attempt to use blacklisted token JTI: %s", claims.ID)
					helper.SendErrorResponse(w, apperrors.ErrUnauthorized)
					return
				}

			} else {
				log.Printf("Warning: JWT claims missing JTI for blacklisting check. Claims: %+v", claims)
				helper.SendErrorResponse(w, apperrors.ErrUnauthorized)
				return
			}
			// set user and JTI into context
			ctx := helper.SetUserInfoToContext(r.Context(), &helper.UserInfo{
				Id:    *claims.Payload.Id,
				Email: claims.Email,
				Name:  claims.Payload.Name,
			})

			ctx = helper.SetJTIToContext(ctx, &helper.JTIInfo{
				Id:             claims.ID,
				ExpirationTime: claims.ExpiresAt.Time,
			})
			r = r.WithContext(ctx)
			// call the next handler
			next.ServeHTTP(w, r)

		})
	}
}
