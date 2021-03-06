package permissions

import (
	"context"

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

//nolint:gocyclo
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

	dataset, err := NewMongo(s.Config).Get(req.UserID)
	if err != nil {
		bugLog.Info(err)
		s := "internal storage error"

		return &pb.AllowedResponse{
			Status: &s,
		}, nil
	}

	for _, perm := range dataset.Permissions {
		if perm.Identifier == req.Resource && perm.Action == req.Action {
			if perm.Filter == "" {
				return &pb.AllowedResponse{
					Allowed: true,
				}, nil
			} else if perm.Filter == req.Filter {
				return &pb.AllowedResponse{
					Allowed: true,
				}, nil
			}
		}
	}

	return &pb.AllowedResponse{
		Allowed: false,
	}, nil
}

func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.PermissionResponse, error) {
	if req.UserID == "" {
		bugLog.Info("missing user-id")
		s := "missing user-id"
		return &pb.PermissionResponse{
			Status: &s,
		}, nil
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

	p := NewPermissions(s.Config, req.UserID)
	var permItems []*pb.PermissionItem

	// assume company owner
	if req.CompanyID == nil {
		perms, err := p.CreateOwner()
		if err != nil {
			bugLog.Info(err)

			s := "internal storage error"
			return &pb.PermissionResponse{
				Status: &s,
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
		return &pb.PermissionResponse{
			Permissions: permItems,
		}, nil
	}

	stat := "unknown error"
	return &pb.PermissionResponse{
		Status: &stat,
	}, nil
}
