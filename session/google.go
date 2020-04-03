package session

import "golang.org/x/net/context"

type GoogleApi struct {}

func (g *GoogleApi) GetIdToken(ctx context.Context, code string, userEntity *UserEntity) error {
	return nil
}
