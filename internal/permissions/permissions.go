package permissions

import (
	"context"

	bugLog "github.com/bugfixes/go-bugfixes/logs"
	"github.com/retro-board/permission-service/internal/config"
	keyBuf "github.com/retro-board/protos/generated/key/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Permissions struct {
	cfg    *config.Config
	UserID string

	Perms []Permission
}

type Permission struct {
	Resource string
	Action   string
	Filter   string
}

func NewPermissions(cfg *config.Config, userID string) *Permissions {
	return &Permissions{
		cfg:    cfg,
		UserID: userID,
	}
}

func (p Permissions) AddPermission(perms ...Permission) error {
	currentPerms, err := NewMongo(p.cfg).Get(p.UserID)
	if err != nil {
		return err
	}
	currentPerms.Permissions = append(currentPerms.Permissions, perms...)

	return NewMongo(p.cfg).Update(*currentPerms)
}

func (p Permissions) UpdatePermissions(perms ...Permission) error {
	newPerms := []Permission{}
	newPerms = append(newPerms, perms...)
	return NewMongo(p.cfg).Update(DataSet{
		UserID:      p.UserID,
		Permissions: newPerms,
	})
}

func (p Permissions) RemovePermission(perm Permission) error {
	currentPerms, err := NewMongo(p.cfg).Get(p.UserID)
	if err != nil {
		return err
	}

	for i, cperm := range currentPerms.Permissions {
		if cperm.Resource == perm.Resource && cperm.Action == perm.Action && cperm.Filter == perm.Filter {
			currentPerms.Permissions = append(currentPerms.Permissions[:i], currentPerms.Permissions[i+1:]...)
			break
		}
	}

	return NewMongo(p.cfg).Update(*currentPerms)
}

func (p Permissions) CheckPerm(perm Permission) (bool, error) {
	perms, err := NewMongo(p.cfg).Get(p.UserID)
	if err != nil {
		return false, err
	}

	for _, cperm := range perms.Permissions {
		if cperm.Filter == "*" || cperm.Filter == perm.Filter {
			if cperm.Resource == "*" || cperm.Resource == perm.Resource {
				if cperm.Action == "*" || cperm.Action == perm.Action {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func (p *Permissions) ValidateServiceKey(serviceKey string) (bool, error) {
	// skip check if development mode
	if p.cfg.Development {
		return true, nil
	}

	conn, err := grpc.Dial(p.cfg.Local.Services.KeyService.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return false, err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			bugLog.Infof("failed to close connection to key service: %v", err)
		}
	}()

	c := keyBuf.NewKeyServiceClient(conn)
	resp, err := c.Validate(context.Background(), &keyBuf.ValidateRequest{
		UserId:     p.UserID,
		ServiceKey: p.cfg.OnePasswordKey,
		CheckKey:   serviceKey,
	})
	if err != nil {
		return false, err
	}
	if resp.Valid {
		return true, nil
	}

	return false, nil
}
