package auth

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

const QueryTokenKey = "token"

func Register(r *gin.Engine, h *Handler) {
	authGroup := r.Group("/auth")
	authGroup.GET("/login", Login(h))
	authGroup.POST("/login", Login(h))
	authGroup.GET("/callback", Callback(h))
	authGroup.POST("/callback", Callback(h))
}

func Login(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var state State
		// Parse form data, check if everything was provided and also check if
		// callback is a valid URL
		err := c.Bind(&state)
		if err != nil {
			c.String(http.StatusBadRequest, "failed to bind json object", err)
			return
		}
		// Redirect to authCodeURL if no error occured
		authCodeURL, err := h.GetAuthCodeURL(&state)
		if err != nil {
			c.String(http.StatusInternalServerError, "failed to acquire auth code url")
		}
		c.Redirect(http.StatusSeeOther, authCodeURL)
	}
}

func Callback(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := c.Request.ParseForm()
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("failed to parse request: %v", err))
			return
		}
		ctx := c.Request.Context()

		// Authorization redirect callback from OAuth2 auth flow.
		if errMsg := c.Request.Form.Get("error"); errMsg != "" {
			c.String(http.StatusBadRequest, errMsg+": "+c.Request.Form.Get("error_description"))
			return
		}
		code := c.Request.Form.Get("code")
		if code == "" {
			c.String(http.StatusBadRequest, fmt.Sprintf("no code in request: %q", c.Request.Form))
			return
		}

		encoded := c.Request.Form.Get("state")
		if encoded == "" {
			c.String(http.StatusBadRequest, fmt.Sprintf("no state in request: %q", c.Request.Form))
			return
		}

		token, err := h.Exchange(ctx, code)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("failed to get token: %v", err))
			return
		}

		state, _, err := h.VerifyStateAndClaims(ctx, token, encoded)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("failed to verify token: %v", err), err)
			return
		}

		callbackURL, err := url.Parse(state.Callback)
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("error parsing url from state: %v", err))
			return
		}
		err = addTokenToQuery(callbackURL, token)
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("failed to add token to query: %v", err))
			return
		}

		c.Redirect(http.StatusSeeOther, callbackURL.String())
	}
}

func addTokenToQuery(u *url.URL, token *oauth2.Token) error {
	q, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return err
	}
	encoded, err := encode(token)
	if err != nil {
		return err
	}
	q.Add(QueryTokenKey, encoded)
	u.RawQuery = q.Encode()
	return nil
}
