package middleware

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-gonic/gin"
)

// CustomClaims contains custom data we want from the token.
type CustomClaims struct {
	Scope string `json:"scope"`
}

// Validate does nothing for this example, but we need
// it to satisfy validator.CustomClaims interface.
func (c CustomClaims) Validate(ctx context.Context) error {
	return nil
}

// EnsureValidToken は、JWTの有効性を確認するGinミドルウェアです
func EnsureValidToken() gin.HandlerFunc {
	issuerURL, err := url.Parse("https://" + os.Getenv("AUTH0_DOMAIN") + "/")
	if err != nil {
		log.Fatalf("Failed to parse the issuer url: %v", err)
	}

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{os.Getenv("AUTH0_AUDIENCE")},
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &CustomClaims{}
			},
		),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		log.Fatalf("Failed to set up the jwt validator")
	}

	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Encountered error while validating JWT: %v", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Failed to validate JWT."}`))
	}

	middleware := jwtmiddleware.New(
		jwtValidator.ValidateToken,
		jwtmiddleware.WithErrorHandler(errorHandler),
	)

	return func(c *gin.Context) {
		encounteredError := true
		var handler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
			encounteredError = false
			// JWTが有効な場合は、クレーム情報をコンテキストから取得
			claims := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)

			// Ginのコンテキストにクレーム情報を設定
			c.Set("claims", claims)

			// カスタムクレームにアクセス
			customClaims, ok := claims.CustomClaims.(*CustomClaims)
			if ok {
				c.Set("scope", customClaims.Scope)
			}

			// 次のハンドラに進む
			c.Next()
		}

		// JWT検証を実行
		middleware.CheckJWT(http.HandlerFunc(handler)).ServeHTTP(c.Writer, c.Request)

		// エラーが発生した場合、そこで処理を中断
		if encounteredError {
			c.Abort()
		}
	}
}

// HasScope は特定のスコープ権限があるか確認するミドルウェア
func HasScope(requiredScope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// クレーム情報の取得
		scope, exists := c.Get("scope")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": "必要な権限がありません",
			})
			return
		}

		// スコープの確認（実際には複数のスコープをスペース区切りで含む場合もあるため、
		// より複雑なチェックが必要になる場合があります）
		if scope != requiredScope {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message":  "必要なスコープがありません",
				"required": requiredScope,
			})
			return
		}

		c.Next()
	}
}
