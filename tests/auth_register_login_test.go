package tests

//
//import (
//	sso "app/pkg/http/grpc-server"
//	"app/tests/suite"
//	"github.com/brianvoe/gofakeit/v6"
//	"github.com/golang-jwt/jwt/v5"
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/require"
//	"testing"
//	"time"
//)
//
//const (
//	emptyAppID = 0
//	appID      = 1
//	appSecret  = "tests-secret"
//
//	passDefaultLen = 10
//)
//
//// TODO: add token fail validation cases
//
//func TestRegisterLogin_Login_HappyPath(t *testing.T) {
//	ctx, st := suite.New(t)
//
//	email := gofakeit.Email()
//	pass := randomFakePassword()
//
//	respReg, err := st.AuthClient.Register(ctx, &sso.RegisterRequest{
//		Username: email,
//		Password: pass,
//	})
//	require.NoError(t, err)
//	assert.NotEmpty(t, respReg.GetUserId())
//
//	respLogin, err := st.AuthClient.Login(ctx, &sso.LoginRequest{
//		Username: email,
//		Password: pass,
//		AppId:    appID,
//	})
//	require.NoError(t, err)
//
//	token := respLogin.GetToken()
//	require.NotEmpty(t, token)
//
//	loginTime := time.Now()
//
//	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
//		return []byte(appSecret), nil
//	})
//	require.NoError(t, err)
//
//	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
//	require.True(t, ok)
//
//	assert.Equal(t, respReg.GetUserId(), int64(claims["uid"].(float64)))
//	assert.Equal(t, email, claims["email"].(string))
//	assert.Equal(t, appID, int(claims["app_id"].(float64)))
//
//	const deltaSeconds = 1
//
//	// check if exp of token is in correct range, ttl get from st.Cfg.TokenTTL
//	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), claims["exp"].(float64), deltaSeconds)
//}
//
//func TestRegisterLogin_DuplicatedRegistration(t *testing.T) {
//	ctx, st := suite.New(t)
//
//	email := gofakeit.Email()
//	pass := randomFakePassword()
//
//	respReg, err := st.AuthClient.Register(ctx, &sso.RegisterRequest{
//		Username: email,
//		Password: pass,
//	})
//	require.NoError(t, err)
//	require.NotEmpty(t, respReg.GetUserId())
//
//	respReg, err = st.AuthClient.Register(ctx, &sso.RegisterRequest{
//		Username: email,
//		Password: pass,
//	})
//	require.Error(t, err)
//	assert.Empty(t, respReg.GetUserId())
//	assert.ErrorContains(t, err, "user already exists")
//}
//
//func TestRegister_FailCases(t *testing.T) {
//	ctx, st := suite.New(t)
//
//	tests := []struct {
//		name        string
//		username    string
//		password    string
//		expectedErr string
//	}{
//		{
//			name:        "Register with Empty Password",
//			username:    gofakeit.Email(),
//			password:    "",
//			expectedErr: "password is required",
//		},
//		{
//			name:        "Register with Empty Email",
//			username:    "",
//			password:    randomFakePassword(),
//			expectedErr: "email is required",
//		},
//		{
//			name:        "Register with Both Empty",
//			username:    "",
//			password:    "",
//			expectedErr: "email is required",
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			_, err := st.AuthClient.Register(ctx, &sso.RegisterRequest{
//				Username: tt.username,
//				Password: tt.password,
//			})
//			require.Error(t, err)
//			require.Contains(t, err.Error(), tt.expectedErr)
//
//		})
//	}
//}
//
//func TestLogin_FailCases(t *testing.T) {
//	ctx, st := suite.New(t)
//
//	tests := []struct {
//		name        string
//		username    string
//		password    string
//		appID       int32
//		expectedErr string
//	}{
//		{
//			name:        "Login with Empty Password",
//			username:    gofakeit.Email(),
//			password:    "",
//			appID:       appID,
//			expectedErr: "password is required",
//		},
//		{
//			name:        "Login with Empty Email",
//			username:    "",
//			password:    randomFakePassword(),
//			appID:       appID,
//			expectedErr: "email is required",
//		},
//		{
//			name:        "Login with Both Empty Email and Password",
//			username:    "",
//			password:    "",
//			appID:       appID,
//			expectedErr: "email is required",
//		},
//		{
//			name:        "Login with Non-Matching Password",
//			username:    gofakeit.Email(),
//			password:    randomFakePassword(),
//			appID:       appID,
//			expectedErr: "invalid email or password",
//		},
//		{
//			name:        "Login without AppID",
//			username:    gofakeit.Email(),
//			password:    randomFakePassword(),
//			appID:       emptyAppID,
//			expectedErr: "app_id is required",
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			_, err := st.AuthClient.Register(ctx, &sso.RegisterRequest{
//				Username: gofakeit.Email(),
//				Password: randomFakePassword(),
//			})
//			require.NoError(t, err)
//
//			_, err = st.AuthClient.Login(ctx, &sso.LoginRequest{
//				Username: tt.username,
//				Password: tt.password,
//				AppId:    tt.appID,
//			})
//			require.Error(t, err)
//			require.Contains(t, err.Error(), tt.expectedErr)
//		})
//	}
//}
//
//func randomFakePassword() string {
//	return gofakeit.Password(true, true, true, true, false, passDefaultLen)
//}
//*/
