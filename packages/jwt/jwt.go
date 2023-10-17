package jwt

import (
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	// nuboJwt "github.com/nubo/jwt"
	"github.com/devinterop/mgdb-core/app/structs"
	cnst "github.com/devinterop/mgdb-core/cnst"
	"github.com/devinterop/mgdb-core/packages/logging"
)

type JwtCtrl struct{}

// jwt service
type JWTService interface {
	GenerateToken(email string, isUser bool) string
	ValidateToken(token string) (*jwt.Token, error)
}
type authCustomClaims struct {
	Name string `json:"name"`
	User bool   `json:"user"`
	jwt.StandardClaims
}

type jwtServices struct {
	secretKey string
	issure    string
}

var logrusFieldJwt = structs.LogrusField{
	Module: "JWTService",
}

// auth-jwt
func JWTAuthService() JWTService {
	return &jwtServices{
		secretKey: getSecretKey(),
		issure:    "Bikash",
	}
}

func getSecretKey() string {
	secret := os.Getenv(cnst.SecretKey)
	if secret == "" {
		secret = "secret"
	}
	return secret
}

func (service *jwtServices) GenerateToken(email string, isUser bool) string {
	logrusField := logrusFieldJwt
	logrusField.Method = "GenerateToken"

	claims := &authCustomClaims{
		email,
		isUser,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 48).Unix(),
			Issuer:    service.issure,
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	//encoded string
	t, err := token.SignedString([]byte(service.secretKey))
	if err != nil {
		//panic(err)
		logging.Logger(cnst.Fatal, err, logrusField)
	}
	return t
}

func (service *jwtServices) ValidateToken(encodedToken string) (*jwt.Token, error) {
	// atClaims := jwt.MapClaims{}
	// atClaims["authorized"] = true
	return jwt.Parse(encodedToken, func(token *jwt.Token) (interface{}, error) {
		if _, isvalid := token.Method.(*jwt.SigningMethodHMAC); !isvalid {
			return nil, fmt.Errorf("Invalid token %s", token.Header["alg"])

		}
		return []byte(service.secretKey), nil
	})

}

func (j JwtCtrl) ExtractClaims(tokenStr string) (jwt.MapClaims, bool) {
	logrusField := logrusFieldJwt
	logrusField.Method = "ExtractClaims"

	hmacSecretString := getSecretKey() // Value
	hmacSecret := []byte(hmacSecretString)
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {

		// check token signing method etc
		return hmacSecret, nil
	})
	//fmt.Println("Parse ] ]", err)
	if err != nil {
		logging.Logger(cnst.Fatal, fmt.Sprint("JWT Parse error: ", err), logrusField)
		return nil, false
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, true
	} else {
		//log.Printf("Invalid JWT Token")
		logging.Logger(cnst.Error, "Invalid JWT Token", logrusField)
		return nil, false
	}
}
