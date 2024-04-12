package user

import (
	"context"

	"git.sr.ht/~emersion/gqlclient"
)

type userResponse struct {
	Me User
}

func QueryUser(client *gqlclient.Client) (User, error) {
	resp := userResponse{User{}}

	op := queryUserOperation()
	if err := client.Execute(context.Background(), op, &resp); err != nil {
		return resp.Me, err
	}

	return resp.Me, nil
}

func queryUserOperation() *gqlclient.Operation {
	op := gqlclient.NewOperation(`
	query Me {
		me {
			fullName
			memberships {
				role
				organization {
					id
					name
					slug
				}
			}
		}
	}
	`)

	return op
}
