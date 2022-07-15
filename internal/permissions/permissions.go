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
	Identifier string
	Action     string
	Filter     string
}

func NewPermissions(cfg *config.Config, userID string) *Permissions {
	return &Permissions{
		cfg:    cfg,
		UserID: userID,
	}
}

func (p Permissions) CreateUser(companyID string) ([]Permission, error) {
	userPerms := p.AddPerms(p.UserID, "user", "password", "email", "name", "avatar", "view")
	userPerms = append(userPerms, p.AddPerm(companyID, "topic", "create")...)
	userPerms = append(userPerms, p.AddPerm(companyID, "leader", "list")...)
	userPerms = append(userPerms, p.AddPerm(companyID, "board", "list")...)

	if companyID != "" {
		userPerms = append(userPerms, p.AddPerm(companyID, "company", "view")...)
	}

	return userPerms, nil
}

func (p Permissions) CreateLeader(companyID string) ([]Permission, error) {
	userPerms, err := p.CreateUser(companyID)
	if err != nil {
		return nil, err
	}

	leaderPerms := p.AddPerms("", "timer", "start", "stop", "extend")
	leaderPerms = append(leaderPerms, p.AddPerms(companyID, "retro", "start", "stop")...)
	leaderPerms = append(leaderPerms, p.AddPerms(companyID, "leader", "add", "remove")...)

	return append(userPerms, leaderPerms...), nil
}

func (p Permissions) CreateOwner() ([]Permission, error) {
	userPerms := p.AddPerms(p.UserID, "user", "password", "email", "name", "avatar", "view")

	ownerPerms := p.AddPerm("", "company", "create")

	if err := NewMongo(p.cfg).Create(DataSet{
		UserID:      p.UserID,
		Permissions: ownerPerms,
	}); err != nil {
		return nil, err
	}

	return append(userPerms, ownerPerms...), nil
}

func (p Permissions) UpdateOwnerWithCorrectPerms(companyID string) ([]Permission, error) {
	leaderPerms, err := p.CreateLeader(companyID)
	if err != nil {
		return nil, err
	}

	ownerPerms := p.updatePerms(companyID, "board", "create")
	ownerPerms = append(ownerPerms, p.updatePerms(companyID, "topic", "create")...)
	ownerPerms = append(ownerPerms, p.updatePerms(companyID, "leader", "add", "remove")...)
	ownerPerms = append(ownerPerms, p.updatePerms(companyID, "company", "edit", "delete")...)
	ownerPerms = append(ownerPerms, p.updatePerms(companyID, "billing", "create", "list", "edit", "delete")...)

	return append(leaderPerms, ownerPerms...), nil
}

func (p Permissions) AddPerm(filter, identifier, action string) []Permission {
	return []Permission{
		{
			Identifier: identifier,
			Action:     action,
			Filter:     filter,
		},
	}
}

func (p Permissions) AddPerms(filter, identifier string, action ...string) []Permission {
	perms := make([]Permission, len(action))
	for i, a := range action {
		perms[i] = Permission{
			Identifier: identifier,
			Action:     a,
			Filter:     filter,
		}
	}

	return perms
}

func (p Permissions) updatePerms(filter, identifier string, action ...string) []Permission {
	perms := make([]Permission, len(action))
	for i, a := range action {
		perms[i] = Permission{
			Identifier: identifier,
			Action:     a,
			Filter:     filter,
		}
	}

	return perms
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

func (p Permissions) CheckPerm(userID, filter, identifier, action string) (bool, error) {
	perms, err := NewMongo(p.cfg).Get(userID)
	if err != nil {
		return false, err
	}

	for _, perm := range perms.Permissions {
		if filter == "" {
			if perm.Identifier == identifier && perm.Action == action {
				return true, nil
			}
		} else {
			if perm.Identifier == identifier && perm.Action == action && perm.Filter == filter {
				return true, nil
			}
		}
	}

	return false, nil
}
