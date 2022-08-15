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
	var storePerm []Permission

	for _, req := range reqs.Requests {
		perms = append(perms, &pb.PermissionItem{
			Action:   req.Action,
			Resource: req.Resource,
			Filter:   req.Filter,
		})
		storePerm = append(storePerm, Permission{
			Action:   req.Action,
			Resource: req.Resource,
			Filter:   req.Filter,
		})
	}

	err := NewPermissions(s.Config, reqs.Requests[0].UserID).UpdatePermissions(storePerm...)
	if err != nil {
		return nil, err
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

	allowed, err := NewPermissions(s.Config, req.UserID).CheckPerm(Permission{
		Action:   req.Action,
		Resource: req.Resource,
		Filter:   req.Filter,
	})
	if err != nil {
		return nil, err
	}

	return &pb.AllowedResponse{
		Allowed: allowed,
	}, nil
}
