package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type registerResponse struct {
	User struct {
		ID        int64  `json:"id"`
		Username  string `json:"username"`
		Email     string `json:"email"`
		Activated bool   `json:"activated"`
	} `json:"user"`
	ActivationToken string `json:"activation_token"`
}

type accessTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type securedResponse struct {
	UserID int64    `json:"userId"`
	Roles  []string `json:"roles"`
}

type userDetailsResponse struct {
	UserDetails struct {
		ID    int64    `json:"id"`
		Email string   `json:"email"`
		Roles []string `json:"roles"`
	} `json:"user_details"`
}

type errorStringResponse struct {
	Error string `json:"error"`
}

type errorMapResponse struct {
	Error map[string]string `json:"error"`
}

var credentialCounter uint64

func TestAcceptanceIdentity_UserSessionLifecycle(t *testing.T) {
	host := startIdentityServer(t)
	client := &http.Client{}
	username, email, password := uniqueCredentials()

	// Step 1: Register a new user and capture the activation token.
	registerRes := doRequest(t, client, http.MethodPost, host+"/register", map[string]string{
		"username": username,
		"email":    email,
		"password": password,
	}, "", "")
	assertStatus(t, registerRes, http.StatusAccepted)
	registered := decodeJSONBody[registerResponse](t, registerRes)

	if registered.User.ID <= 0 {
		t.Fatal("expected a persisted user id after register")
	}
	if registered.User.Activated {
		t.Fatal("expected newly registered user to be inactive")
	}
	if len(registered.ActivationToken) != 26 {
		t.Fatalf("expected activation token length 26, got %d", len(registered.ActivationToken))
	}

	// Step 2: Activate the user and capture both access and refresh tokens.
	activateRes := doRequest(t, client, http.MethodPut, host+"/activate", map[string]string{
		"activationToken": registered.ActivationToken,
	}, "", "")
	assertStatus(t, activateRes, http.StatusOK)
	activated := decodeJSONBody[accessTokenResponse](t, activateRes)
	refreshToken := cookieValue(t, activateRes, "refresh_token")

	if activated.AccessToken == "" {
		t.Fatal("expected access token on activation")
	}
	if len(refreshToken) != 26 {
		t.Fatalf("expected refresh token length 26, got %d", len(refreshToken))
	}

	// Step 3: Access a secured endpoint with the activation-issued access token.
	securedRes := doRequest(t, client, http.MethodGet, host+"/secured", nil, activated.AccessToken, "")
	assertStatus(t, securedRes, http.StatusOK)
	secured := decodeJSONBody[securedResponse](t, securedRes)

	if secured.UserID != registered.User.ID {
		t.Fatalf("expected secured userId %d, got %d", registered.User.ID, secured.UserID)
	}
	if !slices.Contains(secured.Roles, "user") {
		t.Fatalf("expected role list to include user, got %v", secured.Roles)
	}

	// Step 4: Refresh tokens using the refresh cookie value.
	refreshRes := doRequest(t, client, http.MethodPost, host+"/refresh", nil, "", refreshToken)
	assertStatus(t, refreshRes, http.StatusOK)
	refreshed := decodeJSONBody[accessTokenResponse](t, refreshRes)
	rotatedRefreshToken := cookieValue(t, refreshRes, "refresh_token")

	if refreshed.AccessToken == "" {
		t.Fatal("expected access token on refresh")
	}
	if len(rotatedRefreshToken) != 26 {
		t.Fatalf("expected rotated refresh token length 26, got %d", len(rotatedRefreshToken))
	}

	// Step 5: Verify the refreshed access token is accepted by the secured endpoint.
	securedAfterRefreshRes := doRequest(t, client, http.MethodGet, host+"/secured", nil, refreshed.AccessToken, "")
	assertStatus(t, securedAfterRefreshRes, http.StatusOK)
	securedAfterRefresh := decodeJSONBody[securedResponse](t, securedAfterRefreshRes)

	if securedAfterRefresh.UserID != registered.User.ID {
		t.Fatalf("expected secured userId %d after refresh, got %d", registered.User.ID, securedAfterRefresh.UserID)
	}

	// Step 6: Logout with a valid access token.
	logoutRes := doRequest(t, client, http.MethodPost, host+"/logout", nil, refreshed.AccessToken, "")
	assertStatus(t, logoutRes, http.StatusNoContent)
	clearedRefreshToken := cookieValue(t, logoutRes, "refresh_token")
	if err := logoutRes.Body.Close(); err != nil {
		t.Fatalf("could not close response body: %v", err)
	}

	if clearedRefreshToken != "" {
		t.Fatalf("expected logout to clear refresh cookie, got %q", clearedRefreshToken)
	}

	// Step 7: Verify old refresh token can no longer mint access tokens after logout.
	refreshAfterLogoutRes := doRequest(t, client, http.MethodPost, host+"/refresh", nil, "", rotatedRefreshToken)
	assertStatus(t, refreshAfterLogoutRes, http.StatusBadRequest)
	refreshAfterLogoutBody := decodeJSONBody[errorStringResponse](t, refreshAfterLogoutRes)

	if refreshAfterLogoutBody.Error != "invalid or expired refresh token" {
		t.Fatalf("expected invalid refresh token error, got %q", refreshAfterLogoutBody.Error)
	}

	// Step 8: Login again with valid credentials.
	authenticateRes := doRequest(t, client, http.MethodPost, host+"/authenticate", map[string]string{
		"email":    email,
		"password": password,
	}, "", "")
	assertStatus(t, authenticateRes, http.StatusOK)
	authenticated := decodeJSONBody[accessTokenResponse](t, authenticateRes)
	authRefreshToken := cookieValue(t, authenticateRes, "refresh_token")

	if authenticated.AccessToken == "" {
		t.Fatal("expected access token on authenticate")
	}
	if len(authRefreshToken) != 26 {
		t.Fatalf("expected refresh token length 26 after login, got %d", len(authRefreshToken))
	}

	// Step 9: Verify the newly issued access token works on the secured endpoint.
	securedAfterLoginRes := doRequest(t, client, http.MethodGet, host+"/secured", nil, authenticated.AccessToken, "")
	assertStatus(t, securedAfterLoginRes, http.StatusOK)
	securedAfterLogin := decodeJSONBody[securedResponse](t, securedAfterLoginRes)

	if securedAfterLogin.UserID != registered.User.ID {
		t.Fatalf("expected secured userId %d after login, got %d", registered.User.ID, securedAfterLogin.UserID)
	}
}

func TestAcceptanceIdentity_ActivationIsRequiredAndTokenIsSingleUse(t *testing.T) {
	host := startIdentityServer(t)
	client := &http.Client{}
	username, email, password := uniqueCredentials()

	// Step 1: Register a user and keep the activation token.
	registerRes := doRequest(t, client, http.MethodPost, host+"/register", map[string]string{
		"username": username,
		"email":    email,
		"password": password,
	}, "", "")
	assertStatus(t, registerRes, http.StatusAccepted)
	registered := decodeJSONBody[registerResponse](t, registerRes)

	// Step 2: Attempt login before activation and verify account is blocked.
	loginBeforeActivationRes := doRequest(t, client, http.MethodPost, host+"/authenticate", map[string]string{
		"email":    email,
		"password": password,
	}, "", "")
	assertStatus(t, loginBeforeActivationRes, http.StatusForbidden)
	inactiveAccount := decodeJSONBody[errorStringResponse](t, loginBeforeActivationRes)

	if inactiveAccount.Error != "your user account must be activated to access this resource" {
		t.Fatalf("unexpected pre-activation login error: %q", inactiveAccount.Error)
	}

	// Step 3: Activate once and verify success.
	activateRes := doRequest(t, client, http.MethodPut, host+"/activate", map[string]string{
		"activationToken": registered.ActivationToken,
	}, "", "")
	assertStatus(t, activateRes, http.StatusOK)
	activated := decodeJSONBody[accessTokenResponse](t, activateRes)

	if activated.AccessToken == "" {
		t.Fatal("expected access token on successful activation")
	}

	// Step 4: Reuse the same activation token and verify it is rejected.
	reuseActivationRes := doRequest(t, client, http.MethodPut, host+"/activate", map[string]string{
		"activationToken": registered.ActivationToken,
	}, "", "")
	assertStatus(t, reuseActivationRes, http.StatusUnprocessableEntity)
	reuseActivationBody := decodeJSONBody[errorMapResponse](t, reuseActivationRes)

	if reuseActivationBody.Error["token"] != "invalid or expired activation token" {
		t.Fatalf("expected token reuse error, got %v", reuseActivationBody.Error)
	}
}

func TestAcceptanceIdentity_OwnershipAndAdminAuthorization(t *testing.T) {
	host := startIdentityServer(t)
	client := &http.Client{}

	// Step 1: Create and activate user A.
	userAID, userAAccessToken := registerAndActivateUser(t, client, host)

	// Step 2: Create and activate user B.
	userBID, _ := registerAndActivateUser(t, client, host)

	// Step 3: User A can access their own user details endpoint.
	myInfoRes := doRequest(t, client, http.MethodGet, host+"/users/"+strconv.FormatInt(userAID, 10), nil, userAAccessToken, "")
	assertStatus(t, myInfoRes, http.StatusOK)
	myInfo := decodeJSONBody[userDetailsResponse](t, myInfoRes)

	if myInfo.UserDetails.ID != userAID {
		t.Fatalf("expected user details id %d, got %d", userAID, myInfo.UserDetails.ID)
	}
	if !slices.Contains(myInfo.UserDetails.Roles, "user") {
		t.Fatalf("expected user details roles to include user, got %v", myInfo.UserDetails.Roles)
	}

	// Step 4: User A cannot access user B details.
	otherUserRes := doRequest(t, client, http.MethodGet, host+"/users/"+strconv.FormatInt(userBID, 10), nil, userAAccessToken, "")
	assertStatus(t, otherUserRes, http.StatusUnauthorized)
	otherUserBody := decodeJSONBody[errorStringResponse](t, otherUserRes)

	if otherUserBody.Error != "invalid authentication credentials" {
		t.Fatalf("unexpected cross-user access error: %q", otherUserBody.Error)
	}

	// Step 5: User A is not admin, so admin portal must reject access.
	adminPortalRes := doRequest(t, client, http.MethodGet, host+"/admin-portal", nil, userAAccessToken, "")
	assertStatus(t, adminPortalRes, http.StatusForbidden)
	adminPortalBody := decodeJSONBody[errorStringResponse](t, adminPortalRes)

	if adminPortalBody.Error != "your user account does not have the necessary roles/permissions to access this resource" {
		t.Fatalf("unexpected admin authorization error: %q", adminPortalBody.Error)
	}
}

func registerAndActivateUser(t *testing.T, client *http.Client, host string) (userID int64, accessToken string) {
	t.Helper()

	username, email, password := uniqueCredentials()

	registerRes := doRequest(t, client, http.MethodPost, host+"/register", map[string]string{
		"username": username,
		"email":    email,
		"password": password,
	}, "", "")
	assertStatus(t, registerRes, http.StatusAccepted)
	registered := decodeJSONBody[registerResponse](t, registerRes)

	activateRes := doRequest(t, client, http.MethodPut, host+"/activate", map[string]string{
		"activationToken": registered.ActivationToken,
	}, "", "")
	assertStatus(t, activateRes, http.StatusOK)
	activated := decodeJSONBody[accessTokenResponse](t, activateRes)

	if activated.AccessToken == "" {
		t.Fatal("expected access token on activation")
	}

	return registered.User.ID, activated.AccessToken
}

func startIdentityServer(t *testing.T) string {
	t.Helper()

	port := freePort(t)

	cleanup, err := LaunchTestProgram(port)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(cleanup)

	return "http://localhost:" + port + "/v1/identity"
}

func freePort(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("could not acquire free tcp port: %v", err)
	}
	defer listener.Close()

	_, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatalf("could not parse free tcp port: %v", err)
	}

	return port
}

func uniqueCredentials() (username string, email string, password string) {
	sequence := atomic.AddUint64(&credentialCounter, 1)
	suffix := strconv.FormatInt(time.Now().UnixNano(), 36) + "_" + strconv.FormatUint(sequence, 36)
	username = "user_" + suffix
	email = "user_" + suffix + "@example.com"
	password = "supersecure-password"
	return
}

func doRequest(t *testing.T, client *http.Client, method string, url string, payload any, bearerToken string, refreshToken string) *http.Response {
	t.Helper()

	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("could not encode request payload: %v", err)
		}
		body = bytes.NewReader(encoded)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("could not create %s request: %v", method, err)
	}

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}

	if refreshToken != "" {
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken})
	}

	res, err := client.Do(req)
	if err != nil {
		t.Fatalf("could not perform %s %s request: %v", method, url, err)
	}

	return res
}

func assertStatus(t *testing.T, res *http.Response, wantStatus int) {
	t.Helper()

	if res.StatusCode == wantStatus {
		return
	}

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	t.Fatalf("expected status %d, got %d. body=%s", wantStatus, res.StatusCode, strings.TrimSpace(string(body)))
}

func cookieValue(t *testing.T, res *http.Response, cookieName string) string {
	t.Helper()

	for _, cookie := range res.Cookies() {
		if cookie.Name == cookieName {
			return cookie.Value
		}
	}

	t.Fatalf("expected %s cookie in response", cookieName)
	return ""
}

func decodeJSONBody[T any](t *testing.T, res *http.Response) T {
	t.Helper()
	defer res.Body.Close()

	var out T
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatalf("could not decode response body into %T: %v", out, err)
	}

	return out
}
