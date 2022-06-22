package permissions

import (
	"github.com/retro-board/permission-service/internal/config"
)

type Permissions struct {
	cfg *config.Config

	Perms []Permission
}

type Permission struct {
	Identifier string
	Action     string
}

func NewPermissions(cfg *config.Config) *Permissions {
	return &Permissions{
		cfg: cfg,
	}
}

func CreateUser(userID, companyID string) []Permission {
	return []Permission{
		{
			Identifier: userID,
			Action:     "password",
		},
		{
			Identifier: userID,
			Action:     "email",
		},
		{
			Identifier: userID,
			Action:     "name",
		},
		{
			Identifier: userID,
			Action:     "avatar",
		},
		{
			Identifier: "retro",
			Action:     "list",
		},
		{
			Identifier: "topic",
			Action:     "list",
		},
		{
			Identifier: "topic",
			Action:     "create",
		},
		{
			Identifier: "topic",
			Action:     "delete",
		},
		{
			Identifier: "topic",
			Action:     "edit",
		},
		{
			Identifier: "board",
			Action:     "list",
		},
		{
			Identifier: "leader",
			Action:     "list",
		},
		{
			Identifier: "vote",
			Action:     "add",
		},
		{
			Identifier: "vote",
			Action:     "remove",
		},
		{
			Identifier: userID,
			Action:     "view",
		},
	}
}

func CreateLeader(userID, companyID string) []Permission {
	userPerms := CreateUser(userID, companyID)
	leaderPerms := []Permission{
		{
			Identifier: "board",
			Action:     "edit",
		},
		{
			Identifier: "timer",
			Action:     "start",
		},
		{
			Identifier: "timer",
			Action:     "stop",
		},
		{
			Identifier: "timer",
			Action:     "extend",
		},
		{
			Identifier: "retro",
			Action:     "start",
		},
		{
			Identifier: "retro",
			Action:     "end",
		},
		{
			Identifier: "action",
			Action:     "create",
		},
		{
			Identifier: "action",
			Action:     "actioned",
		},
		{
			Identifier: "action",
			Action:     "delete",
		},
		{
			Identifier: "action",
			Action:     "edit",
		},
		{
			Identifier: "leader",
			Action:     "add",
		},
		{
			Identifier: "leader",
			Action:     "remove",
		},
	}

	return append(userPerms, leaderPerms...)
}

func CreateOwner(userID, companyID string) []Permission {
	leaderPerms := CreateLeader(userID, companyID)
	ownerPerms := []Permission{
		{
			Identifier: "board",
			Action:     "delete",
		},
		{
			Identifier: "board",
			Action:     "create",
		},
		{
			Identifier: "company",
			Action:     "create",
		},
		{
			Identifier: companyID,
			Action:     "delete",
		},
		{
			Identifier: companyID,
			Action:     "edit",
		},
		{
			Identifier: "billing",
			Action:     "create",
		},
		{
			Identifier: "billing",
			Action:     "delete",
		},
		{
			Identifier: "billing",
			Action:     "edit",
		},
	}

	return append(leaderPerms, ownerPerms...)
}

func AddPerm(identifier, action string) []Permission {
	return []Permission{
		{
			Identifier: identifier,
			Action:     action,
		},
	}
}
