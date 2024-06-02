package api

import (
	"context"
	"errors"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"
)

type Authenticator interface {
	VerifyToken(ctx context.Context, token string, verifyEmail bool) (*auth.Token, error)
}

type FirebaseAuthenticator struct {
	app *firebase.App
}

func NewFirebaseAuthenticator() Authenticator {
	opt := option.WithCredentialsFile("credentials.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatal().Err(err).Msg("error initializing firebase app")
	}
	return &FirebaseAuthenticator{app: app}
}

func (auth *FirebaseAuthenticator) VerifyToken(ctx context.Context, idToken string, verifyEmail bool) (*auth.Token, error) {
	client, err := auth.app.Auth(ctx)
	if err != nil {
		return nil, err
	}
	token, err := client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, err
	}
	if verifyEmail {
		user, e := client.GetUser(ctx, token.UID)
		if e != nil {
			return nil, e
		}
		if !user.EmailVerified {
			return nil, errors.New("email address is not verified")
		}
	}
	return token, err
}
