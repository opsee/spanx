package service

import (
	"github.com/opsee/basic/com"
	"github.com/opsee/basic/tp"
	"github.com/opsee/spanx/roler"
	"golang.org/x/net/context"
	"net/http"
	"time"
)

const (
	userKey = iota
	requestKey
)

func (s *service) StartHTTP(addr string) {
	router := tp.NewHTTPRouter(context.Background())

	// json api
	router.Handle("PUT", "/credentials", decoders(com.User{}, ResolveCredentialsRequest{}), s.resolveCredentials())
	router.Handle("GET", "/credentials", []tp.DecodeFunc{tp.AuthorizationDecodeFunc(userKey, com.User{})}, s.getCredentials())

	// set a big timeout bc aws be slow
	router.Timeout(5 * time.Minute)

	http.ListenAndServe(addr, router)
}

func decoders(userType interface{}, requestType interface{}) []tp.DecodeFunc {
	return []tp.DecodeFunc{
		tp.AuthorizationDecodeFunc(userKey, userType),
		tp.RequestDecodeFunc(requestKey, requestType),
	}
}

func (s *service) resolveCredentials() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		request, ok := ctx.Value(requestKey).(*ResolveCredentialsRequest)
		if !ok {
			return ctx, http.StatusBadRequest, errUnknown
		}

		user, ok := ctx.Value(userKey).(*com.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errUnknown
		}

		creds, err := s.ResolveCredentials(user, request)
		if err != nil {
			if err == roler.InsufficientPermissions {
				return ctx, http.StatusUnauthorized, err
			}

			return nil, http.StatusInternalServerError, err
		}

		return creds, http.StatusOK, nil
	}
}

func (s *service) getCredentials() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		user, ok := ctx.Value(userKey).(*com.User)
		if !ok {
			return ctx, http.StatusUnauthorized, errUnknown
		}

		creds, err := s.GetCredentials(user)
		if err != nil {
			if err == roler.AccountNotFound {
				return ctx, http.StatusNotFound, err
			}

			return nil, http.StatusInternalServerError, err
		}

		return creds, http.StatusOK, nil
	}
}
