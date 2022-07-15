package permissions

import (
	"context"
	"fmt"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
	"github.com/retro-board/permission-service/internal/config"
	pb "github.com/retro-board/protos/generated/permissions/v1"
)

type Server struct {
	pb.UnimplementedPermissionsServiceServer
	Config *config.Config
}

func (s *Server) Permission(ctx context.Context, req *pb.PermissionRequest) (*pb.PermissionResponse, error) {
	if req.UserID == "" {
		return nil, bugLog.Errorf("missing user-id")
	}
	p := NewPermissions(s.Config, req.UserID)

	if req.APIKey == "" {
		return nil, bugLog.Errorf("missing service-key")
	} else {
		valid, err := p.ValidateServiceKey(req.APIKey)
		if err != nil {
			return nil, err
		}
		if !valid {
			return nil, bugLog.Errorf("invalid service key")
		}
	}

	return nil, nil
}

func (s *Server) MultiPermissions(ctx context.Context, reqs *pb.MultiRequest) (*pb.PermissionResponse, error) {
	if reqs.Requests[0].UserID == "" {
		return nil, bugLog.Errorf("missing user-id")
	}

	if reqs.Requests[0].APIKey == "" {
		return nil, bugLog.Errorf("missing service-key")
	} else {
		valid, err := NewPermissions(s.Config, reqs.Requests[0].UserID).ValidateServiceKey(reqs.Requests[0].APIKey)
		if err != nil {
			return nil, err
		}
		if !valid {
			return nil, bugLog.Errorf("invalid service key")
		}
	}

	var perms []*pb.PermissionItem

	for _, req := range reqs.Requests {
		perms = append(perms, &pb.PermissionItem{
			Action:   req.Action,
			Resource: req.Resource,
			Filter:   req.Filter,
		})

		fmt.Printf("action: %s, resource: %s, filter: %s\n", req.Action, req.Resource, req.Filter)
	}

	return &pb.PermissionResponse{
		Permissions: perms,
	}, nil
}

func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	if req.UserID == "" {
		return nil, bugLog.Errorf("missing user-id")
	}
	p := NewPermissions(s.Config, req.UserID)

	if req.APIKey == "" {
		return nil, bugLog.Errorf("missing service-key")
	} else {
		valid, err := p.ValidateServiceKey(req.APIKey)
		if err != nil {
			return nil, err
		}
		if !valid {
			return nil, bugLog.Errorf("invalid service key")
		}
	}

	return nil, nil
}

func (s *Server) CanDo(ctx context.Context, req *pb.AllowedRequest) (*pb.AllowedResponse, error) {
	if req.UserID == "" {
		return nil, bugLog.Errorf("missing user-id")
	}

	if req.APIKey == "" {
		return nil, bugLog.Errorf("missing service-key")
	} else {
		valid, err := NewPermissions(s.Config, req.UserID).ValidateServiceKey(req.APIKey)
		if err != nil {
			return nil, err
		}
		if !valid {
			return nil, bugLog.Errorf("invalid service key")
		}
	}

	return nil, nil
}

func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.PermissionResponse, error) {
	if req.UserID == "" {
		bugLog.Info("missing user-id")
		return &pb.PermissionResponse{
			Status: "missing user-id",
		}, nil
	}

	p := NewPermissions(s.Config, req.UserID)
	var permItems []*pb.PermissionItem

	// assume company owner
	if req.CompanyID == nil {
		perms, err := p.CreateOwner()
		if err != nil {
			return &pb.PermissionResponse{
				Status: "internal storage error",
			}, nil
		}

		for _, perm := range perms {
			permItem := &pb.PermissionItem{
				Action:   perm.Action,
				Resource: perm.Identifier,
				Filter:   perm.Filter,
			}
			permItems = append(permItems, permItem)
		}
		var permItems []*pb.PermissionItem
		return &pb.PermissionResponse{
			Permissions: permItems,
			Status:      "created owner",
		}, nil
	}

	return nil, nil
}
