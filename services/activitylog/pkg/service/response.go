package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	gateway "github.com/cs3org/go-cs3apis/cs3/gateway/v1beta1"
	group "github.com/cs3org/go-cs3apis/cs3/identity/group/v1beta1"
	user "github.com/cs3org/go-cs3apis/cs3/identity/user/v1beta1"
	rpc "github.com/cs3org/go-cs3apis/cs3/rpc/v1beta1"
	provider "github.com/cs3org/go-cs3apis/cs3/storage/provider/v1beta1"
	"github.com/cs3org/reva/v2/pkg/storagespace"
	"github.com/cs3org/reva/v2/pkg/utils"
	libregraph "github.com/owncloud/libre-graph-api-go"

	"github.com/owncloud/ocis/v2/ocis-pkg/l10n"
)

// Translations
var (
	MessageResourceCreated = l10n.Template("{user} added {resource} to {folder}")
	MessageResourceUpdated = l10n.Template("{user} updated {resource} in {folder}")
	MessageResourceTrashed = l10n.Template("{user} deleted {resource} from {folder}")
	MessageResourceMoved   = l10n.Template("{user} moved {resource} to {folder}")
	MessageResourceRenamed = l10n.Template("{user} renamed {oldResource} to {resource}")
	MessageShareCreated    = l10n.Template("{user} shared {resource} with {sharee}")
	MessageShareUpdated    = l10n.Template("{user} updated {field} for the {resource}")
	MessageShareDeleted    = l10n.Template("{user} removed {sharee} from {resource}")
	MessageLinkCreated     = l10n.Template("{user} shared {resource} via link")
	MessageLinkUpdated     = l10n.Template("{user} updated {field} for a link {token} on {resource}")
	MessageLinkDeleted     = l10n.Template("{user} removed link to {resource}")
	MessageSpaceShared     = l10n.Template("{user} added {sharee} as member of {space}")
	MessageSpaceUnshared   = l10n.Template("{user} removed {sharee} from {space}")

	StrSomeField      = l10n.Template(_someField)
	StrPermission     = l10n.Template(_permission)
	StrPassword       = l10n.Template(_password)
	StrExpirationDate = l10n.Template(_expirationDate)
	StrDisplayName    = l10n.Template(_displayName)
	StrDescription    = l10n.Template(_description)
)

const (
	_someField      = "some field"
	_permission     = "permission"
	_password       = "password"
	_expirationDate = "expiration date"
	_displayName    = "display name"
	_description    = "description"
)

// GetActivitiesResponse is the response on GET activities requests
type GetActivitiesResponse struct {
	Activities []libregraph.Activity `json:"value"`
}

// Resource represents an item such as a file or folder
type Resource struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Actor represents a user
type Actor struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

// ActivityOption allows setting variables for an activity
type ActivityOption func(context.Context, gateway.GatewayAPIClient, map[string]interface{}) error

// WithResource sets the resource variable for an activity
func WithResource(ref *provider.Reference, addSpace bool) ActivityOption {
	return func(ctx context.Context, gwc gateway.GatewayAPIClient, vars map[string]interface{}) error {
		info, err := utils.GetResource(ctx, ref, gwc)
		if err != nil {
			vars["resource"] = Resource{
				Name: filepath.Base(ref.GetPath()),
			}
			return err
		}

		vars["resource"] = Resource{
			ID:   storagespace.FormatResourceID(info.GetId()),
			Name: info.GetName(),
		}

		parent, err := utils.GetResourceByID(ctx, info.GetParentId(), gwc)
		if err != nil {
			return err
		}

		vars["folder"] = Resource{
			ID:   info.GetParentId().GetOpaqueId(),
			Name: parent.GetName(),
		}

		if addSpace {
			vars["space"] = Resource{
				ID:   info.GetSpace().GetId().GetOpaqueId(),
				Name: info.GetSpace().GetName(),
			}
		}

		return nil
	}
}

// WithOldResource sets the oldResource variable for an activity
func WithOldResource(ref *provider.Reference) ActivityOption {
	return func(_ context.Context, _ gateway.GatewayAPIClient, vars map[string]interface{}) error {
		name := filepath.Base(ref.GetPath())
		vars["oldResource"] = Resource{
			Name: name,
		}
		return nil
	}
}

// WithTrashedResource sets the resource variable if the resource is trashed
func WithTrashedResource(ref *provider.Reference, rid *provider.ResourceId) ActivityOption {
	return func(ctx context.Context, gwc gateway.GatewayAPIClient, vars map[string]interface{}) error {
		vars["resource"] = Resource{
			Name: filepath.Base(ref.GetPath()),
		}

		resp, err := gwc.ListRecycle(ctx, &provider.ListRecycleRequest{
			Ref: ref,
			Key: rid.GetOpaqueId(),
		})
		if err != nil {
			return err
		}
		if resp.GetStatus().GetCode() != rpc.Code_CODE_OK {
			return fmt.Errorf("error listing recycle: %s", resp.GetStatus().GetMessage())
		}

		for _, item := range resp.GetRecycleItems() {
			if item.GetKey() == rid.GetOpaqueId() {

				vars["resource"] = Resource{
					ID:   storagespace.FormatResourceID(rid),
					Name: filepath.Base(item.GetRef().GetPath()),
				}

				return nil
			}
		}

		return nil
	}
}

// WithUser sets the user variable for an Activity
func WithUser(uid *user.UserId, username string) ActivityOption {
	return func(_ context.Context, gwc gateway.GatewayAPIClient, vars map[string]interface{}) error {
		if username == "" {
			u, err := utils.GetUser(uid, gwc)
			if err != nil {
				vars["user"] = Actor{
					ID:          uid.GetOpaqueId(),
					DisplayName: "DeletedUser",
				}
				return err
			}
			username = u.GetUsername()
		}

		vars["user"] = Actor{
			ID:          uid.GetOpaqueId(),
			DisplayName: username,
		}

		return nil
	}
}

// WithSharee sets the sharee variable for an activity
func WithSharee(uid *user.UserId, gid *group.GroupId) ActivityOption {
	return func(ctx context.Context, gwc gateway.GatewayAPIClient, vars map[string]interface{}) error {
		switch {
		case uid != nil:
			u, err := utils.GetUser(uid, gwc)
			if err != nil {
				vars["sharee"] = Actor{
					DisplayName: "DeletedUser",
				}
				return err
			}

			vars["sharee"] = Actor{
				ID:          uid.GetOpaqueId(),
				DisplayName: u.GetUsername(),
			}
		case gid != nil:
			vars["sharee"] = Actor{
				ID:          gid.GetOpaqueId(),
				DisplayName: "DeletedGroup",
			}
			r, err := gwc.GetGroup(ctx, &group.GetGroupRequest{GroupId: gid})
			if err != nil {
				return fmt.Errorf("error getting group: %w", err)
			}

			if r.GetStatus().GetCode() != rpc.Code_CODE_OK {
				return fmt.Errorf("error getting group: %s", r.GetStatus().GetMessage())
			}

			vars["sharee"] = Actor{
				ID:          gid.GetOpaqueId(),
				DisplayName: r.GetGroup().GetDisplayName(),
			}

		}

		return nil
	}
}

// WithSpace sets the space variable for an activity
func WithSpace(spaceid *provider.StorageSpaceId) ActivityOption {
	return func(ctx context.Context, gwc gateway.GatewayAPIClient, vars map[string]interface{}) error {
		if _, ok := vars["space"]; ok {
			// do not override space if already set
			return nil
		}

		s, err := utils.GetSpace(ctx, spaceid.GetOpaqueId(), gwc)
		if err != nil {
			vars["space"] = Resource{
				ID:   spaceid.GetOpaqueId(),
				Name: "DeletedSpace",
			}
			return err
		}
		vars["space"] = Resource{
			ID:   s.GetId().GetOpaqueId(),
			Name: s.GetName(),
		}

		return nil
	}
}

// WithTranslation sets a variable that translation is needed for
func WithTranslation(t *l10n.Translator, locale string, key string, values []string) ActivityOption {
	return func(_ context.Context, _ gateway.GatewayAPIClient, vars map[string]interface{}) error {
		f := t.Translate(StrSomeField, locale)
		if len(values) > 0 {
			for i := range values {
				values[i] = t.Translate(mapField(values[i]), locale)
			}
			f = strings.Join(values, ", ")
		}
		vars[key] = Resource{
			Name: f,
		}
		return nil
	}
}

// WithVar sets a variable for an activity
func WithVar(key, id, name string) ActivityOption {
	return func(_ context.Context, _ gateway.GatewayAPIClient, vars map[string]interface{}) error {
		vars[key] = Resource{
			ID:   id,
			Name: name,
		}
		return nil
	}
}

// NewActivity creates a new activity
func NewActivity(message string, ts time.Time, eventID string, vars map[string]interface{}) libregraph.Activity {
	return libregraph.Activity{
		Id:    eventID,
		Times: libregraph.ActivityTimes{RecordedTime: ts},
		Template: libregraph.ActivityTemplate{
			Message:   message,
			Variables: vars,
		},
	}
}

// GetVars calls other service to gather the required data for the activity variables
func (s *ActivitylogService) GetVars(ctx context.Context, opts ...ActivityOption) (map[string]interface{}, error) {
	gwc, err := s.gws.Next()
	if err != nil {
		return nil, err
	}

	vars := make(map[string]interface{})
	for _, opt := range opts {
		if err := opt(ctx, gwc, vars); err != nil {
			s.log.Info().Err(err).Msg("error getting activity vars")
		}
	}

	return vars, nil
}

func mapField(val string) string {
	switch val {
	case "TYPE_PERMISSIONS", "permission":
		return _permission
	case "TYPE_PASSWORD", "password":
		return _password
	case "TYPE_EXPIRATION", "expiration":
		return _expirationDate
	case "TYPE_DISPLAYNAME":
		return _displayName
	case "TYPE_DESCRIPTION":
		return _description
	}
	return _someField
}
